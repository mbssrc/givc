package main

import (
	"context"
	"crypto/tls"
	"os"
	"strings"
	"time"

	givc_grpc "givc/internal/pkgs/grpc"
	"givc/internal/pkgs/systemmanager"
	"givc/internal/pkgs/types"
	givc_util "givc/internal/pkgs/utility"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type GrpcServiceRegistration interface {
	Name() string
	RegisterService(*grpc.Server)
}

var (
	cfgServer = types.EndpointConfig{
		Name: "localhost",
		Transport: types.TransportConfig{
			Address:   "127.0.0.1",
			Port:      "9000",
			Protocol:  "tcp",
			TlsConfig: nil,
		},
	}
)

func main() {

	log.Infof("Executing %s \n", os.Args[0])

	name := os.Getenv("NAME")
	if name == "" {
		log.Fatalf("No 'NAME' environment variable present.")
	}
	cfgServer.Name = name

	address := os.Getenv("ADDR")
	if address == "" {
		log.Fatalf("No 'ADDR' environment variable present.")
	}
	cfgServer.Transport.Address = address

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatalf("No 'PORT' environment variable present.")
	}
	cfgServer.Transport.Port = port

	protocol := os.Getenv("PROTO")
	if protocol == "" {
		log.Fatalf("No 'PROTO' environment variable present.")
	}
	cfgServer.Transport.Protocol = protocol

	services := strings.Split(os.Getenv("SERVICES"), " ")
	if len(services) < 1 {
		log.Fatalf("A space-separated list of services (host and system-vms) is required in environment variable $SERVICES.")
	}
	cfgServer.Services = append(cfgServer.Services, services...)
	log.Infof("Initialized services: %v\n", cfgServer.Services)

	withTLS := true
	if os.Getenv("TLS") == "false" {
		withTLS = false
	}

	var tlsConfig *tls.Config
	if withTLS {
		cacert := os.Getenv("CA_CERT")
		if cacert == "" {
			log.Fatalf("No 'CA_CERT' environment variable present. To turn off TLS use 'NOTLS'.")
		}
		cert := os.Getenv("HOST_CERT")
		if cert == "" {
			log.Fatalf("No 'HOST_CERT' environment variable present. To turn off TLS use 'NOTLS'.")
		}
		key := os.Getenv("HOST_KEY")
		if key == "" {
			log.Fatalf("No 'HOST_KEY' environment variable present. To turn off TLS use 'NOTLS'.")
		}
		// @TODO add path and file checks

		tlsConfig = givc_util.TlsServerConfig(cacert, cert, key, true)
		cfgServer.Transport.TlsConfig = tlsConfig
	}

	// Create admin server
	adminServer := systemmanager.NewAdminServer(&cfgServer)

	// Start monitoring
	go func() {
		time.Sleep(4 * time.Second)
		adminServer.AdminService.Monitor()
	}()

	// Start server
	grpcServerConfig := givc_grpc.NewGrpcServerConfig(&cfgServer)
	grpcServerConfig.Services = append(grpcServerConfig.Services, adminServer)

	grpcServer, err := givc_grpc.NewServer(grpcServerConfig)
	if err != nil {
		log.Fatalf("Cannot create grpc server config")
	}

	ctx := context.Background()
	err = grpcServer.ListenAndServe(ctx)
	if err != nil {
		log.Fatalf("Grpc server failed: %s", err)
	}
}
