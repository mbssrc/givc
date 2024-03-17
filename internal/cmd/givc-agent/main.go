package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"givc/api/admin"
	givc_grpc "givc/internal/pkgs/grpc"
	"givc/internal/pkgs/serviceclient"
	"givc/internal/pkgs/servicemanager"
	"givc/internal/pkgs/types"
	givc_util "givc/internal/pkgs/utility"

	log "github.com/sirupsen/logrus"
)

func main() {

	execName := filepath.Base(os.Args[0])
	log.Infof("Executing %s \n", execName)

	name := os.Getenv("NAME")
	if name == "" {
		log.Fatalf("No 'NAME' environment variable present.")
	}

	address := os.Getenv("ADDR")
	if address == "" {
		log.Fatalf("No 'ADDR' environment variable present.")
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatalf("No 'PORT' environment variable present.")
	}

	protocol := os.Getenv("PROTO")
	if protocol == "" {
		log.Fatalf("No 'PROTO' environment variable present.")
	}

	parentName := os.Getenv("PARENT")

	var services []string
	services_string, services_present := os.LookupEnv("SERVICES")
	if services_present {
		services = strings.Split(services_string, " ")
	}

	var applications map[string]string
	jsonApplicationString := os.Getenv("APPLICATIONS")
	if jsonApplicationString != "" {
		applications = make(map[string]string)
		err := json.Unmarshal([]byte(jsonApplicationString), &applications)
		if err != nil {
			log.Fatalf("Error unmarshalling JSON string.")
		}
	}

	adminServerName := os.Getenv("ADMIN_SERVER_NAME")
	if adminServerName == "" {
		log.Fatalf("A name for the admin server is required in environment variable $ADMIN_SERVER_NAME.")
	}

	adminServerAddr := os.Getenv("ADMIN_SERVER_ADDR")
	if adminServerAddr == "" {
		log.Fatalf("An address for the admin server is required in environment variable $ADMIN_SERVER_ADDR.")
	}

	adminServerPort := os.Getenv("ADMIN_SERVER_PORT")
	if adminServerPort == "" {
		log.Fatalf("An port address for the admin server is required in environment variable $ADMIN_SERVER_PORT.")
	}

	adminServerProtocol := os.Getenv("ADMIN_SERVER_PROTO")
	if adminServerProtocol == "" {
		log.Fatalf("An address for the admin server is required in environment variable $ADMIN_SERVER_PROTO.")
	}

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
	}

	cfgAdminServer := &types.EndpointConfig{
		Name: adminServerName,
		Transport: types.TransportConfig{
			Address:   adminServerAddr,
			Port:      adminServerPort,
			Protocol:  adminServerProtocol,
			TlsConfig: tlsConfig,
		},
	}

	// TEMP: IP resolution with hardcoded interface
	var err error
	if address == "dynamic" {
		i := 0
		for i < 100 {
			address, err = givc_util.GetInterfaceIpv4("ethint0")
			if err != nil || address == "" {
				log.Warningf("Cannot resolve IP (%s) waiting...", err)
				time.Sleep(1 * time.Second)
				i += 1
			} else {
				break
			}
		}
	}
	if address == "" {
		log.Fatalf("Cannot resolve IP. Exiting.")
	}
	// TEMP: IP resolution

	// Set agent configurations
	agentServiceName := "givc-" + name + ".service"

	var agentType types.UnitType
	var agentSubType types.UnitType
	switch name {
	case "host":
		agentType = types.UNIT_TYPE_HOST_MGR
		agentSubType = types.UNIT_TYPE_HOST_SVC
	case "net-vm":
		agentType = types.UNIT_TYPE_SYSVM_MGR
		agentSubType = types.UNIT_TYPE_SYSVM_SVC
	case "gui-vm":
		agentType = types.UNIT_TYPE_SYSVM_MGR
		agentSubType = types.UNIT_TYPE_SYSVM_SVC
	default:
		agentType = types.UNIT_TYPE_APPVM_MGR
		agentSubType = types.UNIT_TYPE_APPVM_APP
	}

	cfgAgent := &types.EndpointConfig{
		Name: name,
		Transport: types.TransportConfig{
			Address:   address,
			Port:      port,
			Protocol:  protocol,
			TlsConfig: tlsConfig,
		},
	}
	// Add services
	cfgAgent.Services = append(cfgAgent.Services, agentServiceName)
	if services_present {
		cfgAgent.Services = append(cfgAgent.Services, services...)
	}
	log.Infof("Started with services: %v\n", cfgAgent.Services)

	agentEntryRequest := &admin.RegistryRequest{
		Name:   agentServiceName,
		Type:   uint32(agentType),
		Parent: parentName,
		Transport: &admin.TransportConfig{
			Protocol: cfgAgent.Transport.Protocol,
			Address:  cfgAgent.Transport.Address,
			Port:     cfgAgent.Transport.Port,
			WithTls:  withTLS,
		},
		State: &admin.UnitStatus{
			Name: agentServiceName,
		},
	}

	// Register this instance
	// @TODO: Add sync with server, sd_notify + ctx handling

	go func() {
		time.Sleep(1 * time.Second)

		// Register agent
		serviceclient.RegisterRemoteService(cfgAdminServer, agentEntryRequest)

		// Register services
		for _, service := range services {
			if strings.Contains(service, ".service") {
				serviceEntryRequest := &admin.RegistryRequest{
					Name:   service,
					Parent: agentServiceName,
					Type:   uint32(agentSubType),
					Transport: &admin.TransportConfig{
						Protocol: cfgAgent.Transport.Protocol,
						Address:  cfgAgent.Transport.Address,
						Port:     cfgAgent.Transport.Port,
						WithTls:  withTLS,
					},
					State: &admin.UnitStatus{
						Name: service,
					},
				}
				log.Infof("Trying to register service: %s", service)
				serviceclient.RegisterRemoteService(cfgAdminServer, serviceEntryRequest)
			}
		}
	}()

	grpcServerConfig := givc_grpc.NewGrpcServerConfig(cfgAgent)

	systemdControlServer, err := servicemanager.NewSystemdControlServer(cfgAgent.Services, applications)
	if err != nil {
		log.Fatalf("Cannot create systemd control server")
	}
	grpcServerConfig.Services = append(grpcServerConfig.Services, systemdControlServer)

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
