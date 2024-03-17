package utility

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	log "github.com/sirupsen/logrus"
)

var (
	cipherSuites = []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
	}
)

func TlsServerConfig(CACertFilePath string, CertFilePath string, KeyFilePath string, mutual bool) *tls.Config {

	// Load TLS certificates and key
	serverTLSCert, err := tls.LoadX509KeyPair(CertFilePath, KeyFilePath)
	if err != nil {
		log.Fatalf("[TlsServerConfig] Error loading server certificate and key file: %v", err)
	}
	certPool := x509.NewCertPool()
	caCertPEM, err := os.ReadFile(CACertFilePath)
	if err != nil {
		log.Fatalf("[TlsServerConfig] Error loading CA certificate: %v", err)
	}
	ok := certPool.AppendCertsFromPEM(caCertPEM)
	if !ok {
		log.Fatalf("[TlsServerConfig] Invalid CA certificate.")
	}

	var clientAuth tls.ClientAuthType
	if mutual {
		clientAuth = tls.RequireAndVerifyClientCert
	} else {
		clientAuth = tls.NoClientCert
	}

	// Set TLS configuration
	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS13,
		ClientAuth:   clientAuth,
		ClientCAs:    certPool,
		Certificates: []tls.Certificate{serverTLSCert},
		CipherSuites: cipherSuites,
	}

	return tlsConfig
}

func TlsClientConfig(CACertFilePath string, CertFilePath string, KeyFilePath string) *tls.Config {

	// Load TLS certificates and key
	clientTLSCert, err := tls.LoadX509KeyPair(CertFilePath, KeyFilePath)
	if err != nil {
		log.Fatalf("[TlsEndpointConfig] Error loading client certificate and key file: %v", err)
	}
	certPool := x509.NewCertPool()
	caCertPEM, err := os.ReadFile(CACertFilePath)
	if err != nil {
		log.Fatalf("[TlsEndpointConfig] Error loading CA certificate: %v", err)
	}
	ok := certPool.AppendCertsFromPEM(caCertPEM)
	if !ok {
		log.Fatalf("[TlsEndpointConfig] Invalid CA certificate.")
	}

	// Set TLS configuration
	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS13,
		RootCAs:      certPool,
		Certificates: []tls.Certificate{clientTLSCert},
		CipherSuites: cipherSuites,
	}

	return tlsConfig
}
