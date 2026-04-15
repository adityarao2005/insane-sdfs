package auth

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"math/big"
	"time"
)

type DeviceInfo struct {
	csr *x509.CertificateRequest
}

type ICertificateService interface {
	// IssueCertificate generates a new certificate for authentication
	IssueCertificate(deviceInfo *DeviceInfo) (*x509.Certificate, error)
	// Get Client CA Certificate pool
	GetCertificatePool() (*x509.CertPool, error)
	// Get Server Certificate
	GetServerCertificate() (*tls.Certificate, error)
	// Get Server TLS Config
	GetServerTlsConfig() (*tls.Config, error)
}


type CertificateService struct {
	serverCert tls.Certificate
	clientCACert *tls.Certificate
}

func (s *CertificateService) IssueCertificate(deviceInfo *DeviceInfo) (*x509.Certificate, error) {
	err := deviceInfo.csr.CheckSignature()
	if err != nil {
		return nil, err
	}

	// create template certificate based on the CSR
	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128)) // Generate a random serial number
	template := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               deviceInfo.csr.Subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0,), // valid for 1 year
		KeyUsage:             x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:          []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		DNSNames: deviceInfo.csr.DNSNames,
		IPAddresses: deviceInfo.csr.IPAddresses,
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
	}

	// sign the certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, template, cert.Leaf, deviceInfo.csr.PublicKey, s.clientCACert.PrivateKey)
	if err != nil {
		return nil, err
	}

	return x509.ParseCertificate(certBytes)
}

func (s *CertificateService) GetCertificatePool() (*x509.CertPool, error) {
	certPool := x509.NewCertPool()
	cert, err := x509.ParseCertificate(s.clientCACert.Certificate[0])
	if err != nil {
		return nil, err
	}

	certPool.AddCert(cert)
	return certPool, nil
}

func (s *CertificateService) GetServerCertificate() (*tls.Certificate, error) {
	return &s.serverCert, nil
}

func (s *CertificateService) GetServerTlsConfig() (*tls.Config, error) {
	certPool, err := s.GetCertificatePool()
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{s.serverCert},
		ClientCAs: certPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}

	return tlsConfig, nil
}

func NewCertificateService() ICertificateService {
	// TODO: Implement certificate service initialization logic
	return nil
}
