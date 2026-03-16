package models

import (
	"database/sql"
	"fmt"
	"time"
)

type User struct {
	ID             int       `json:"id"`
	Username       string    `json:"username"`
	Password       string    `json:"-"`
	Role           string    `json:"role"`
	MinExt         int       `json:"min_ext"`
	MaxExt         int       `json:"max_ext"`
	CallLogAccess  bool      `json:"call_log_access"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	u := &User{}
	err := db.QueryRow(
		"SELECT id, username, password, role, min_ext, max_ext, call_log_access, created_at, updated_at FROM users WHERE username = ?",
		username,
	).Scan(&u.ID, &u.Username, &u.Password, &u.Role, &u.MinExt, &u.MaxExt, &u.CallLogAccess, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func GetUserByID(db *sql.DB, id int) (*User, error) {
	u := &User{}
	err := db.QueryRow(
		"SELECT id, username, password, role, min_ext, max_ext, call_log_access, created_at, updated_at FROM users WHERE id = ?",
		id,
	).Scan(&u.ID, &u.Username, &u.Password, &u.Role, &u.MinExt, &u.MaxExt, &u.CallLogAccess, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func ListUsers(db *sql.DB) ([]User, error) {
	rows, err := db.Query("SELECT id, username, password, role, min_ext, max_ext, call_log_access, created_at, updated_at FROM users ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Password, &u.Role, &u.MinExt, &u.MaxExt, &u.CallLogAccess, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func CreateUser(db *sql.DB, u *User) error {
	res, err := db.Exec(
		"INSERT INTO users (username, password, role, min_ext, max_ext, call_log_access) VALUES (?, ?, ?, ?, ?, ?)",
		u.Username, u.Password, u.Role, u.MinExt, u.MaxExt, u.CallLogAccess,
	)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	id, _ := res.LastInsertId()
	u.ID = int(id)
	return nil
}

func UpdateUser(db *sql.DB, u *User) error {
	if u.Password != "" {
		_, err := db.Exec(
			"UPDATE users SET username=?, password=?, role=?, min_ext=?, max_ext=?, call_log_access=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
			u.Username, u.Password, u.Role, u.MinExt, u.MaxExt, u.CallLogAccess, u.ID,
		)
		return err
	}
	_, err := db.Exec(
		"UPDATE users SET username=?, role=?, min_ext=?, max_ext=?, call_log_access=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		u.Username, u.Role, u.MinExt, u.MaxExt, u.CallLogAccess, u.ID,
	)
	return err
}

func UpdateUserPassword(db *sql.DB, id int, hashedPassword string) error {
	_, err := db.Exec("UPDATE users SET password=?, updated_at=CURRENT_TIMESTAMP WHERE id=?", hashedPassword, id)
	return err
}

func DeleteUser(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}
