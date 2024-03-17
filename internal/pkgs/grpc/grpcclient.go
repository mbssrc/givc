package grpc

import (
	"context"
	"crypto/tls"
	"time"

	"givc/internal/pkgs/types"

	"google.golang.org/grpc"
	grpc_codes "google.golang.org/grpc/codes"
	grpc_creds "google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	grpc_metadata "google.golang.org/grpc/metadata"
	grpc_status "google.golang.org/grpc/status"

	log "github.com/sirupsen/logrus"
)

// TODO: parametrize retry timeout according to situation, will block functionality
//
//	needs to be more lenient for startup
var (
	RETRY_TIME     = 1 * time.Second
	RETRY_ATTEMPTS = 20
)

type GrpcEndpointConfig struct {
	Name      string
	Address   string
	Port      string
	Protocol  string
	TlsConfig *tls.Config
}

func NewDial(cfg *types.EndpointConfig) (*grpc.ClientConn, error) {

	// @TODO Input validation

	options := []grpc.DialOption{}

	// Setup TLS credentials
	var tlsCredentials grpc.DialOption
	if cfg.Transport.TlsConfig != nil {
		tlsCredentials = grpc.WithTransportCredentials(grpc_creds.NewTLS(cfg.Transport.TlsConfig))
	} else {
		tlsCredentials = grpc.WithTransportCredentials(insecure.NewCredentials())
	}
	options = append(options, tlsCredentials)

	// Setup GRPC config
	interceptors := []grpc.UnaryClientInterceptor{
		withOutgoingContext,
		withRequestRetries,
	}
	options = append(options, grpc.WithChainUnaryInterceptor(interceptors...))

	return grpc.Dial(cfg.Transport.Address+":"+cfg.Transport.Port, options...)
}

func withOutgoingContext(ctx context.Context, method string, req, resp interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

	var outmd grpc_metadata.MD
	if md, ok := grpc_metadata.FromOutgoingContext(ctx); ok {
		outmd = md.Copy()
	}

	ctx = grpc_metadata.NewOutgoingContext(context.Background(), outmd)

	return invoker(ctx, method, req, resp, cc, opts...)
}

func withRequestRetries(ctx context.Context, method string, req, resp interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

	var err error
	for attempt := 0; attempt < RETRY_ATTEMPTS; attempt++ {
		err = invoker(ctx, method, req, resp, cc, opts...)

		if grpc_status.Code(err) == grpc_codes.Unavailable {
			log.Debugf("Cannot reach %s, retrying...", cc.Target())
			time.Sleep(RETRY_TIME)
			continue
		}

		break
	}

	return err
}
