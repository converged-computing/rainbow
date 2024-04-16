package certs

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc/credentials"
)

// Certificate is a manager for a certificate.
// Is it intended to be used as a client OR server but not both
// at the same time.
type Certificate struct {
	caCertFile string
	certFile   string
	keyFile    string

	certPool   *x509.CertPool
	serverCert *tls.Certificate
	clientCert *tls.Certificate
}

// IsEmpty determines if the certificate is empty
func (c *Certificate) IsEmpty() bool {
	return c.caCertFile == "" && c.certFile == "" && c.keyFile == ""
}

// Validate all fields are defined
func (c *Certificate) Validate() error {
	if c.IsEmpty() {
		return nil
	}
	if c.caCertFile == "" {
		return fmt.Errorf("the caCertFile is missing")
	}
	if c.certFile == "" {
		return fmt.Errorf("the certFile is missing")
	}
	if c.keyFile == "" {
		return fmt.Errorf("the keyFile is missing")
	}
	return nil
}

func (c *Certificate) CreatePool() error {

	// read ca's cert, verify to client's certificate
	caPem, err := os.ReadFile(c.caCertFile)
	if err != nil {
		return err
	}

	// create cert pool and append ca's cert
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caPem) {
		return err
	}
	c.certPool = certPool
	return nil
}

// GetCredentials for the server
func (c *Certificate) GetServerCredentials() credentials.TransportCredentials {

	// configuration of the certificate what we want to
	conf := &tls.Config{
		Certificates: []tls.Certificate{*c.serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    c.certPool,
	}
	return credentials.NewTLS(conf)
}

// GenerateServerCert generates the server certificate
func (c *Certificate) GenerateServerCert() error {
	serverCert, err := tls.LoadX509KeyPair(c.certFile, c.keyFile)
	if err != nil {
		return err
	}
	c.serverCert = &serverCert
	return nil
}

// NewCertificate generates a client agnostic certificate manager
func NewCertificate(caCertFile, certFile, keyFile string) (*Certificate, error) {
	// Cut out early if we aren't using TLS
	cert := Certificate{caCertFile: caCertFile, keyFile: keyFile, certFile: certFile}
	if cert.IsEmpty() {
		return &cert, nil
	}
	err := cert.Validate()
	if err != nil {
		return nil, fmt.Errorf("the certificate did not validate: %s", err)
	}

	// If we get here, create the server certificate and pool
	err = cert.CreatePool()
	return &cert, err
}

func NewClientCertificate(caCertFile, certFile, keyFile string) (*Certificate, error) {
	cert, err := NewCertificate(caCertFile, certFile, keyFile)
	if err != nil || cert.IsEmpty() {
		return cert, err
	}
	err = cert.GenerateClientCert()
	return cert, err
}

func (c *Certificate) GenerateClientCert() error {
	//read client cert
	clientCert, err := tls.LoadX509KeyPair(c.certFile, c.keyFile)
	if err != nil {
		return err
	}
	c.clientCert = &clientCert
	return nil
}

// NewCertificate generates a new certificate holder
func NewServerCertificate(caCertFile, certFile, keyFile string) (*Certificate, error) {
	cert, err := NewCertificate(caCertFile, certFile, keyFile)
	if err != nil || cert.IsEmpty() {
		return cert, err
	}
	err = cert.GenerateServerCert()
	return cert, err
}

// GenerateClientCert generates the client tls certificate
func (c *Certificate) GetClientCredentials() (credentials.TransportCredentials, error) {
	config := &tls.Config{
		Certificates: []tls.Certificate{*c.clientCert},
		RootCAs:      c.certPool,
	}
	return credentials.NewTLS(config), nil
}
