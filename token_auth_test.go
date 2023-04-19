package grpcx

import (
	"context"
	"net"
	"strings"
	"testing"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/stretchr/testify/require"
	"github.com/whitekid/goxp/log"
	"github.com/whitekid/grpcx/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type serviceImpl struct {
	proto.UnimplementedSampleServiceServer
}

func (s *serviceImpl) Echo(ctx context.Context, req *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
	return wrapperspb.String(req.Value), nil
}

func (s *serviceImpl) Echox(ctx context.Context, req *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
	return wrapperspb.String(req.Value), nil
}

func (s *serviceImpl) UnrayInterceptor() []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{s.authInterceptor()}
}

var authMap = map[string]string{
	"echo":  "",
	"echox": "required",
}

func (s *serviceImpl) authInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		p := strings.Split(info.FullMethod, "/")
		method := p[len(p)-1]

		auth, ok := authMap[method]
		if !ok {
			auth = "required"
		}

		switch auth {
		case "": // no authentication
		case "required": // user required
			_, err := grpc_auth.AuthFromMD(ctx, "bearer")
			if err != nil {
				return nil, err
			}

		default:
			return nil, status.Errorf(codes.Internal, "invalid auth method: %v", auth)
		}

		return handler(ctx, req)
	}
}

func (s *serviceImpl) StreamInterceptor() []grpc.StreamServerInterceptor { return nil }

func TestTokenAuth(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	service := &serviceImpl{}

	logger := log.Zap(log.New(zap.AddCallerSkip(2)))

	unaryInterceptors := []grpc.UnaryServerInterceptor{
		grpc_zap.UnaryServerInterceptor(logger),
	}
	unaryInterceptors = append(unaryInterceptors, service.UnrayInterceptor()...)

	streamInterceptors := []grpc.StreamServerInterceptor{
		grpc_zap.StreamServerInterceptor(logger),
	}
	streamInterceptors = append(streamInterceptors, service.StreamInterceptor()...)

	g := grpc.NewServer(
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	)

	proto.RegisterSampleServiceServer(g, service)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	go g.Serve(ln)

	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	tokenAuth := NewTokenAuth("")
	conn, err := grpc.DialContext(ctx, ln.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(tokenAuth),
	)
	require.NoError(t, err)

	client := proto.NewSampleServiceClient(conn)

	type args struct {
		token string
		fn    func(ctx context.Context, in *wrapperspb.StringValue, opts ...grpc.CallOption) (*wrapperspb.StringValue, error)
	}
	tests := [...]struct {
		name    string
		args    args
		wantErr bool
	}{
		{`token not required`, args{"", client.Echo}, false},
		{`token required`, args{"", client.Echox}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := "hello workd"
			tokenAuth.SetToken(tt.args.token)

			got, err := tt.args.fn(ctx, wrapperspb.String(value))
			require.Truef(t, (err != nil) == tt.wantErr, `echo() failed: error = %+v, wantErr = %v`, err, tt.wantErr)
			if tt.wantErr {
				return
			}
			require.Equal(t, value, got.Value)
		})
	}
}
