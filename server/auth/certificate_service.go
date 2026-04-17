package auth

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

type DeviceInfo struct {
	CSR *x509.CertificateRequest
}

func NewDeviceInfo(csr *x509.CertificateRequest) *DeviceInfo {
	return &DeviceInfo{CSR: csr}
}

type ICertificateService interface {
	// IssueCertificate generates a new certificate for authentication
	IssueCertificate(deviceInfo *DeviceInfo) (*x509.Certificate, error)
	// Get Client CA Certificate pool
	GetClientCACertificatePool() (*x509.CertPool, error)
	// Get Server Certificate
	GetServerCertificate() tls.Certificate
	// Get Server TLS Config
	GetServerTlsConfig() (*tls.Config, error)
}

type CertificateService struct {
	serverCert   tls.Certificate
	clientCACert tls.Certificate
}

func (s *CertificateService) IssueCertificate(deviceInfo *DeviceInfo) (*x509.Certificate, error) {
	if deviceInfo == nil || deviceInfo.CSR == nil {
		return nil, fmt.Errorf("device info CSR is required")
	}

	err := deviceInfo.CSR.CheckSignature()
	if err != nil {
		return nil, err
	}

	// create template certificate based on the CSR
	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128)) // Generate a random serial number
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      deviceInfo.CSR.Subject,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0), // valid for 1 year
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		DNSNames:     deviceInfo.CSR.DNSNames,
		IPAddresses:  deviceInfo.CSR.IPAddresses,
	}

	// parse the CA certificate
	var cert tls.Certificate
	if s.clientCACert.Leaf == nil {
		// pars ethe first certificate in the chain if Leaf is not set
		c, err := x509.ParseCertificate(s.clientCACert.Certificate[0])
		if err != nil {
			return nil, err
		}

		cert.Leaf = c
	} else {
		cert.Leaf = s.clientCACert.Leaf
	}

	// sign the certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, template, cert.Leaf, deviceInfo.CSR.PublicKey, s.clientCACert.PrivateKey)
	if err != nil {
		return nil, err
	}

	return x509.ParseCertificate(certBytes)
}

func (s *CertificateService) GetClientCACertificatePool() (*x509.CertPool, error) {
	certPool := x509.NewCertPool()
	cert, err := x509.ParseCertificate(s.clientCACert.Certificate[0])
	if err != nil {
		return nil, err
	}

	certPool.AddCert(cert)
	return certPool, nil
}

func (s CertificateService) GetServerCertificate() tls.Certificate {
	return s.serverCert
}

func (s *CertificateService) GetServerTlsConfig() (*tls.Config, error) {
	certPool, err := s.GetClientCACertificatePool()
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{s.serverCert},
		ClientCAs:    certPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	return tlsConfig, nil
}

const (
	server           = "server"
	clientCA         = "clientCA"
	ServerCertFile   = server + ".pem"
	ServerKeyFile    = server + ".key"
	ClientCACertFile = clientCA + ".pem"
	ClientCAKeyFile  = clientCA + ".key"
)

func certExists(certPath, keyPath string) bool {
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return false
	}

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return false
	}

	return true
}

func generateCertificate(certPath, keyPath, host string, isCA bool) error {

	// if the certificate and key files already exist, do not regenerate them
	if certExists(certPath, keyPath) {
		return nil
	}

	// remove the certs and keys if they exist but are invalid
	os.RemoveAll(certPath)
	os.RemoveAll(keyPath)

	// generate a new ECDSA private key
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	// Generate a random serial number
	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	// create a certificate template
	template := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{CommonName: host},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0), // valid for 1 year
		DNSNames:              []string{host, "localhost"},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		IsCA:                  isCA,
		BasicConstraintsValid: true,
	}

	if isCA {
		template.KeyUsage |= x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	} else {
		template.KeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
	}

	// self-sign the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, publicKey, privateKey)
	if err != nil {
		return err
	}

	// create the certificate and key files
	certOut, err := os.Create(certPath)
	if err != nil {
		return err
	}
	defer certOut.Close()

	keyOut, err := os.Create(keyPath)
	if err != nil {
		return err
	}
	defer keyOut.Close()

	// Serialize private key in PKCS#8 ASN.1 format for tls.LoadX509KeyPair compatibility.
	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return err
	}

	// write the certificate and key to files
	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if err != nil {
		return err
	}

	err = pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyDER})
	if err != nil {
		return err
	}

	return nil
}

func NewCertificateService(certDir string) (ICertificateService, error) {

	// check if the certificate files exist
	os.MkdirAll(certDir, 0755)

	// get absolute paths for the certificate and key files
	serverCertPath, err := filepath.Abs(filepath.Join(certDir, ServerCertFile))
	if err != nil {
		return nil, err
	}
	serverKeyPath, err := filepath.Abs(filepath.Join(certDir, ServerKeyFile))
	if err != nil {
		return nil, err
	}
	clientCACertPath, err := filepath.Abs(filepath.Join(certDir, ClientCACertFile))
	if err != nil {
		return nil, err
	}
	clientCAKeyPath, err := filepath.Abs(filepath.Join(certDir, ClientCAKeyFile))
	if err != nil {
		return nil, err
	}

	// regenerate the server certificate and key
	host, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	// If not, generate new certificates and save them to the specified directory
	err = generateCertificate(serverCertPath, serverKeyPath, host, false)
	if err != nil {
		return nil, err
	}
	// regenerate the client CA certificate and key
	err = generateCertificate(clientCACertPath, clientCAKeyPath, host, true)
	if err != nil {
		return nil, err
	}

	// load the server certificate and client CA certificate
	serverCert, err := tls.LoadX509KeyPair(serverCertPath, serverKeyPath)
	if err != nil {
		return nil, err
	}

	clientCACert, err := tls.LoadX509KeyPair(clientCACertPath, clientCAKeyPath)
	if err != nil {
		return nil, err
	}

	// return the certificate service
	return &CertificateService{
		serverCert:   serverCert,
		clientCACert: clientCACert,
	}, nil
}
