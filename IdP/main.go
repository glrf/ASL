package main

import (
	"crypto/tls"
	"flag"
	log "github.com/sirupsen/logrus"
	"html"
	"net/http"
)

type adminClient interface {
	GetLoginInfo(challenge string) (LoginInfo, error)
	AcceptLogin(challenge string, req AcceptLoginRequest) (AcceptLoginResponse, error)
	GetConsentInfo(challenge string) (ConsentInfo, error)
	AcceptConsent(challenge string, req AcceptConsentRequest) (AcceptConsentResponse, error)
}

type Server struct {
	mux    *http.ServeMux
	client adminClient
}

var adminURl = flag.String("admin-url", "https://localhost:9001", "url of the hydra admin api")
var listen = flag.String("listen", ":8088", "on what url to start the server on")

func main() {
	flag.Parse()
	// Setting up client to communicate with hydra
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := HydraClient{client: &http.Client{Transport: tr}, adminUrl: *adminURl}

	// Prepare HTTP server
	mux := http.NewServeMux()
	ser := Server{client: &client, mux: mux}
	mux.HandleFunc("/login", ser.Login)
	mux.HandleFunc("/consent", ser.Consent)

	// Run
	log.Fatal(http.ListenAndServe(*listen, mux))
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
		http.Redirect(w, r, accRes.RedirectTo, 302)
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
	http.Redirect(w, r, conRes.RedirectTo, 302)
}

func (s Server) httpInternalError(w http.ResponseWriter, e error) {
	if e != nil {
		log.Printf("Error: %v\n", e)
	}
	http.Error(w, http.StatusText(500), 500)
}

func (s Server) httpUnauthorized(w http.ResponseWriter) {
	http.Error(w, http.StatusText(403), 403)
}

func (s Server) httpBadRequest(w http.ResponseWriter, error string) {
	http.Error(w, error, 400)
}
