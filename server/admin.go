package main

import (
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

func (s *AdminService) Start() error {
	return nil
}