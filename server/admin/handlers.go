package admin

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"mime"
	"net/http"
	"path/filepath"
	"server/auth"
)

type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expiresAt"`
}

type RegisterDeviceRequest struct {
	Token string `json:"token"`
	CSR   string `json:"csr"`
}

type CertificateResponse struct {
	Certificate string `json:"certificate"`
	ExpiresAt   string `json:"expiresAt"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (s *AdminService) handleCreateDeviceToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token, err := s.CreateAddDeviceRequest()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Token expires in 10 minutes from now
	expiresAt := "2026-05-03T12:10:00Z" // Placeholder; should calculate based on token service config
	respondWithJSON(w, http.StatusOK, TokenResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	})
}

func (s *AdminService) handleRegisterDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterDeviceRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Token == "" || req.CSR == "" {
		respondWithError(w, http.StatusBadRequest, "token and csr are required")
		return
	}

	// Parse CSR from PEM format
	block, _ := pem.Decode([]byte(req.CSR))
	if block == nil {
		respondWithError(w, http.StatusBadRequest, "Invalid CSR format, expected PEM")
		return
	}

	csrBytes := block.Bytes
	csr, err := x509.ParseCertificateRequest(csrBytes)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to parse CSR: "+err.Error())
		return
	}

	// Convert x509.CertificateRequest to auth.DeviceInfo
	deviceInfo := &auth.DeviceInfo{
		CSR: csr,
	}

	// Register device and get certificate
	cert, err := s.AddDevice(req.Token, deviceInfo)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to register device: "+err.Error())
		return
	}

	// Encode certificate to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})

	respondWithJSON(w, http.StatusOK, CertificateResponse{
		Certificate: string(certPEM),
		ExpiresAt:   cert.NotAfter.String(),
	})
}

func (s *AdminService) handleServeUI(w http.ResponseWriter, r *http.Request) {
	// If this looks like a file request, serve static content directly.
	if filepath.Ext(r.URL.Path) != "" {
		s.handleServeStatic(w, r)
		return
	}

	// Serve index.html for all routes (SPA fallback)
	data, err := s.uiFS.ReadFile("ui/dist/index.html")
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (s *AdminService) handleServeStatic(w http.ResponseWriter, r *http.Request) {
	// Serve static files from ui/dist with cache headers
	path := "ui/dist" + r.URL.Path

	data, err := s.uiFS.ReadFile(path)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Determine content type based on file extension
	ext := filepath.Ext(r.URL.Path)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}

func respondWithError(w http.ResponseWriter, status int, message string) {
	respondWithJSON(w, status, ErrorResponse{Error: message})
}
