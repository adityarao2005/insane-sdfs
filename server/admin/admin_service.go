package admin

import (
	"crypto/x509"
	"embed"
	"fmt"
	"net/http"
	"os"
	"server/auth"
)

//go:embed ui/dist
var uiAssets embed.FS

type AdminService struct {
	tokenService       auth.ITokenService
	certificateService auth.ICertificateService
	httpServer         *http.Server
	uiFS               embed.FS
}

func NewAdminService(tokenService auth.ITokenService, certificateService auth.ICertificateService) (*AdminService, error) {
	return &AdminService{
		tokenService:       tokenService,
		certificateService: certificateService,
		uiFS:               uiAssets,
	}, nil
}

func (s *AdminService) CreateAddDeviceRequest() (string, error) {
	return s.tokenService.IssueToken(), nil
}

func (s *AdminService) AddDevice(token string, csr *auth.DeviceInfo) (*x509.Certificate, error) {
	err := s.tokenService.AcceptToken(token)
	if err != nil {
		return nil, err
	}

	cert, err := s.certificateService.IssueCertificate(csr)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func (s *AdminService) Start() error {
	// Get HTTP port from environment or use default
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8081"
	}

	// Create HTTP mux and register routes
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/admin/device-token", s.handleCreateDeviceToken)
	mux.HandleFunc("/api/admin/device/register", s.handleRegisterDevice)

	// Static files and SPA fallback
	mux.HandleFunc("/assets/", s.handleServeStatic)
	mux.HandleFunc("/favicon.svg", s.handleServeStatic)
	mux.HandleFunc("/icons.svg", s.handleServeStatic)
	mux.HandleFunc("/", s.handleServeUI)

	// Create HTTP server
	addr := ":" + httpPort
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Start HTTP server in a goroutine
	go func() {
		fmt.Printf("Starting admin console HTTP server on %s\n", addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Admin console HTTP server error: %v\n", err)
		}
	}()

	return nil
}
