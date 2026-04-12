package filesystem_service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"server/auth"
	"server/pb"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type GrpcFileSystemServer struct {
	server   *grpc.Server
	listener net.Listener
}

func NewGrpcFileSystemServer(certificateService *auth.ICertificateService) *GrpcFileSystemServer {
	// create the server
	server := GrpcFileSystemServer{}
	var opts []grpc.ServerOption
	
	// if the server options aren't null then we can set up TLS for the server
	if certificateService != nil {
		// get the servers tls config from the certificate service
		tlsConfig, err := (*certificateService).GetServerTlsConfig()
		if err != nil {
			fmt.Printf("Error getting server TLS config: %v\n", err)
			return nil
		}

		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}
	// create the gRPC server with the specified options
	server.server = grpc.NewServer(opts...)

	return &server
}

func (s *GrpcFileSystemServer) Start(port string) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return err
	}
	s.listener = listener
	return s.server.Serve(listener)
}

func (s *GrpcFileSystemServer) Stop() error {
	s.server.GracefulStop()
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// file service wrapper to implement the gRPC server interface

type FileSystemServerImpl struct {
	pb.UnimplementedFileSystemServer
	service *FileSystemService
}

func (s FileSystemServerImpl) UploadFile(req grpc.ClientStreamingServer[pb.UploadFileRequestFragment, pb.SuccessResponse]) error {
	if s.service == nil {
		return fmt.Errorf("file system service not initialized")
	}

	firstFragment, err := req.Recv()
	if err != nil {
		return fmt.Errorf("error receiving first file fragment: %w", err)
	}

	path := firstFragment.GetPath()
	if path == "" {
		return fmt.Errorf("file path is required")
	}

	dataChannel := make(chan bytes.Buffer)

	go func() {
		defer close(dataChannel)

		dataChannel <- *bytes.NewBuffer(firstFragment.GetData())
		fmt.Printf("Received file fragment for path: %s, size: %d bytes\n", path, len(firstFragment.GetData()))

		for {
			fragment, err := req.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Printf("Error receiving file fragment: %v\n", err)
				return
			}

			dataChannel <- *bytes.NewBuffer(fragment.GetData())
			fmt.Printf("Received file fragment for path: %s, size: %d bytes\n", path, len(fragment.GetData()))
		}
	}()

	err = s.service.UploadFile(path, dataChannel)
	if err != nil {
		return fmt.Errorf("error uploading file: %v", err)
	}

	resp := &pb.SuccessResponse{Success: true}
	return req.SendAndClose(resp)
}

func (s FileSystemServerImpl) DownloadFile(req *pb.DownloadFileRequest, stream grpc.ServerStreamingServer[pb.DownloadFileResponseFragment]) error {
	if s.service == nil {
		return fmt.Errorf("file system service not initialized")
	}

	dataChannel, err := s.service.DownloadFile(req.GetPath())
	if err != nil {
		return fmt.Errorf("error downloading file: %v", err)
	}

	for data := range dataChannel {
		if err := stream.Send(&pb.DownloadFileResponseFragment{Data: data.Bytes()}); err != nil {
			return fmt.Errorf("error sending file fragment: %v", err)
		}
	}

	return nil
}

func (s FileSystemServerImpl) DeleteFile(_ context.Context, req *pb.FileInfoRequest) (*pb.SuccessResponse, error) {
	if s.service == nil {
		return nil, fmt.Errorf("file system service not initialized")
	}

	if err := s.service.DeleteFile(req.GetPath()); err != nil {
		return nil, fmt.Errorf("error deleting file: %v", err)
	}

	return &pb.SuccessResponse{Success: true}, nil
}

func (s FileSystemServerImpl) GetFileInfo(_ context.Context, req *pb.FileInfoRequest) (*pb.FileInfo, error) {
	if s.service == nil {
		return nil, fmt.Errorf("file system service not initialized")
	}

	info, err := s.service.GetFileInfo(req.GetPath())
	if err != nil {
		return nil, fmt.Errorf("error getting file info: %v", err)
	}

	return &pb.FileInfo{
		Path:             info.Path,
		Size:             info.Size,
		LastModifiedTime: strconv.FormatInt(info.LastModified, 10),
		IsDirectory:      info.IsDirectory,
	}, nil
}

func (s FileSystemServerImpl) ListFiles(req *pb.FileInfoRequest, stream grpc.ServerStreamingServer[pb.FileInfo]) error {
	if s.service == nil {
		return fmt.Errorf("file system service not initialized")
	}

	files, err := s.service.ListFiles(req.GetPath())
	if err != nil {
		return fmt.Errorf("error listing files: %v", err)
	}

	for _, file := range files {
		if err := stream.Send(&pb.FileInfo{
			Path:             file.Path,
			Size:             file.Size,
			LastModifiedTime: strconv.FormatInt(file.LastModified, 10),
			IsDirectory:      file.IsDirectory,
		}); err != nil {
			return fmt.Errorf("error sending file info: %v", err)
		}
	}

	return nil
}

func (s FileSystemServerImpl) CreateDirectory(_ context.Context, req *pb.FileInfoRequest) (*pb.SuccessResponse, error) {
	if s.service == nil {
		return nil, fmt.Errorf("file system service not initialized")
	}

	if err := s.service.CreateDirectory(req.GetPath()); err != nil {
		return nil, fmt.Errorf("error creating directory: %v", err)
	}

	return &pb.SuccessResponse{Success: true}, nil
}

func (s FileSystemServerImpl) Ping(_ context.Context, req *pb.PingRequest) (*pb.PongResponse, error) {
	if s.service == nil {
		return nil, fmt.Errorf("file system service not initialized")
	}

	return &pb.PongResponse{Message: "Pong: " + req.GetMessage()}, nil
}

// helper function to create a new gRPC file system server implementation with the provided service
func NewGrpcFileSystemServerImpl(service* FileSystemService) FileSystemServerImpl {
	return FileSystemServerImpl{service: service}
}

// adds the filesystem service
func (s *GrpcFileSystemServer) AddService(service* FileSystemService) error {
	pb.RegisterFileSystemServer(s.server, NewGrpcFileSystemServerImpl(service))

	return nil
}
