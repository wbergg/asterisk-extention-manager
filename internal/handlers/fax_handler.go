package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/wbergg/asterisk-extention-manager/internal/auth"
	"github.com/wbergg/asterisk-extention-manager/internal/config"
	"github.com/wbergg/asterisk-extention-manager/internal/fax"
	"github.com/wbergg/asterisk-extention-manager/internal/models"
)

type FaxHandler struct {
	DB     *sql.DB
	Config *config.Config
}

type faxDestination struct {
	Extension int    `json:"extension"`
	Name      string `json:"name"`
}

func (h *FaxHandler) ListFaxDestinations(w http.ResponseWriter, r *http.Request) {
	exts, err := models.ListFaxExtensions(h.DB)
	if err != nil {
		http.Error(w, `{"error":"failed to list fax destinations"}`, http.StatusInternalServerError)
		return
	}

	dests := make([]faxDestination, 0, len(exts))
	for _, ext := range exts {
		name := ext.CallerID
		if name == "" {
			name = fmt.Sprintf("%d", ext.Extension)
		}
		dests = append(dests, faxDestination{
			Extension: ext.Extension,
			Name:      name,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dests)
}

var allowedFaxExts = map[string]bool{
	".pdf":  true,
	".png":  true,
	".jpg":  true,
	".jpeg": true,
}

func (h *FaxHandler) SendFax(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, `{"error":"failed to parse form (max 10MB)"}`, http.StatusBadRequest)
		return
	}

	subject := r.FormValue("subject")
	destStr := r.FormValue("destination")
	if destStr == "" {
		http.Error(w, `{"error":"destination is required"}`, http.StatusBadRequest)
		return
	}

	destExt, err := strconv.Atoi(destStr)
	if err != nil {
		http.Error(w, `{"error":"invalid destination extension"}`, http.StatusBadRequest)
		return
	}

	// Validate destination is a fax extension
	faxExts, err := models.ListFaxExtensions(h.DB)
	if err != nil {
		http.Error(w, `{"error":"failed to validate destination"}`, http.StatusInternalServerError)
		return
	}
	validDest := false
	for _, fe := range faxExts {
		if fe.Extension == destExt {
			validDest = true
			break
		}
	}
	if !validDest {
		http.Error(w, `{"error":"invalid fax destination"}`, http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, `{"error":"file is required"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileExt := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedFaxExts[fileExt] {
		http.Error(w, `{"error":"unsupported file type, use PDF, PNG, or JPG"}`, http.StatusBadRequest)
		return
	}

	// Create job directory (resolve to absolute path so Asterisk can find the TIFF)
	storagePath, err := filepath.Abs(h.Config.FaxStoragePath)
	if err != nil {
		http.Error(w, `{"error":"failed to resolve storage path"}`, http.StatusInternalServerError)
		return
	}
	jobDir := filepath.Join(storagePath, fmt.Sprintf("%d", claims.UserID), fmt.Sprintf("%d", time.Now().UnixNano()))
	if err := os.MkdirAll(jobDir, 0775); err != nil {
		http.Error(w, `{"error":"failed to create storage directory"}`, http.StatusInternalServerError)
		return
	}

	// Save uploaded file
	originalName := "original" + fileExt
	originalPath := filepath.Join(jobDir, originalName)
	dst, err := os.Create(originalPath)
	if err != nil {
		http.Error(w, `{"error":"failed to save file"}`, http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(dst, file); err != nil {
		dst.Close()
		http.Error(w, `{"error":"failed to save file"}`, http.StatusInternalServerError)
		return
	}
	dst.Close()

	// Create DB record
	job := &models.FaxJob{
		UserID:         claims.UserID,
		Subject:        subject,
		DestinationExt: destExt,
		OriginalFile:   originalPath,
		Status:         "converting",
	}
	if err := models.CreateFaxJob(h.DB, job); err != nil {
		http.Error(w, `{"error":"failed to create fax job"}`, http.StatusInternalServerError)
		return
	}

	// Convert to TIFF
	tiffPath, err := fax.ConvertToTIFF(originalPath, jobDir)
	if err != nil {
		models.UpdateFaxJobStatus(h.DB, job.ID, "failed", err.Error())
		http.Error(w, fmt.Sprintf(`{"error":"conversion failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	job.TIFFFile = tiffPath

	// Write .call file
	callFileName, err := fax.WriteCallFile(h.Config.FaxSpoolPath, job.ID, destExt, tiffPath, jobDir)
	if err != nil {
		models.UpdateFaxJobStatus(h.DB, job.ID, "failed", err.Error())
		http.Error(w, fmt.Sprintf(`{"error":"failed to queue fax: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	job.CallFile = callFileName
	job.Status = "queued"

	models.UpdateFaxJobStatus(h.DB, job.ID, "queued", "")
	// Also update tiff_file and call_file in DB
	h.DB.Exec("UPDATE fax_jobs SET tiff_file=?, call_file=? WHERE id=?", tiffPath, callFileName, job.ID)

	models.AuditLog(h.DB, claims.UserID, "send_fax", fmt.Sprintf("dest=%d subject=%s", destExt, subject))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(job)
}

func (h *FaxHandler) ListFaxJobs(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())

	jobs, err := models.ListFaxJobsByUser(h.DB, claims.UserID)
	if err != nil {
		http.Error(w, `{"error":"failed to list fax jobs"}`, http.StatusInternalServerError)
		return
	}

	// Update status for queued jobs by checking spool
	for i := range jobs {
		if jobs[i].Status == "queued" && jobs[i].CallFile != "" {
			if !fax.CheckCallFileExists(h.Config.FaxSpoolPath, jobs[i].CallFile) {
				jobs[i].Status = "attempted"
				models.UpdateFaxJobStatus(h.DB, jobs[i].ID, "attempted", "")
			}
		}
	}

	if jobs == nil {
		jobs = []models.FaxJob{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}
