package grpcx

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/whitekid/grpcx/proto"
)

func TestTokenAuth(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	service := &serviceImpl{}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	serve(ctx, ln,
		grpc.ChainUnaryInterceptor(service.UnrayInterceptor()...),
		grpc.ChainStreamInterceptor(service.StreamInterceptor()...),
	)

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
			require.Truef(t, (err != nil) == tt.wantErr, `%v failed: error = %+v, wantErr = %v`, tt.args.fn, err, tt.wantErr)
			if tt.wantErr {
				return
			}
			require.Equal(t, value, got.Value)
		})
	}
}
