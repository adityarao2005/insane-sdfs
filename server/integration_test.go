package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"flag"
	"io"
	"net"
	"os"
	"path/filepath"
	"server/admin"
	"server/auth"
	"server/pb"
	"time"

	// "server/pb"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const PORT = "8080"

var addr = flag.String("addr", "localhost:"+PORT, "The address of the grpc server")

func getRequiredCertificates(t *testing.T, certificateService auth.ICertificateService) (*x509.CertPool, *x509.CertPool) {
	clientCAPool, err := certificateService.GetClientCACertificatePool()
	if err != nil {
		t.Fatalf("failed to get client CA certificate pool: %v", err)
	}

	serverCAPool, err := certificateService.GetServerCACertificate()
	if err != nil {
		t.Fatalf("failed to get server CA certificate: %v", err)
	}

	return serverCAPool, clientCAPool
}

func issueDeviceCreationRequest(t *testing.T, adminService *admin.AdminService) string {
	// create a CSR for the device
	token, err := adminService.CreateAddDeviceRequest()
	if err != nil {
		t.Fatalf("failed to create add device request: %v", err)
	}
	return token
}

func createDevice(t *testing.T, token string, adminService *admin.AdminService) (string, tls.Certificate) {

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

func createGrpcClient(t *testing.T, certificateService auth.ICertificateService, token string, adminService *admin.AdminService) *grpc.ClientConn {
	// get the required certificates for the client
	serverCAPool, clientCAPool := getRequiredCertificates(t, certificateService)

	// add the device
	hostname, cert := createDevice(t, token, adminService)
	tlsConfig := &tls.Config{
		ServerName:   hostname,
		Certificates: []tls.Certificate{cert},
		RootCAs:      serverCAPool,
		ClientCAs:    clientCAPool,
	}

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	if err != nil {
		t.Fatalf("failed to create gRPC client: %v", err)
	}
	return conn
}

func integrationTestSetup(t *testing.T) (*grpc.ClientConn, string) {
	basePath := t.TempDir()
	rootPath := filepath.Join(basePath, "root")

	// create the services
	_, certificateService, _, grpcServer, adminService := CreateServices(basePath)

	// start the admin service
	if err := adminService.Start(); err != nil {
		t.Fatalf("failed to start admin service: %v", err)
	}

	go grpcServer.Start(PORT)
	t.Cleanup(func() {
		// Stop the in-process gRPC server when the test exits.
		grpcServer.Stop()
	})

	// issue a token for adding a device
	token := issueDeviceCreationRequest(t, adminService)

	conn := createGrpcClient(t, certificateService, token, adminService)
	t.Cleanup(func() {
		// Ensure client connections are always released.
		conn.Close()
	})

	return conn, rootPath
}

func callUnaryGetFileInfo(t *testing.T, fss pb.FileSystemClient, path string) *pb.FileInfo {
	// Bound each RPC with a timeout so failures don't hang the suite.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := fss.GetFileInfo(ctx, &pb.FileInfoRequest{
		Path: path,
	})
	if err != nil {
		t.Fatalf("failed to get file info: %v", err)
	}
	return resp
}

func callUnaryListFiles(t *testing.T, fss pb.FileSystemClient, path string) []*pb.FileInfo {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := fss.ListFiles(ctx, &pb.FileInfoRequest{
		Path: path,
	})

	if err != nil {
		t.Fatalf("failed to list files: %v", err)
	}

	files := make([]*pb.FileInfo, 0)

	// ListFiles is server-streaming; consume until EOF.
	for {
		info, err := resp.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("failed to receive file info: %v", err)
		}

		files = append(files, info)
	}

	return files
}

func callUnaryCreateDirectory(t *testing.T, fss pb.FileSystemClient, path string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := fss.CreateDirectory(ctx, &pb.FileInfoRequest{
		Path: path,
	})
	if err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
}

func callUnaryDeleteFile(t *testing.T, fss pb.FileSystemClient, path string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := fss.DeleteFile(ctx, &pb.FileInfoRequest{
		Path: path,
	})
	if err != nil {
		t.Fatalf("failed to delete file: %v", err)
	}
}

func callUnaryUploadFile(t *testing.T, fss pb.FileSystemClient, path string, data chan bytes.Buffer) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := fss.UploadFile(ctx)
	if err != nil {
		t.Fatalf("failed to upload file: %v", err)
	}

	// UploadFile is client-streaming; send chunks then close to commit.
	for buf := range data {
		if err := stream.Send(&pb.UploadFileRequestFragment{
			Data: buf.Bytes(),
			Path: path,
		}); err != nil {
			t.Fatalf("failed to send file chunk: %v", err)
		}
	}
	if _, err := stream.CloseAndRecv(); err != nil {
		t.Fatalf("failed to close upload stream: %v", err)
	}
}

func callUnaryDownloadFile(t *testing.T, fss pb.FileSystemClient, path string) bytes.Buffer {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := fss.DownloadFile(ctx, &pb.DownloadFileRequest{
		Path: path,
	})

	if err != nil {
		t.Fatalf("failed to download file: %v", err)
	}

	var data bytes.Buffer

	// DownloadFile is server-streaming; aggregate chunks into one buffer.
	for {
		info, err := resp.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("failed to receive file info: %v", err)
		}

		data.Write(info.Data)
	}

	return data
}

func TestIntegrationUploadAndDownloadFile(t *testing.T) {
	// End-to-end file transfer over mTLS gRPC.
	conn, rootDir := integrationTestSetup(t)
	fss := pb.NewFileSystemClient(conn)

	targetPath := filepath.Join(rootDir, "nested", "payload.txt")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatalf("create target directory: %v", err)
	}

	input := make(chan bytes.Buffer)
	go func() {
		// Send two chunks so streaming code paths are exercised.
		input <- *bytes.NewBufferString("hello ")
		input <- *bytes.NewBufferString("world")
		close(input)
	}()

	callUnaryUploadFile(t, fss, targetPath, input)

	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read uploaded file: %v", err)
	}
	if string(content) != "hello world" {
		t.Fatalf("unexpected uploaded content: %q", string(content))
	}

	downloaded := callUnaryDownloadFile(t, fss, targetPath)

	if downloaded.String() != "hello world" {
		t.Fatalf("unexpected downloaded content: %q", downloaded.String())
	}
}

func TestIntegrationGetFileInfoAndListFiles(t *testing.T) {
	// Validate metadata RPCs against a known file + directory layout.
	conn, rootDir := integrationTestSetup(t)
	fss := pb.NewFileSystemClient(conn)

	filePath := filepath.Join(rootDir, "alpha.txt")
	dirPath := filepath.Join(rootDir, "nested")

	if err := os.WriteFile(filePath, []byte("abc"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		t.Fatalf("create directory: %v", err)
	}

	info := callUnaryGetFileInfo(t, fss, filePath)
	if info.Path != filePath || info.Size != 3 || info.IsDirectory {
		t.Fatalf("unexpected file info: %+v", info)
	}

	dirInfo := callUnaryGetFileInfo(t, fss, dirPath)
	if dirInfo.Path != dirPath || !dirInfo.IsDirectory {
		t.Fatalf("unexpected directory info: %+v", dirInfo)
	}

	listed := callUnaryListFiles(t, fss, rootDir)

	if len(listed) != 2 {
		t.Fatalf("unexpected file count: got %d want %d", len(listed), 2)
	}
	if listed[0].Path != filePath || listed[1].Path != dirPath {
		t.Fatalf("unexpected listing order or paths: %+v", listed)
	}
	if listed[0].Size != 3 || listed[0].IsDirectory {
		t.Fatalf("unexpected first entry: %+v", listed[0])
	}
	if !listed[1].IsDirectory {
		t.Fatalf("unexpected second entry: %+v", listed[1])
	}
}

func TestIntegrationCreateAndDeleteDirectory(t *testing.T) {
	// Validate directory lifecycle operations through gRPC.
	conn, rootDir := integrationTestSetup(t)
	fss := pb.NewFileSystemClient(conn)

	targetDir := filepath.Join(rootDir, "a", "b", "c")

	callUnaryCreateDirectory(t, fss, targetDir)

	if _, err := os.Stat(targetDir); err != nil {
		t.Fatalf("stat created directory: %v", err)
	}

	filePath := filepath.Join(targetDir, "delete-me.txt")
	if err := os.WriteFile(filePath, []byte("remove me"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	callUnaryDeleteFile(t, fss, targetDir)
	if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
		t.Fatalf("expected directory to be deleted, got err=%v", err)
	}
}
