package grpcx

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/whitekid/grpcx/proto"
)

func TestError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	serve(ctx, ln)

	conn, err := grpc.DialContext(ctx, ln.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	client := proto.NewSampleServiceClient(conn)

	type args struct {
		code uint32
	}
	tests := [...]struct {
		name     string
		args     args
		wantCode codes.Code
	}{
		{`valid`, args{ERR_INVALID_ARG}, codes.InvalidArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.RaiseError(ctx, wrapperspb.UInt32(uint32(tt.args.code)))
			require.Error(t, err)
			serr, ok := status.FromError(err)
			require.True(t, ok)

			require.Equalf(t, tt.wantCode, serr.Code(), "want %v but got %v", tt.wantCode.String(), serr.Code().String())
		})
	}
}
