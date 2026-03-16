package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/wbergg/asterisk-extention-manager/internal/auth"
	"github.com/wbergg/asterisk-extention-manager/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	DB *sql.DB
}

type userRequest struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	Role          string `json:"role"`
	MinExt        int    `json:"min_ext"`
	MaxExt        int    `json:"max_ext"`
	CallLogAccess *bool  `json:"call_log_access"`
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := models.ListUsers(h.DB)
	if err != nil {
		http.Error(w, `{"error":"failed to list users"}`, http.StatusInternalServerError)
		return
	}
	if users == nil {
		users = []models.User{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req userRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, `{"error":"username and password required"}`, http.StatusBadRequest)
		return
	}
	if req.Role == "" {
		req.Role = "user"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error":"failed to hash password"}`, http.StatusInternalServerError)
		return
	}

	callLogAccess := true
	if req.CallLogAccess != nil {
		callLogAccess = *req.CallLogAccess
	}

	user := &models.User{
		Username:      req.Username,
		Password:      string(hash),
		Role:          req.Role,
		MinExt:        req.MinExt,
		MaxExt:        req.MaxExt,
		CallLogAccess: callLogAccess,
	}

	if err := models.CreateUser(h.DB, user); err != nil {
		http.Error(w, `{"error":"failed to create user (username may be taken)"}`, http.StatusConflict)
		return
	}

	claims := auth.ClaimsFromContext(r.Context())
	models.AuditLog(h.DB, claims.UserID, "create_user", user.Username)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid user id"}`, http.StatusBadRequest)
		return
	}

	var req userRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	callLogAccess := true
	if req.CallLogAccess != nil {
		callLogAccess = *req.CallLogAccess
	}

	user := &models.User{
		ID:            id,
		Username:      req.Username,
		Role:          req.Role,
		MinExt:        req.MinExt,
		MaxExt:        req.MaxExt,
		CallLogAccess: callLogAccess,
	}

	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, `{"error":"failed to hash password"}`, http.StatusInternalServerError)
			return
		}
		user.Password = string(hash)
	}

	if err := models.UpdateUser(h.DB, user); err != nil {
		http.Error(w, `{"error":"failed to update user"}`, http.StatusInternalServerError)
		return
	}

	claims := auth.ClaimsFromContext(r.Context())
	models.AuditLog(h.DB, claims.UserID, "update_user", user.Username)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid user id"}`, http.StatusBadRequest)
		return
	}

	if err := models.DeleteUser(h.DB, id); err != nil {
		http.Error(w, `{"error":"failed to delete user"}`, http.StatusInternalServerError)
		return
	}

	claims := auth.ClaimsFromContext(r.Context())
	models.AuditLog(h.DB, claims.UserID, "delete_user", strconv.Itoa(id))

	w.WriteHeader(http.StatusNoContent)
}
