package models

import (
	"database/sql"
	"time"
)

type BlockedExtension struct {
	ID        int       `json:"id"`
	Extension int       `json:"extension"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

func IsExtensionBlocked(db *sql.DB, ext int) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM blocked_extensions WHERE extension = ?", ext).Scan(&count)
	return count > 0, err
}

func ListBlockedExtensions(db *sql.DB) ([]BlockedExtension, error) {
	rows, err := db.Query("SELECT id, extension, reason, created_at FROM blocked_extensions ORDER BY extension")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocked []BlockedExtension
	for rows.Next() {
		var b BlockedExtension
		if err := rows.Scan(&b.ID, &b.Extension, &b.Reason, &b.CreatedAt); err != nil {
			return nil, err
		}
		blocked = append(blocked, b)
	}
	return blocked, rows.Err()
}

func BlockExtension(db *sql.DB, ext int, reason string) error {
	_, err := db.Exec("INSERT INTO blocked_extensions (extension, reason) VALUES (?, ?)", ext, reason)
	return err
}

func UnblockExtension(db *sql.DB, ext int) error {
	_, err := db.Exec("DELETE FROM blocked_extensions WHERE extension = ?", ext)
	return err
}
