package main

import (
	"context"
	"crypto/tls"
	"flag"
	"givc/api/admin"
	givc_grpc "givc/internal/pkgs/grpc"
	"givc/internal/pkgs/types"
	givc_util "givc/internal/pkgs/utility"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {

	execName := filepath.Base(os.Args[0])
	log.Infof("Executing %s \n", execName)

	flag.Usage = func() {
		log.Infof("Usage of %s:\n", execName)
		log.Infof("%s [OPTIONS] \n", execName)
		flag.PrintDefaults()
	}

	name := flag.String("name", "<application>", "Name of application to start")
	address := flag.String("ip", "127.0.0.1", "Host ip")
	port := flag.String("port", "9000", "Host port")
	protocol := flag.String("protocol", "tcp", "Transport protocol")
	notls := flag.Bool("notls", false, "Disable TLS")
	flag.Parse()

	var tlsConfig *tls.Config
	if !*notls {

		// @TODO add path and file checks

		cacert := os.Getenv("CA_CERT")
		if cacert == "" {
			log.Fatalf("No 'CA_CERT' environment variable present. To turn off TLS use '-notls'.")
		}

		cert := os.Getenv("HOST_CERT")
		if cert == "" {
			log.Fatalf("No 'HOST_CERT' environment variable present. To turn off TLS use '-notls'.")
		}

		key := os.Getenv("HOST_KEY")
		if key == "" {
			log.Fatalf("No 'HOST_KEY' environment variable present. To turn off TLS use '-notls'.")
		}

		tlsConfig = givc_util.TlsClientConfig(cacert, cert, key)
	}

	cfgAdminServer := &types.EndpointConfig{
		Name: "Admin Server",
		Transport: types.TransportConfig{
			Address:   *address,
			Port:      *port,
			Protocol:  *protocol,
			TlsConfig: tlsConfig,
		},
	}

	// Setup and dial GRPC client
	var conn *grpc.ClientConn
	conn, err := givc_grpc.NewDial(cfgAdminServer)
	if err != nil {
		log.Fatalf("Cannot create grpc client: %v", err)
	}
	defer conn.Close()

	// Create client
	client := admin.NewAdminServiceClient(conn)
	if client == nil {
		log.Fatalf("Failed to create 'NewAdminServiceClient'")
	}

	ctx := context.Background()
	switch *name {
	case "poweroff":
		_, err := client.Poweroff(ctx, &admin.Empty{})
		if err != nil {
			log.Errorf("Error executing poweroff: %s", err)
		}
	case "reboot":
		_, err := client.Reboot(ctx, &admin.Empty{})
		if err != nil {
			log.Errorf("Error executing reboot: %s", err)
		}
	default:
		req := &admin.ApplicationRequest{
			AppName: *name,
		}
		resp, err := client.StartApplication(ctx, req)
		if err != nil {
			log.Errorf("Error executing application: %s", err)
		}
		log.Infoln(resp)
	}

}
