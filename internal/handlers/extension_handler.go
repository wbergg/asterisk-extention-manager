package handlers

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/wbergg/asterisk-extention-manager/internal/asterisk"
	"github.com/wbergg/asterisk-extention-manager/internal/auth"
	"github.com/wbergg/asterisk-extention-manager/internal/config"
	"github.com/wbergg/asterisk-extention-manager/internal/models"
)

type ExtensionHandler struct {
	DB     *sql.DB
	Config *config.Config
}

type extensionRequest struct {
	Extension   int    `json:"extension"`
	CallerID    string `json:"callerid"`
	SIPPassword string `json:"sip_password"`
}

func (h *ExtensionHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	exts, err := models.ListExtensionsByUser(h.DB, claims.UserID)
	if err != nil {
		http.Error(w, `{"error":"failed to list extensions"}`, http.StatusInternalServerError)
		return
	}
	if exts == nil {
		exts = []models.Extension{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exts)
}

func (h *ExtensionHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())

	var req extensionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if claims.Role != "admin" && (req.Extension < claims.MinExt || req.Extension > claims.MaxExt) {
		http.Error(w, `{"error":"extension outside your permitted range"}`, http.StatusForbidden)
		return
	}

	if claims.Role != "admin" {
		blocked, err := models.IsExtensionBlocked(h.DB, req.Extension)
		if err != nil {
			http.Error(w, `{"error":"failed to check extension"}`, http.StatusInternalServerError)
			return
		}
		if blocked {
			http.Error(w, fmt.Sprintf(`{"error":"extension %d has been blocked by admin"}`, req.Extension), http.StatusForbidden)
			return
		}
	}

	if req.SIPPassword != "" {
		if msg := ValidatePassword(req.SIPPassword, 8); msg != "" {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, msg), http.StatusBadRequest)
			return
		}
	}

	ext := &models.Extension{
		Extension: req.Extension,
		UserID:    claims.UserID,
		CallerID:  req.CallerID,
	}

	if err := models.CreateExtension(h.DB, ext, req.SIPPassword); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			http.Error(w, `{"error":"extension already taken"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error":"failed to create extension"}`, http.StatusInternalServerError)
		return
	}

	models.AuditLog(h.DB, claims.UserID, "create_extension", fmt.Sprintf("ext=%d", ext.Extension))

	if err := asterisk.SyncAndReload(h.DB, h.Config); err != nil {
		fmt.Printf("WARNING: asterisk sync failed: %v\n", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ext)
}

func (h *ExtensionHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	extNum, err := strconv.Atoi(chi.URLParam(r, "ext"))
	if err != nil {
		http.Error(w, `{"error":"invalid extension number"}`, http.StatusBadRequest)
		return
	}

	ext, err := models.GetExtension(h.DB, extNum, claims.UserID)
	if err != nil {
		http.Error(w, `{"error":"extension not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ext)
}

func (h *ExtensionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	extNum, err := strconv.Atoi(chi.URLParam(r, "ext"))
	if err != nil {
		http.Error(w, `{"error":"invalid extension number"}`, http.StatusBadRequest)
		return
	}

	if err := models.DeleteExtension(h.DB, extNum, claims.UserID); err != nil {
		http.Error(w, `{"error":"extension not found"}`, http.StatusNotFound)
		return
	}

	models.AuditLog(h.DB, claims.UserID, "delete_extension", fmt.Sprintf("ext=%d", extNum))

	if err := asterisk.SyncAndReload(h.DB, h.Config); err != nil {
		fmt.Printf("WARNING: asterisk sync failed: %v\n", err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ExtensionHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	extNum, err := strconv.Atoi(chi.URLParam(r, "ext"))
	if err != nil {
		http.Error(w, `{"error":"invalid extension number"}`, http.StatusBadRequest)
		return
	}

	// Verify the extension belongs to this user
	ext, err := models.GetExtension(h.DB, extNum, claims.UserID)
	if err != nil {
		http.Error(w, `{"error":"extension not found"}`, http.StatusNotFound)
		return
	}

	var req extensionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.SIPPassword != "" {
		if msg := ValidatePassword(req.SIPPassword, 8); msg != "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": msg})
			return
		}
	}

	if err := models.UpdateExtension(h.DB, extNum, req.CallerID, req.SIPPassword); err != nil {
		http.Error(w, `{"error":"failed to update extension"}`, http.StatusInternalServerError)
		return
	}

	models.AuditLog(h.DB, claims.UserID, "update_extension", fmt.Sprintf("ext=%d", extNum))

	if err := asterisk.SyncAndReload(h.DB, h.Config); err != nil {
		fmt.Printf("WARNING: asterisk sync failed: %v\n", err)
	}

	ext, _ = models.GetExtension(h.DB, extNum, claims.UserID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ext)
}

func (h *ExtensionHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	exts, err := models.ListAllExtensions(h.DB)
	if err != nil {
		http.Error(w, `{"error":"failed to list extensions"}`, http.StatusInternalServerError)
		return
	}
	if exts == nil {
		exts = []models.Extension{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exts)
}

type adminExtensionUpdateRequest struct {
	CallerID    string `json:"callerid"`
	SIPPassword string `json:"sip_password"`
}

func (h *ExtensionHandler) AdminUpdate(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	extNum, err := strconv.Atoi(chi.URLParam(r, "ext"))
	if err != nil {
		http.Error(w, `{"error":"invalid extension number"}`, http.StatusBadRequest)
		return
	}

	var req adminExtensionUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := models.UpdateExtension(h.DB, extNum, req.CallerID, req.SIPPassword); err != nil {
		http.Error(w, `{"error":"failed to update extension"}`, http.StatusInternalServerError)
		return
	}

	models.AuditLog(h.DB, claims.UserID, "admin_update_extension", fmt.Sprintf("ext=%d", extNum))

	if err := asterisk.SyncAndReload(h.DB, h.Config); err != nil {
		fmt.Printf("WARNING: asterisk sync failed: %v\n", err)
	}

	ext, err := models.GetExtensionByNumber(h.DB, extNum)
	if err != nil {
		http.Error(w, `{"error":"extension not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ext)
}

func (h *ExtensionHandler) AdminDelete(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	extNum, err := strconv.Atoi(chi.URLParam(r, "ext"))
	if err != nil {
		http.Error(w, `{"error":"invalid extension number"}`, http.StatusBadRequest)
		return
	}

	if err := models.DeleteExtensionByNumber(h.DB, extNum); err != nil {
		http.Error(w, `{"error":"extension not found"}`, http.StatusNotFound)
		return
	}

	models.AuditLog(h.DB, claims.UserID, "admin_delete_extension", fmt.Sprintf("ext=%d", extNum))

	if err := asterisk.SyncAndReload(h.DB, h.Config); err != nil {
		fmt.Printf("WARNING: asterisk sync failed: %v\n", err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ExtensionHandler) ForceSync(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())

	if err := asterisk.SyncAndReload(h.DB, h.Config); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"sync failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	models.AuditLog(h.DB, claims.UserID, "force_sync", "")

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *ExtensionHandler) ListBlocked(w http.ResponseWriter, r *http.Request) {
	blocked, err := models.ListBlockedExtensions(h.DB)
	if err != nil {
		http.Error(w, `{"error":"failed to list blocked extensions"}`, http.StatusInternalServerError)
		return
	}
	if blocked == nil {
		blocked = []models.BlockedExtension{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(blocked)
}

type blockRequest struct {
	Extension int    `json:"extension"`
	Reason    string `json:"reason"`
}

func (h *ExtensionHandler) BlockExtension(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())

	var req blockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Extension <= 0 {
		http.Error(w, `{"error":"invalid extension number"}`, http.StatusBadRequest)
		return
	}

	if err := models.BlockExtension(h.DB, req.Extension, req.Reason); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			http.Error(w, `{"error":"extension already blocked"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error":"failed to block extension"}`, http.StatusInternalServerError)
		return
	}

	models.AuditLog(h.DB, claims.UserID, "block_extension", fmt.Sprintf("ext=%d reason=%s", req.Extension, req.Reason))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *ExtensionHandler) UnblockExtension(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	extNum, err := strconv.Atoi(chi.URLParam(r, "ext"))
	if err != nil {
		http.Error(w, `{"error":"invalid extension number"}`, http.StatusBadRequest)
		return
	}

	if err := models.UnblockExtension(h.DB, extNum); err != nil {
		http.Error(w, `{"error":"failed to unblock extension"}`, http.StatusInternalServerError)
		return
	}

	models.AuditLog(h.DB, claims.UserID, "unblock_extension", fmt.Sprintf("ext=%d", extNum))
	w.WriteHeader(http.StatusNoContent)
}

type directoryEntry struct {
	Extension int    `json:"extension"`
	Name      string `json:"name"`
}

func (h *ExtensionHandler) DirectoryJSON(w http.ResponseWriter, r *http.Request) {
	exts, err := models.ListAllExtensions(h.DB)
	if err != nil {
		http.Error(w, `{"error":"failed to list extensions"}`, http.StatusInternalServerError)
		return
	}

	entries := make([]directoryEntry, 0, len(exts))
	for _, ext := range exts {
		name := ext.CallerID
		if name == "" {
			name = fmt.Sprintf("%d", ext.Extension)
		}
		entries = append(entries, directoryEntry{
			Extension: ext.Extension,
			Name:      name,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

type ciscoDirectoryEntry struct {
	XMLName   xml.Name `xml:"DirectoryEntry"`
	Name      string   `xml:"Name"`
	Telephone string   `xml:"Telephone"`
}

type ciscoIPPhoneDirectory struct {
	XMLName xml.Name              `xml:"CiscoIPPhoneDirectory"`
	Title   string                `xml:"Title"`
	Prompt  string                `xml:"Prompt"`
	Entries []ciscoDirectoryEntry `xml:",any"`
}

func (h *ExtensionHandler) Directory(w http.ResponseWriter, r *http.Request) {
	exts, err := models.ListAllExtensions(h.DB)
	if err != nil {
		http.Error(w, "failed to list extensions", http.StatusInternalServerError)
		return
	}

	dir := ciscoIPPhoneDirectory{
		Title:  "Company Directory",
		Prompt: "Select to dial",
	}

	for _, ext := range exts {
		name := ext.CallerID
		if name == "" {
			name = fmt.Sprintf("%d", ext.Extension)
		}
		dir.Entries = append(dir.Entries, ciscoDirectoryEntry{
			Name:      name,
			Telephone: fmt.Sprintf("%d", ext.Extension),
		})
	}

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.Write([]byte(xml.Header))
	xml.NewEncoder(w).Encode(dir)
}
