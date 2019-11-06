package main

import (
	"crypto/tls"
	"fmt"
	"html"
	"log"
	"net/http"
)

func main() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		log.Printf("%s, %q", r.Method, html.EscapeString(r.URL.Path))
		fmt.Fprintf(w, "ok")
		keys, ok := r.URL.Query()["login_challenge"]
		if ok {
			log.Printf("Challenge: %s", keys[0])
		}
		//first fetch information about the request
		//res1, err1 := client.Get(fmt.Sprintf("https://hydra/oauth2/auth/requests/login?&login_challenge=%s", keys[0]))
		//https://hydra/oauth2/auth/requests/login?
		//get JWT back

		//here the user needs to login.

		//tell hydra, that the user got logged in. and tell information about the user.
		res, err := client.Get(fmt.Sprintf("https://localhost:9000/oauth2/auth/requests/login/accept?client_id=special_laura&scope=openid&response_type=id_token&state=AAAAAAAA&redirect_uri=http://127.0.0.1:9010/callback&login_verifier=%s", keys[0]))
		//https://localhost:9000/oauth2/auth?client_id=special_laura&response_type=token&state=111111111111111111&scope=openid
		//http://hydra/oauth2/auth?client_id=...&...&login_verifier=4321
		if err != nil {
			fmt.Println(err)
		} else{
			fmt.Println(res)
		}
	})
	log.Fatal(http.ListenAndServe(":8088", nil))
}