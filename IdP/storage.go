package main

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"io"
)

type dbUser struct {
	uid       string
	firstname string
	lastname  string
	email     string
	pwd       string // this is a SHA1 hash
}

type User struct {
	UserID    string `json:"uid"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

type storage struct {
	db *sql.DB
}

// NewStorage returns a new storage client which is capable of talking to a mySQL DB.
func NewStorage(dsn string) (*storage, error) {
	pool, err := sql.Open("mysql", dsn)
	if err != nil {
		log.WithError(err).Error("Failed to sql.Open.")
		return nil, err
	}
	err = pool.PingContext(context.Background())
	if err != nil {
		log.WithError(err).Error("Failed to connect to DB.")
		return nil, err
	}
	return &storage{db: pool}, nil
}

// GetUser retrieves a specific user from the database. It returns sql.ErrNoRows if the user was not
// found.
func (s *storage) GetUser(ctx context.Context, userID string) (User, error) {
	u := dbUser{}
	row := s.db.QueryRowContext(ctx, `SELECT uid, firstname, lastname, email FROM users WHERE uid=?`, userID)
	err := row.Scan(&u.uid, &u.firstname, &u.lastname, &u.email)
	if err != nil {
		if err != sql.ErrNoRows {
			log.WithError(err).Error("Failed to query DB for user.")
		}
		return User{}, err
	}
	return userFromDBUser(u), nil
}

// ChangePassword allows a user to change their password.
func (s *storage) ChangePassword(ctx context.Context, userID string, password string) error {
	//hash password
	h := sha1.New()
	h.Write([]byte(password))
	pwHash := hex.EncodeToString(h.Sum(nil))

	// Check if user exists
	var uid string
	row := s.db.QueryRowContext(ctx, `SELECT uid FROM users WHERE uid=?`, userID)
	err := row.Scan(&uid)
	if err != nil {
		if err != sql.ErrNoRows {
			log.WithError(err).Error("Failed to query DB for user.")
		} else {
			log.WithError(err).Error("No user with uID found.")
		}
		return err
	}

	//change password.
	_, err = s.db.ExecContext(ctx, `UPDATE users SET pwd = ? WHERE uid=?`, pwHash, userID)
	if err != nil {
		if err != sql.ErrNoRows {
			log.WithError(err).Error("Failed to change password.")
		}
		return err
	}
	return nil
}

func (s *storage) EdditUser(ctx context.Context, user dbUser) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET firstname = ?, lastname = ?, email = ? WHERE uid=?`, user.firstname, user.lastname, user.email, user.uid)
	if err != nil {
		if err != sql.ErrNoRows {
			log.WithError(err).Error("Failed to edit user")
		}
		return err
	}
	return nil
}

// Login returns true if the hashed password matches our database record.
func (s *storage) Login(ctx context.Context, userID string, password string) bool {
	//hash password
	h := sha1.New()
	io.WriteString(h, password)
	pwHash := hex.EncodeToString(h.Sum(nil))

	row := s.db.QueryRowContext(ctx, `SELECT uid FROM users WHERE uid=? AND pwd=?`, userID, pwHash)

	var uid string
	err := row.Scan(&uid)
	if err != nil {
		if err != sql.ErrNoRows {
			log.WithError(err).Error("Failed to query DB for user.")
		} else {
			log.WithError(err).Error("Failed login attempt")
		}
		return false
	}
	if uid == userID {
		return true
	} else {
		return false
	}
}

func userFromDBUser(u dbUser) User {
	return User{
		UserID:    u.uid,
		FirstName: u.firstname,
		LastName:  u.lastname,
		Email:     u.email,
	}
}
