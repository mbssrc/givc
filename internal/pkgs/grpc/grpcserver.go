package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"givc/internal/pkgs/types"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"

	"golang.org/x/sync/errgroup"
	grpc "google.golang.org/grpc"
	grpc_creds "google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	log "github.com/sirupsen/logrus"
)

type GrpcServiceRegistration interface {
	Name() string
	RegisterGrpcService(*grpc.Server)
}

type GrpcServerConfig struct {
	Name      string
	Address   string
	Port      string
	Protocol  string
	TlsConfig *tls.Config
	Services  []GrpcServiceRegistration
}

type GrpcServer struct {
	config *GrpcServerConfig
	// sockpath   string
	grpcServer *grpc.Server
}

func NewGrpcServerConfig(cfg *types.EndpointConfig) *GrpcServerConfig {

	serverConfig := GrpcServerConfig{
		Name:      cfg.Name,
		Address:   cfg.Transport.Address,
		Port:      cfg.Transport.Port,
		Protocol:  cfg.Transport.Protocol,
		TlsConfig: cfg.Transport.TlsConfig,
	}

	return &serverConfig
}

func NewServer(cfg *GrpcServerConfig) (*GrpcServer, error) {
	// // Unix Socket path
	// var sockpath string

	// if len(conf.BindSocket) > 0 {
	// 	sockpath = conf.BindSocket
	// } else {
	// 	sockpath = filepath.Join("/run", fmt.Sprintf("%s_%d.sock", filepath.Base(os.Args[0]), os.Getpid()))
	// }

	// if runtime.GOOS == "linux" {
	// 	sockpath = "@" + sockpath
	// }

	// // Composite Server
	// srv := Server{
	// 	conf:     conf,
	// 	sockpath: sockpath,
	// }

	// GRPC Server
	srv := GrpcServer{
		config: cfg,
	}

	// TLS gRPC creds option
	var grpcTlsConfig grpc.ServerOption
	if srv.config.TlsConfig != nil {
		grpcTlsConfig = grpc.Creds(grpc_creds.NewTLS(srv.config.TlsConfig))
	} else {
		grpcTlsConfig = grpc.Creds(insecure.NewCredentials())
	}

	// GRPC Server
	srv.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.TagBasedRequestFieldExtractor("log"))),
				grpc.UnaryServerInterceptor(unaryLogRequestInterceptor),
				grpc_logrus.UnaryServerInterceptor(log.NewEntry(log.StandardLogger())),
			),
		),
		grpcTlsConfig,
	)

	// Register gRPC services
	for _, s := range srv.config.Services {
		log.Info("Registering service: ", s.Name())
		s.RegisterGrpcService(srv.grpcServer)
	}

	return &srv, nil
}

func (s *GrpcServer) ListenAndServe(ctx context.Context) error {
	// grpcLL, err := s.config.Listeners()
	// if err != nil {
	// 	return err
	// }

	// Default GRPC on Unix Socket
	// if l, err := net.Listen("unix", s.sockpath); err == nil {
	// 	defer l.Close()

	// 	grpcLL = append(grpcLL, l)
	// } else {
	// 	return err
	// }
	listener, err := net.Listen(s.config.Protocol, s.config.Address+":"+s.config.Port)
	if err != nil {
		return err
	}
	defer listener.Close()

	group, ctx := errgroup.WithContext(ctx)
	idleConnsClosed := make(chan struct{})
	go func() {
		<-ctx.Done()

		s.grpcServer.GracefulStop()

		close(idleConnsClosed)
	}()

	// for _, l := range grpcLL {
	// 	listener := l
	// 	group.Go(func() error {
	// 		log.WithFields(log.Fields{"addr": listener.Addr().String()}).Info("Starting GRPC server")

	// 		if err := s.grpcServer.Serve(listener); err != nil {
	// 			return err
	// 		}

	// 		log.WithFields(log.Fields{"addr": listener.Addr().String()}).Info("GRPC server stopped")

	// 		return nil
	// 	})
	// }

	group.Go(func() error {
		log.WithFields(log.Fields{"addr": listener.Addr().String()}).Info("Starting GRPC server")

		if err := s.grpcServer.Serve(listener); err != nil {
			return err
		}

		log.WithFields(log.Fields{"addr": listener.Addr().String()}).Info("GRPC server stopped")

		return nil
	})

	<-idleConnsClosed

	if err := group.Wait(); err != nil {
		return fmt.Errorf("GRPC Server error: %s", err)
	}

	return nil
}

func unaryLogRequestInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.WithFields(grpc_ctxtags.Extract(ctx).Values()).Info("GRPC Request: ", info.FullMethod)
	return handler(ctx, req)
}
