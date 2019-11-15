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
	"html/template"
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
	router          *mux.Router
	client          adminClient
	db              Storage
	templateLogin   *template.Template
	templateConsent *template.Template
}

type Storage interface {
	GetUser(ctx context.Context, userID string) (User, error)
	ChangePassword(ctx context.Context, userID string, password string) error
	Login(ctx context.Context, userID string, password string) bool
	EditUser(ctx context.Context, user User) error
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
	ser := Server{client: &client, router: r, db: db}

	// Prepare template
	ser.templateLogin, err = template.ParseFiles("./template/login.html")
	if err != nil {
		log.WithError(err).Fatal("Failed to parse login template.")
	}
	ser.templateConsent, err = template.ParseFiles("./template/consent.html")
	if err != nil {
		log.WithError(err).Fatal("Failed to parse consent template.")
	}

	r.HandleFunc("/login", ser.Login)
	r.HandleFunc("/consent", ser.Consent)
	r.HandleFunc("/user/{id}", ser.GetUser).Methods("GET")
	r.HandleFunc("/user/{id}", ser.EditUser).Methods("PUT")
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
	log.Debugf("%s, %q", r.Method, html.EscapeString(r.URL.Path))
	//keys[0] contains the challenge
	keys, ok := r.URL.Query()["login_challenge"]
	if !ok {
		log.Info("No login challenge provided")
		s.httpBadRequest(w, "no login challenge provided")
		return
	}
	info, err := s.client.GetLoginInfo(keys[0])
	if err != nil {
		log.Error("Error getting login info", "error", err)
		s.httpInternalError(w, err)
		return
	}

	authenticated := info.Skip
	username := info.Subject

	// TODO(Fischi): We don't actually use the information we get. We should

	if r.Method == http.MethodGet && !info.Skip {
		err := s.templateLogin.Execute(w, map[string]interface{}{})
		if err != nil {
			s.httpInternalError(w, err)
		}
		return
	}

	if r.Method == http.MethodPost {
		r.ParseForm()
		username = r.FormValue("username")
		password := r.FormValue("password")
		authenticated = s.db.Login(r.Context(), username, password)
		log.WithField("username", username).Info("Tried to log in")
	}

	// Accept login request
	if authenticated {
		log.WithField("username", username).Info("Authenticated")
		acceptBody := AcceptLoginRequest{Subject: username, Remember: false, RememberFor: 300}
		accRes, err := s.client.AcceptLogin(keys[0], acceptBody)
		if err != nil {
			log.Error("Error accepting login", "error", err)
			s.httpInternalError(w, err)
			return
		}

		// redirect
		http.Redirect(w, r, accRes.RedirectTo, http.StatusFound)
		return
	}
	s.httpUnauthorized(w)
}

func (s Server) Consent(w http.ResponseWriter, r *http.Request) {
	log.Debugf("%s, %q", r.Method, html.EscapeString(r.URL.Path))
	//keys[0] contains the challenge
	keys, ok := r.URL.Query()["consent_challenge"]
	if !ok {
		log.Info("No consent challenge provided")
		s.httpBadRequest(w, "no login challenge provided")
		return
	}
	challenge := keys[0]

	//fetch information about the request
	cinfo, err := s.client.GetConsentInfo(challenge)
	if err != nil {
		log.Error("Error getting consent info", "error", err)
		s.httpInternalError(w, err)
		return
	}
	consent := cinfo.Skip

	if r.Method == http.MethodGet && !cinfo.Skip {
		err := s.templateConsent.Execute(w, map[string]interface{}{})
		if err != nil {
			s.httpInternalError(w, err)
		}
		return
	}
	if r.Method == http.MethodPost {
		//TODO: check, whether the user gave consent and user should give consent if not...
		consent = true
	}

	if consent {
		requestBody := AcceptConsentRequest{GrantScope: cinfo.RequestedScope, GrantAccessTokenAudience: cinfo.RequestedAudience, Remember: true, RememberFor: 300}
		conRes, err := s.client.AcceptConsent(keys[0], requestBody)
		if err != nil {
			log.Error("Error giving consent", "error", err)
			s.httpInternalError(w, err)
			return
		}
		http.Redirect(w, r, conRes.RedirectTo, http.StatusFound)
		return
	}
	s.httpUnauthorized(w)
}

func (s Server) GetUser(w http.ResponseWriter, r *http.Request) {
	log.Debugf("%s, %q", r.Method, html.EscapeString(r.URL.Path))
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

func (s Server) EditUser(w http.ResponseWriter, r *http.Request){
	log.Debugf("%s, %q", r.Method, html.EscapeString(r.URL.Path))
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	id, ok := vars["uid"]
	if !ok {
		log.Error("EditUser request without ID.")
		s.httpBadRequest(w, "missing user id")
		return
	}
	firstName, ok := vars["firstName"]
	if !ok {
		log.Error("EditUser request without firstName.")
		s.httpBadRequest(w, "missing firstName")
		return
	}
	lastName, ok := vars["lastName"]
	if !ok {
		log.Error("EditUser request without lastName.")
		s.httpBadRequest(w, "missing lastName")
		return
	}
	email, ok := vars["email"]
	if !ok {
		log.Error("EditUser request without email.")
		s.httpBadRequest(w, "missing email")
		return
	}

	user := new(User)
	user.UserID = id
	user.Email = email
	user.LastName = lastName
	user.FirstName = firstName

	err := s.db.EditUser(ctx, *user)

	//print errors
	if err != nil {
		if err == sql.ErrNoRows {
			log.WithField("user-id", user.UserID).Warn("user not found")
			s.httpNotFound(w)
			return
		}
		log.WithError(err).WithField("user-id", user.UserID).Error("Fialed to edit user.")
		s.httpInternalError(w, fmt.Errorf("failed to edit user"))
		return
	}

	w.Header().Set("content-type", "application/json")
	err = json.NewEncoder(w).Encode(true)
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
