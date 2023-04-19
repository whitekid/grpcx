package grpcx

import (
	"context"

	"google.golang.org/grpc/credentials"
)

// Token client token authorizer...
type TokenAuth struct {
	token string
}

var _ credentials.PerRPCCredentials = (*TokenAuth)(nil)

func NewTokenAuth(token string) *TokenAuth {
	return &TokenAuth{token: token}
}

func (t *TokenAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	metadata := map[string]string{}
	if t.token != "" {
		metadata["Authorization"] = "Bearer " + t.token
	}

	return metadata, nil
}

func (t *TokenAuth) SetToken(token string)          { t.token = token }
func (t *TokenAuth) RequireTransportSecurity() bool { return false }
