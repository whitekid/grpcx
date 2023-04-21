package grpcx

import (
	"context"
	"errors"
	"strings"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/whitekid/grpcx/proto"
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
	return []grpc.UnaryServerInterceptor{s.authInterceptor(), s.errorInterceptor()}
}

var authMap = map[string]string{
	"Echo":       "",
	"Echox":      "required",
	"RaiseError": "",
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

func (s *serviceImpl) errorInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		result, err := handler(ctx, req)

		switch {
		case errors.Is(err, errInvalidArgs):
			err = status.Error(codes.InvalidArgument, "")
		}

		return result, err
	}
}

func (s *serviceImpl) StreamInterceptor() []grpc.StreamServerInterceptor { return nil }

const (
	_               uint32 = iota
	ERR_INVALID_ARG uint32 = iota + 20
)

var (
	errInvalidArgs = errors.New("invalid args")
)

func (s *serviceImpl) RaiseError(ctx context.Context, req *wrapperspb.UInt32Value) (*emptypb.Empty, error) {
	switch req.Value {
	case ERR_INVALID_ARG:
		return nil, errInvalidArgs
	}

	return &emptypb.Empty{}, nil
}
