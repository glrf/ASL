package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
)

type AcceptRequest struct {
	Subject     string `json:"subject"`
	Remember     bool `json:"remember"`
	RememberFor int `json:"remember_for"`
}

type AcceptResponse struct {
	RedirectTo string `json:"redirect_to"`
}

func main() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request){
		log.Printf("%s, %q", r.Method, html.EscapeString(r.URL.Path))
		//fmt.Fprintf(w, "ok")
		keys, ok := r.URL.Query()["login_challenge"]
		if ok {
			log.Printf("Challenge: %s", keys[0])

			//fetch information about the request
			request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://localhost:9001/oauth2/auth/requests/login?login_challenge=%s", keys[0]), nil)
			res, err := client.Do(request)
			if err != nil{
				fmt.Fprintf(w, "Not ok")
			}
			bu, _ := ioutil.ReadAll(res.Body)
			fmt.Println(string(bu))

			//redirect the user
			requestBody := AcceptRequest{Subject: "sub", Remember:false, RememberFor: 300}
			buf, err := json.Marshal(requestBody)
			if(err != nil) {
				fmt.Fprintf(w, "errorrrr")
			}

			url := fmt.Sprintf("https://localhost:9001/oauth2/auth/requests/login/accept?login_challenge=%s", keys[0])
			fmt.Println(url)
			request, err = http.NewRequest(http.MethodPut, url, bytes.NewReader(buf))

			if err != nil{
				fmt.Fprintf(w, "not ok")
			}
			res, err = client.Do(request)
			if err != nil{
				fmt.Fprintf(w, "Not ok")
			}
			bu, _ = ioutil.ReadAll(res.Body)
			fmt.Println(string(bu))

			//redirect the user
			resonseRedirect := AcceptResponse{}
			err = json.Unmarshal(bu, &resonseRedirect)
			http.Redirect(w, r, resonseRedirect.RedirectTo, 302)
		}
		//first fetch information about the request
		//res1, err1 := client.Get(fmt.Sprintf("https://hydra/oauth2/auth/requests/login?&login_challenge=%s", keys[0]))
		//https://hydra/oauth2/auth/requests/login?
		//get JWT back

		//here the user needs to login.

		//tell hydra, that the user got logged in. and tell information about the user.

		//res, err := client.Put(fmt.Sprintf("https://localhost:9000/oauth2/auth/requests/login/accept?client_id=special_laura&scope=openid&response_type=id_token&state=AAAAAAAA&redirect_uri=http://127.0.0.1:9010/callback&login_verifier=%s", keys[0]))
		//https://localhost:9000/oauth2/auth?client_id=special_laura&response_type=token&state=111111111111111111&scope=openid
		//http://hydra/oauth2/auth?client_id=...&...&login_verifier=4321

	})
	http.HandleFunc("/consent", func(w http.ResponseWriter, r *http.Request){
		//do Consent!
	})
	log.Fatal(http.ListenAndServe(":8088", nil))
}