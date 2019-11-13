package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
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
	// TODO(bimmlerd)
	return fmt.Errorf("not implemented")
}

// Login returns true if the hashed password matches our database record.
func (s *storage) Login(ctx context.Context, userID string, password string) bool {
	// TODO(bimmlerd)
	return false
}

func userFromDBUser(u dbUser) User {
	return User{
		UserID:    u.uid,
		FirstName: u.firstname,
		LastName:  u.lastname,
		Email:     u.email,
	}
}
