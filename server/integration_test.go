package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"flag"
	"net"
	"os"
	"server/auth"
	// "server/pb"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const PORT = "8080"

var addr = flag.String("addr", "localhost:"+PORT, "The address of the grpc server")

func getRequiredCertificates(t *testing.T, certificateService auth.ICertificateService) (tls.Certificate, *x509.CertPool) {
	clientCAPool, err := certificateService.GetClientCACertificatePool()
	if err != nil {
		t.Fatalf("failed to get client CA certificate pool: %v", err)
	}

	return certificateService.GetServerCertificate(), clientCAPool
}

func issueDeviceCreationRequest(t *testing.T, adminService *AdminService) string {
	// create a CSR for the device
	token, err := adminService.CreateAddDeviceRequest()
	if err != nil {
		t.Fatalf("failed to create add device request: %v", err)
	}
	return token
}

func createDevice(t *testing.T, token string, adminService *AdminService) (string, tls.Certificate) {

	// create a new self signed cert
	// issue a certificate from the client CA
	_, devicePrivateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate device key: %v", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		t.Fatalf("failed to get hostname: %v", err)
	}

	// create a CSR for the device
	csrTemplate := &x509.CertificateRequest{
		Subject:     pkix.Name{CommonName: "device-1"},
		DNSNames:    []string{hostname},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
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

	deviceInfo := auth.NewDeviceInfo(csr)

	// add the device
	cert, err := adminService.AddDevice(token, deviceInfo)
	if err != nil {
		t.Fatalf("failed to add device: %v", err)
	}

	clientCert := tls.Certificate{
		Certificate: [][]byte{cert.Raw},
		PrivateKey:  devicePrivateKey,
		Leaf:        cert,
	}

	return hostname, clientCert
}

func createGrpcClient(t *testing.T, certificateService auth.ICertificateService, token string, adminService *AdminService) *grpc.ClientConn {
	// get the required certificates for the client
	_, clientCAPool := getRequiredCertificates(t, certificateService)

	// add the device
	hostname, cert := createDevice(t, token, adminService)
	tlsConfig := &tls.Config{
		ServerName:   hostname,
		Certificates: []tls.Certificate{cert},
		RootCAs:      clientCAPool,
	}

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	if err != nil {
		t.Fatalf("failed to create gRPC client: %v", err)
	}
	return conn
}

func TestIntegration(t *testing.T) {
	basePath := t.TempDir()

	// create the services
	_, certificateService, _, grpcServer, adminService := CreateServices(basePath)

	// start the admin service
	if err := adminService.Start(); err != nil {
		t.Fatalf("failed to start admin service: %v", err)
	}

	// start the gRPC server
	if err := grpcServer.Start(PORT); err != nil {
		t.Fatalf("failed to start grpc server: %v", err)
	}

	// issue a token for adding a device
	token := issueDeviceCreationRequest(t, adminService)

	conn := createGrpcClient(t, certificateService, token, adminService)
	defer conn.Close()
	
	// fss := pb.NewFileSystemClient(conn)
}
