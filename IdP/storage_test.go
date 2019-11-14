package main

import (
	"context"
	"flag"
	"testing"
	log "github.com/sirupsen/logrus"
)

var testDSN = flag.String("testdsn", "user:pass@(localhost)/imovies", "DSN of the DB to connect to: user:password@/dbname")

// Simple complete test to check for obvious errors
func TestStorage(t *testing.T) {
	flag.Parse()
	db, err := NewStorage(*testDSN)
	if err != nil {
		t.Errorf("Failed to create storage component. %v", err)
		return
	}

	//We assume the user a3 already exists.
	log.Infof("Getting user a3")
	u, err := db.GetUser(context.Background(), "a3")
	if err != nil {
		t.Errorf("Failed to get user. %v", err)
		return
	}
	if u.Email != "anderson@imovies.ch" {
		t.Errorf("A3 has unexpected email. Expected anderson@imovies.ch, Got %s", u.Email)
		return
	}

	// Check if we can Login
	log.Infof("logging in with old pw")
	logedin := db.Login(context.Background(), "a3", "Astrid")
	if !logedin {
		t.Error("Could not login a3 with known good password")
		return
	}

	// Check if we can change pw
	log.Infof("Changing pw")
	err = db.ChangePassword(context.Background(), "a3", "Astrid2")
	if !logedin {
		t.Errorf("Could not change pw, %v", err)
		return
	}
	log.Infof("Login with new pw")
	// Check if we can Login with new pw
	logedin = db.Login(context.Background(), "a3", "Astrid2")
	if !logedin {
		t.Error("Could not login a3 with changed password")
		return
	}
	// Check if we cannot Login with old pw
	log.Infof("Try to login with old pw")
	logedin = db.Login(context.Background(), "a3", "Astrid")
	if logedin {
		t.Error("Could login with wrong pw")
		return
	}
	// change pw back
	log.Infof("Change back pw")
	err = db.ChangePassword(context.Background(), "a3", "Astrid")
	if err != nil {
		t.Errorf("Could not change pw, %v", err)
		return
	}
	// Check if we can Login with old pw
	log.Infof("Login with original pw")
	logedin = db.Login(context.Background(), "a3", "Astrid")
	if !logedin {
		t.Error("Could not login a3 with old password")
		return
	}

}
