package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"html"
	"net/http"
	"time"
)

type adminClient interface {
	GetLoginInfo(challenge string) (LoginInfo, error)
	AcceptLogin(challenge string, req AcceptLoginRequest) (AcceptLoginResponse, error)
	GetConsentInfo(challenge string) (ConsentInfo, error)
	AcceptConsent(challenge string, req AcceptConsentRequest) (AcceptConsentResponse, error)
}

type Server struct {
	router *mux.Router
	client adminClient
	db     Storage
}

type Storage interface {
	GetUser(ctx context.Context, userID string) (User, error)
	ChangePassword(ctx context.Context, userID string, password string) error
	Login(ctx context.Context, userID string, password string) bool
}

var adminURl = flag.String("admin-url", "https://localhost:9001", "url of the hydra admin api")
var listen = flag.String("listen", ":8088", "on what url to start the server on")
var dsn = flag.String("dsn", "", "DSN of the DB to connect to: user:password@/dbname")

func main() {
	log.SetLevel(log.TraceLevel) // log all the things
	flag.Parse()
	// Setting up client to communicate with hydra
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := HydraClient{client: &http.Client{Transport: tr}, adminUrl: *adminURl}
	if *dsn == "" {
		log.Error("Empty DSN passed.")
	}
	db, err := NewStorage(*dsn)
	if err != nil {
		log.WithError(err).Fatal("Failed to create storage component.")
	}

	// Prepare HTTP server
	r := mux.NewRouter()
	ser := Server{client: &client, router: r, db:db}
	r.HandleFunc("/login", ser.Login)
	r.HandleFunc("/consent", ser.Consent)
	r.HandleFunc("/user/{id}", ser.GetUser)

	// Kind of a smoke test.
	u, err := ser.db.GetUser(context.Background(), "a3")
	if err != nil {
		log.WithError(err).Fatalf("Failed to execute known good query")
	} else {
		log.WithField("user", u).Info("Found user.")
	}
	// Run
	log.Fatal(http.ListenAndServe(*listen, r))
}

func (s Server) Login(w http.ResponseWriter, r *http.Request) {
	log.Debug("%s, %q", r.Method, html.EscapeString(r.URL.Path))
	//keys[0] contains the challenge
	keys, ok := r.URL.Query()["login_challenge"]
	if !ok {
		log.Info("No login challenge provided")
		s.httpBadRequest(w, "no login challenge provided")
		return
	}
	_, err := s.client.GetLoginInfo(keys[0])
	if err != nil {
		log.Error("Error getting login info", "error", err)
		s.httpInternalError(w, err)
		return
	}

	authenticated := false
	skip := false

	// TODO(Fischi): We don't actually use the information we get. We should

	if r.Method == http.MethodGet && !skip {
		// TODO(Fischi): Show login screen
		authenticated = true
	}

	if r.Method == http.MethodPost {
		// TODO(Fischi): Check authentication
		// use db.Login
	}

	// Accept login request
	if authenticated {
		acceptBody := AcceptLoginRequest{Subject: "sub", Remember: false, RememberFor: 300}
		accRes, err := s.client.AcceptLogin(keys[0], acceptBody)
		if err != nil {
			log.Error("Error accepting login", "error", err)
			s.httpInternalError(w, err)
			return
		}

		// redirect
		http.Redirect(w, r, accRes.RedirectTo, http.StatusFound)
	}
	s.httpUnauthorized(w)
}

func (s Server) Consent(w http.ResponseWriter, r *http.Request) {
	log.Debug("%s, %q", r.Method, html.EscapeString(r.URL.Path))
	//keys[0] contains the challenge
	keys, ok := r.URL.Query()["consent_challenge"]
	if !ok {
		log.Info("No consent challenge provided")
		s.httpBadRequest(w, "no login challenge provided")
		return
	}
	challenge := keys[0]

	//fetch information about the request
	_, err := s.client.GetConsentInfo(challenge)
	if err != nil {
		log.Error("Error getting consent info", "error", err)
		s.httpInternalError(w, err)
		return
	}

	// TODO(Fischi): We don't actually use the information we get. We should

	//TODO: check, whether the user gave consent and user should give consent if not...
	//TODO: check whether skip is true... only show UI if skip false.

	requestBody := AcceptConsentRequest{GrantScope: []string{"scope"}, GrantAccessTokenAudience: []string{"hi"}, Remember: false, RememberFor: 300}
	conRes, err := s.client.AcceptConsent(keys[0], requestBody)
	if err != nil {
		s.httpInternalError(w, err)
		return
	}
	http.Redirect(w, r, conRes.RedirectTo, http.StatusFound)
}

func (s Server) GetUser(w http.ResponseWriter, r *http.Request) {
	log.Debug("%s, %q", r.Method, html.EscapeString(r.URL.Path))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		log.Error("GetUser request without ID.")
		s.httpBadRequest(w, "missing user id")
		return
	}
	u, err := s.db.GetUser(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			log.WithField("user-id", id).Warn("User not found.")
			s.httpNotFound(w)
			return
		}
		log.WithError(err).WithField("user-id", id).Error("Failed to GetUser.")
		s.httpInternalError(w, fmt.Errorf("failed to get user"))
		return
	}

	w.Header().Set("content-type", "application/json")
	err = json.NewEncoder(w).Encode(u)
	if err != nil {
		s.httpInternalError(w, err)
	}
}

func (s Server) httpInternalError(w http.ResponseWriter, e error) {
	if e != nil {
		log.Errorf("Error: %v\n", e)
	}
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (s Server) httpNotFound(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func (s Server) httpUnauthorized(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
}

func (s Server) httpBadRequest(w http.ResponseWriter, error string) {
	http.Error(w, error, http.StatusBadRequest)
}
