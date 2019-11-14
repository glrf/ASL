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

var hydraAdminURL = flag.String("admin-url", "https://localhost:9001", "URL of the hydra admin api")
var listen = flag.String("listen", ":8088", "on what url to start the server on")
var dsn = flag.String("dsn", "", "DSN of the DB to connect to: user:password@/dbname")
var vaultURL = flag.String("vault-url", "https://vault.fadalax.tech:8200","URL of the Vault instance")

type server struct {
	router          *mux.Router
	hydra           hydraAdminClient
	db              storageClient
	vault           vaultClient
	templateLogin   *template.Template
	templateConsent *template.Template
}

type hydraAdminClient interface {
	GetLoginInfo(challenge string) (LoginInfo, error)
	AcceptLogin(challenge string, req AcceptLoginRequest) (AcceptLoginResponse, error)
	GetConsentInfo(challenge string) (ConsentInfo, error)
	AcceptConsent(challenge string, req AcceptConsentRequest) (AcceptConsentResponse, error)
}

type storageClient interface {
	GetUser(ctx context.Context, userID string) (User, error)
	ChangePassword(ctx context.Context, userID string, password string) error
	Login(ctx context.Context, userID string, password string) bool
}

type vaultClient interface {
	PKIRoleExists(role string) (bool, error)
	CreatePKIUser(name string) error
}

func main() {
	log.SetLevel(log.TraceLevel) // log all the things
	flag.Parse()
	// Setting up client to communicate with hydra
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	hydra := HydraClient{client: &http.Client{Transport: tr}, adminUrl: *hydraAdminURL}
	if *dsn == "" {
		log.Error("Empty DSN passed.")
	}
	db, err := NewStorage(*dsn)
	if err != nil {
		log.WithError(err).Fatal("Failed to create storage component.")
	}
	// Reads token from VAULT_TOKEN automatically.
	vc, err := NewVaultClient(*vaultURL)
	if err != nil {
		log.WithError(err).Fatal("Failed to create vault client.")
	}
	
	// Prepare HTTP server
	r := mux.NewRouter()
	ser := server{hydra: &hydra, router: r, db: db, vault: vc}

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

func (s server) Login(w http.ResponseWriter, r *http.Request) {
	l := log.WithContext(r.Context())
	l.Debugf("%s, %q", r.Method, html.EscapeString(r.URL.Path))
	//keys[0] contains the challenge
	keys, ok := r.URL.Query()["login_challenge"]
	if !ok {
		l.Info("No login challenge provided")
		s.httpBadRequest(w, "no login challenge provided")
		return
	}
	info, err := s.hydra.GetLoginInfo(keys[0])
	if err != nil {
		l.Error("Error getting login info", "error", err)
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
		err = r.ParseForm()
		if err != nil {
			s.httpBadRequest(w, "invalid form")
			return
		}
		username = r.FormValue("username")
		password := r.FormValue("password")
		l = l.WithField("username", username)
		authenticated = s.db.Login(r.Context(), username, password)
		l.Info("Login Attempt.")
	}

	// Accept login request
	if authenticated {
		l.Info("Authenticated")
		acceptBody := AcceptLoginRequest{Subject: username, Remember: false, RememberFor: 300}
		accRes, err := s.hydra.AcceptLogin(keys[0], acceptBody)
		if err != nil {
			l.WithError(err).Error("Error accepting login.")
			s.httpInternalError(w, err)
			return
		}

		exists, err := s.vault.PKIRoleExists(username)
		if err != nil {
			l.WithError(err).Error("Failed to check whether a PKI role exists.")
			s.httpInternalError(w, err) // TODO(bimmlerd) do we leak too much information here?
			return
		}
		
		if !exists {
			err := s.vault.CreatePKIUser(username)
			if err != nil {
				l.WithError(err).Error("Failed to create PKI User.")
				s.httpInternalError(w, err) // TODO(bimmlerd) do we leak too much information here?
				return
			}
		}

		// redirect
		http.Redirect(w, r, accRes.RedirectTo, http.StatusFound)
		return
	}
	s.httpUnauthorized(w)
}

func (s server) Consent(w http.ResponseWriter, r *http.Request) {
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
	cinfo, err := s.hydra.GetConsentInfo(challenge)
	if err != nil {
		log.WithError(err).Error("Error getting consent info")
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
		conRes, err := s.hydra.AcceptConsent(keys[0], requestBody)
		if err != nil {
			log.WithError(err).Error("Error giving consent.")
			s.httpInternalError(w, err)
			return
		}
		http.Redirect(w, r, conRes.RedirectTo, http.StatusFound)
		return
	}
	s.httpUnauthorized(w)
}

func (s server) GetUser(w http.ResponseWriter, r *http.Request) {
	log.Debugf("%s, %q", r.Method, html.EscapeString(r.URL.Path))
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
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

func (s server) httpInternalError(w http.ResponseWriter, e error) {
	if e != nil {
		log.Errorf("Error: %v\n", e)
	}
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (s server) httpNotFound(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func (s server) httpUnauthorized(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
}

func (s server) httpBadRequest(w http.ResponseWriter, error string) {
	http.Error(w, error, http.StatusBadRequest)
}
