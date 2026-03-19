package models

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"time"
)

type Extension struct {
	ID            int       `json:"id"`
	Extension     int       `json:"extension"`
	UserID        int       `json:"user_id"`
	SIPUsername   string    `json:"sip_username"`
	SIPPassword   string    `json:"sip_password"`
	CallerID      string    `json:"callerid"`
	Context       string    `json:"context"`
	DirectoryOnly bool      `json:"directory_only"`
	CreatedAt     time.Time `json:"created_at"`
}

func GenerateSIPPassword() (string, error) {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 10)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = chars[b[i]%byte(len(chars))]
	}
	return string(b), nil
}

func CreateExtension(db *sql.DB, ext *Extension, sipPassword string) error {
	if sipPassword != "" {
		ext.SIPPassword = sipPassword
	} else {
		sipPass, err := GenerateSIPPassword()
		if err != nil {
			return fmt.Errorf("generate sip password: %w", err)
		}
		ext.SIPPassword = sipPass
	}
	ext.SIPUsername = fmt.Sprintf("%d", ext.Extension)
	ext.Context = "internal"

	res, err := db.Exec(
		"INSERT INTO extensions (extension, user_id, sip_username, sip_password, callerid, context) VALUES (?, ?, ?, ?, ?, ?)",
		ext.Extension, ext.UserID, ext.SIPUsername, ext.SIPPassword, ext.CallerID, ext.Context,
	)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	ext.ID = int(id)
	return nil
}

func CreateDirectoryOnlyExtension(db *sql.DB, ext *Extension) error {
	ext.SIPUsername = fmt.Sprintf("%d", ext.Extension)
	ext.SIPPassword = ""
	ext.Context = "internal"
	ext.DirectoryOnly = true

	res, err := db.Exec(
		"INSERT INTO extensions (extension, user_id, sip_username, sip_password, callerid, context, directory_only) VALUES (?, ?, ?, ?, ?, ?, 1)",
		ext.Extension, ext.UserID, ext.SIPUsername, ext.SIPPassword, ext.CallerID, ext.Context,
	)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	ext.ID = int(id)
	return nil
}

func ListExtensionsByUser(db *sql.DB, userID int) ([]Extension, error) {
	rows, err := db.Query(
		"SELECT id, extension, user_id, sip_username, sip_password, callerid, context, directory_only, created_at FROM extensions WHERE user_id = ? ORDER BY extension",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exts []Extension
	for rows.Next() {
		var e Extension
		if err := rows.Scan(&e.ID, &e.Extension, &e.UserID, &e.SIPUsername, &e.SIPPassword, &e.CallerID, &e.Context, &e.DirectoryOnly, &e.CreatedAt); err != nil {
			return nil, err
		}
		exts = append(exts, e)
	}
	return exts, rows.Err()
}

func ListAllExtensions(db *sql.DB) ([]Extension, error) {
	rows, err := db.Query("SELECT id, extension, user_id, sip_username, sip_password, callerid, context, directory_only, created_at FROM extensions ORDER BY extension")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exts []Extension
	for rows.Next() {
		var e Extension
		if err := rows.Scan(&e.ID, &e.Extension, &e.UserID, &e.SIPUsername, &e.SIPPassword, &e.CallerID, &e.Context, &e.DirectoryOnly, &e.CreatedAt); err != nil {
			return nil, err
		}
		exts = append(exts, e)
	}
	return exts, rows.Err()
}

func GetExtension(db *sql.DB, extNum int, userID int) (*Extension, error) {
	e := &Extension{}
	err := db.QueryRow(
		"SELECT id, extension, user_id, sip_username, sip_password, callerid, context, directory_only, created_at FROM extensions WHERE extension = ? AND user_id = ?",
		extNum, userID,
	).Scan(&e.ID, &e.Extension, &e.UserID, &e.SIPUsername, &e.SIPPassword, &e.CallerID, &e.Context, &e.DirectoryOnly, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func DeleteExtension(db *sql.DB, extNum int, userID int) error {
	res, err := db.Exec("DELETE FROM extensions WHERE extension = ? AND user_id = ?", extNum, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func GetExtensionByNumber(db *sql.DB, extNum int) (*Extension, error) {
	e := &Extension{}
	err := db.QueryRow(
		"SELECT id, extension, user_id, sip_username, sip_password, callerid, context, directory_only, created_at FROM extensions WHERE extension = ?",
		extNum,
	).Scan(&e.ID, &e.Extension, &e.UserID, &e.SIPUsername, &e.SIPPassword, &e.CallerID, &e.Context, &e.DirectoryOnly, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func UpdateExtension(db *sql.DB, extNum int, callerid string, sipPassword string) error {
	if sipPassword != "" {
		_, err := db.Exec("UPDATE extensions SET callerid=?, sip_password=? WHERE extension=?", callerid, sipPassword, extNum)
		return err
	}
	_, err := db.Exec("UPDATE extensions SET callerid=? WHERE extension=?", callerid, extNum)
	return err
}

func DeleteExtensionByNumber(db *sql.DB, extNum int) error {
	res, err := db.Exec("DELETE FROM extensions WHERE extension = ?", extNum)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func AuditLog(db *sql.DB, userID int, action, detail string) {
	_, _ = db.Exec("INSERT INTO audit_log (user_id, action, detail) VALUES (?, ?, ?)", userID, action, detail)
}
