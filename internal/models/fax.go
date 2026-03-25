package models

import (
	"database/sql"
	"time"
)

type FaxJob struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	Subject        string    `json:"subject"`
	DestinationExt int       `json:"destination_ext"`
	OriginalFile   string    `json:"original_file"`
	TIFFFile       string    `json:"tiff_file"`
	CallFile       string    `json:"call_file"`
	Status         string    `json:"status"`
	ErrorMessage   string    `json:"error_message"`
	CreatedAt      time.Time `json:"created_at"`
}

func CreateFaxJob(db *sql.DB, job *FaxJob) error {
	res, err := db.Exec(
		"INSERT INTO fax_jobs (user_id, subject, destination_ext, original_file, tiff_file, call_file, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
		job.UserID, job.Subject, job.DestinationExt, job.OriginalFile, job.TIFFFile, job.CallFile, job.Status,
	)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	job.ID = int(id)
	job.CreatedAt = time.Now()
	return nil
}

func ListFaxJobsByUser(db *sql.DB, userID int) ([]FaxJob, error) {
	rows, err := db.Query(
		"SELECT id, user_id, subject, destination_ext, original_file, tiff_file, call_file, status, error_message, created_at FROM fax_jobs WHERE user_id = ? ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []FaxJob
	for rows.Next() {
		var j FaxJob
		if err := rows.Scan(&j.ID, &j.UserID, &j.Subject, &j.DestinationExt, &j.OriginalFile, &j.TIFFFile, &j.CallFile, &j.Status, &j.ErrorMessage, &j.CreatedAt); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

func UpdateFaxJobStatus(db *sql.DB, id int, status string, errorMessage string) error {
	_, err := db.Exec("UPDATE fax_jobs SET status=?, error_message=? WHERE id=?", status, errorMessage, id)
	return err
}

func ListFaxExtensions(db *sql.DB) ([]Extension, error) {
	rows, err := db.Query(
		"SELECT id, extension, user_id, sip_username, sip_password, callerid, context, directory_only, created_at FROM extensions WHERE LOWER(callerid) LIKE '%fax%' ORDER BY extension",
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
