package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"insane-sdfs/internal/enrollment"
	"insane-sdfs/internal/filestore"
)

type clientSessionGetRequest struct {
	DevicePublicKey string `json:"devicePublicKey"`
}

func (s *server) handleClientSessionGet(w http.ResponseWriter, r *http.Request) {
	var req clientSessionGetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.DevicePublicKey) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "devicePublicKey is required"})
		return
	}

	sess, err := s.enrollment.StartSessionByPublicKey(req.DevicePublicKey, time.Now().UTC())
	if err != nil {
		switch {
		case errors.Is(err, enrollment.ErrDevicePublicKeyNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "device not approved or not found"})
		case errors.Is(err, enrollment.ErrDeviceRevoked):
			writeJSON(w, http.StatusConflict, map[string]string{"error": "device is revoked"})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not create session"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, sess)
}

func (s *server) withClientSession(next func(http.ResponseWriter, *http.Request, enrollment.Session)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := strings.TrimSpace(r.Header.Get("X-Session-Id"))
		if sessionID == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing session"})
			return
		}

		sess, err := s.enrollment.GetActiveSession(sessionID, time.Now().UTC())
		if err != nil {
			switch {
			case errors.Is(err, enrollment.ErrSessionNotFound):
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid session"})
			case errors.Is(err, enrollment.ErrSessionExpired):
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "session expired"})
			case errors.Is(err, enrollment.ErrSessionInactive):
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "session inactive"})
			default:
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "session validation failed"})
			}
			return
		}
		next(w, r, sess)
	}
}

func (s *server) handleClientUploadFile(w http.ResponseWriter, r *http.Request) {
	s.withClientSession(func(w http.ResponseWriter, r *http.Request, sess enrollment.Session) {
		path := r.URL.Query().Get("path")
		if s.fileStore == nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "file store is not configured"})
			return
		}

		err := s.fileStore.Save(sess.DeviceID, path, r.Body)
		if err != nil {
			switch {
			case errors.Is(err, filestore.ErrInvalidPath):
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid path"})
			default:
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not save file"})
			}
			return
		}

		writeJSON(w, http.StatusCreated, map[string]any{"path": path, "deviceId": sess.DeviceID})
	})(w, r)
}

func (s *server) handleClientDownloadFile(w http.ResponseWriter, r *http.Request) {
	s.withClientSession(func(w http.ResponseWriter, r *http.Request, sess enrollment.Session) {
		path := r.URL.Query().Get("path")
		if s.fileStore == nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "file store is not configured"})
			return
		}

		b, err := s.fileStore.Load(sess.DeviceID, path)
		if err != nil {
			switch {
			case errors.Is(err, filestore.ErrInvalidPath):
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid path"})
			case errors.Is(err, filestore.ErrFileNotFound):
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "file not found"})
			default:
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not load file"})
			}
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
	})(w, r)
}

func (s *server) handleClientListFiles(w http.ResponseWriter, r *http.Request) {
	s.withClientSession(func(w http.ResponseWriter, r *http.Request, sess enrollment.Session) {
		prefix := r.URL.Query().Get("prefix")
		if s.fileStore == nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "file store is not configured"})
			return
		}

		items, err := s.fileStore.List(sess.DeviceID, prefix)
		if err != nil {
			switch {
			case errors.Is(err, filestore.ErrInvalidPath):
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid prefix"})
			default:
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not list files"})
			}
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	})(w, r)
}

func (s *server) handleClientSessionHeartbeat(w http.ResponseWriter, r *http.Request) {
	s.withClientSession(func(w http.ResponseWriter, r *http.Request, sess enrollment.Session) {
		updated, err := s.enrollment.HeartbeatSession(sess.ID, time.Now().UTC())
		if err != nil {
			switch {
			case errors.Is(err, enrollment.ErrSessionExpired):
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "session expired"})
			case errors.Is(err, enrollment.ErrSessionInactive):
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "session inactive"})
			default:
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not heartbeat session"})
			}
			return
		}

		writeJSON(w, http.StatusOK, updated)
	})(w, r)
}
