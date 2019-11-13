package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type HydraClient struct {
	client   *http.Client
	adminUrl string
}

type LoginInfo struct {
}

func (c HydraClient) GetLoginInfo(challenge string) (LoginInfo, error) {
	//fetch information about the request
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/oauth2/auth/requests/login?login_challenge=%s", c.adminUrl, challenge), nil)
	if err != nil {
		return LoginInfo{}, err
	}
	_, err = c.client.Do(request)
	if err != nil {
		return LoginInfo{}, err
	}
	return LoginInfo{}, nil
}

type AcceptLoginRequest struct {
	Subject     string `json:"subject"`
	Remember    bool   `json:"remember"`
	RememberFor int    `json:"remember_for"`
}

type AcceptLoginResponse struct {
	RedirectTo string `json:"redirect_to"`
}

func (c HydraClient) AcceptLogin(challenge string, req AcceptLoginRequest) (AcceptLoginResponse, error) {
	buf, err := json.Marshal(req)
	if err != nil {
		return AcceptLoginResponse{}, err
	}
	url := fmt.Sprintf("%s/oauth2/auth/requests/login/accept?login_challenge=%s", c.adminUrl, challenge)
	request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(buf))
	if err != nil {
		return AcceptLoginResponse{}, err
	}
	res, err := c.client.Do(request)
	if err != nil {
		return AcceptLoginResponse{}, err
	}
	buf, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return AcceptLoginResponse{}, err
	}
	accRes := AcceptLoginResponse{}
	err = json.Unmarshal(buf, &accRes)
	if err != nil {
		return AcceptLoginResponse{}, err
	}
	return accRes, nil
}

type ConsentInfo struct {
	Skip              bool     `json:"skip"`
	Subject           string   `json:"subject"`
	RequestedScope    []string `json:"requested_scope"`
	RequestedAudience []string `json:"requested_access_token_audience"`
}

func (c HydraClient) GetConsentInfo(challenge string) (ConsentInfo, error) {
	//fetch information about the request
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/oauth2/auth/requests/consent?consent_challenge=%s", c.adminUrl, challenge), nil)
	if err != nil {
		return ConsentInfo{}, err
	}
	res, err := c.client.Do(request)
	if err != nil {
		return ConsentInfo{}, err
	}
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ConsentInfo{}, err
	}
	conInfo := ConsentInfo{}
	err = json.Unmarshal(buf, &conInfo)
	if err != nil {
		return ConsentInfo{}, err
	}
	return conInfo, nil
}

type AcceptConsentRequest struct {
	GrantScope               []string `json:"grant_scope"`
	GrantAccessTokenAudience []string `json:"grant_access_token_audience"`
	Remember                 bool     `json:"remember"`
	RememberFor              int      `json:"remember_for"`
	//TODO: evt. add session
}
type AcceptConsentResponse struct {
	RedirectTo string `json:"redirect_to"`
}

func (c HydraClient) AcceptConsent(challenge string, req AcceptConsentRequest) (AcceptConsentResponse, error) {
	buf, err := json.Marshal(req)
	if err != nil {
		return AcceptConsentResponse{}, err
	}

	url := fmt.Sprintf("%s/oauth2/auth/requests/consent/accept?consent_challenge=%s", c.adminUrl, challenge)
	request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(buf))
	if err != nil {
		return AcceptConsentResponse{}, err
	}
	res, err := c.client.Do(request)
	if err != nil {
		return AcceptConsentResponse{}, err
	}
	buf, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return AcceptConsentResponse{}, err
	}
	//redirect
	responseRedirect := AcceptConsentResponse{}
	err = json.Unmarshal(buf, &responseRedirect)
	if err != nil {
		return AcceptConsentResponse{}, err
	}
	return responseRedirect, nil
}
