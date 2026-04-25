package auth

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"net"
	"testing"
)

func TestCertificateService(t *testing.T) {
	certDir := t.TempDir()

	// create the certificate service
	service, err := NewCertificateService(certDir)
	if err != nil {
		t.Fatalf("failed to create certificate service: %v", err)
	}

	// check if the client CA certificate is generated
	certPool, err := service.GetClientCACertificatePool()
	if err != nil {
		t.Fatalf("client CA certificate not loaded")
	}
	if certPool == nil {
		t.Fatalf("client CA certificate pool not created")
	}

	// issue a certificate from the client CA
	_, devicePrivateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate device key: %v", err)
	}

	// create a CSR for the device
	csrTemplate := &x509.CertificateRequest{
		Subject:     pkix.Name{CommonName: "device-1"},
		DNSNames:    []string{"device-1.local"},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}
	csrDER, err := x509.CreateCertificateRequest(rand.Reader, csrTemplate, devicePrivateKey)
	if err != nil {
		t.Fatalf("failed to create csr: %v", err)
	}

	// parse the CSR
	csr, err := x509.ParseCertificateRequest(csrDER)
	if err != nil {
		t.Fatalf("failed to parse csr: %v", err)
	}

	// issue the certificate
	issuedCert, err := service.IssueCertificate(NewDeviceInfo(csr))
	if err != nil {
		t.Fatalf("failed to issue certificate: %v", err)
	}
	if issuedCert == nil {
		t.Fatalf("issued certificate is nil")
	}
	if issuedCert.Subject.CommonName != "device-1" {
		t.Fatalf("unexpected certificate subject: %s", issuedCert.Subject.CommonName)
	}

	// verify the issued certificate against the client CA certificate pool
	if _, err := issuedCert.Verify(x509.VerifyOptions{
		Roots:     certPool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}); err != nil {
		t.Fatalf("failed to verify issued certificate: %v", err)
	}

}
