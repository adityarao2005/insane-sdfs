package main

import (
	"crypto/x509"
	"server/auth"
)

type AdminService struct {
	tokenService *auth.ITokenService
	certificateService *auth.ICertificateService
}

func NewAdminService(tokenService *auth.ITokenService, certificateService *auth.ICertificateService) (*AdminService, error) {
	return &AdminService{
		tokenService: tokenService,
		certificateService: certificateService,
	}, nil
}

func (s *AdminService) CreateAddDeviceRequest() (string, error) {
	return (*s.tokenService).IssueToken()
}

func (s *AdminService) AddDevice(token string, csr *auth.DeviceInfo) (*x509.Certificate, error) {
	err := (*s.tokenService).AcceptToken(token)
	if err != nil {
		return nil, err
	}

	cert, err := (*s.certificateService).IssueCertificate(csr)
	if err != nil {
		return nil, err
	}
	
	return cert, nil
}

func (s *AdminService) Start() error {
	return nil
}