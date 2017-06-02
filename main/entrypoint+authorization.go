package main

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"github.com/kudinovdenis/logger"
	"net/url"
	"github.com/dgrijalva/jwt-go"
)

func authorizationHandler(rw http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")

	if domain == "" {
		http.Error(rw, "Must provide domain name.", http.StatusBadRequest)
		return
	}

	redirectURL := "https://accounts.google.com/o/oauth2/auth?client_id=" +
		CLIENT_ID +
		"&response_type=code&scope=openid%20email&redirect_uri=" +
		REDIRECT_URL +
		"&openid.realm=" +
		REDIRECT_URL + "&domain=" + domain

	http.Redirect(rw, r, redirectURL, http.StatusMovedPermanently)
}

func oauth2CallbackHandler(rw http.ResponseWriter, r *http.Request) {
	logger.LogRequestToService(r, true)
	logger.Logf(logger.LogLevelDefault, "Received OAuth2 callback with request: %#v.", r.RequestURI)

	u, err := url.ParseRequestURI(r.RequestURI)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	code := u.Query().Get("code")

	/// Verify code and get token

	req, err := http.NewRequest("POST", GOOGLE_CHECK_OAUTH_TOKEN_URL, nil)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	req.Form.Set("code", code)
	req.Form.Set("grant_type", "authorization_code")
	req.Form.Set("client_id", CLIENT_ID)
	req.Form.Set("client_secret", CLIENT_SECRET)
	req.Form.Set("redirect_uri", REDIRECT_URL)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	logger.LogRequestFromService(req, true)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	logger.LogResponseToService(res, true)

	/// Get User email from jwt

	type CheckResult struct {
		AccessToken string `json:"access_token"`
		ExpiresIn int `json:"expires_in"`
		IDToken string `json:"id_token"`
		TokenType string `json:"token_type"`
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	token := CheckResult{}
	json.Unmarshal(b, &token)

	parsed_token, err := jwt.Parse(token.IDToken, checkToken)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}


	claims := parsed_token.Claims.(jwt.MapClaims)
	admin_email := claims["email"].(string)
	domain := claims["hd"].(string)

	logger.Logf(logger.LogLevelDefault, "Parsed from JWT: %s, %s", admin_email, domain)


	redirect_url := SERVER_URL + "/client?domain=" + domain + "&admin_email=" + admin_email
	http.Redirect(rw, r, redirect_url, http.StatusMovedPermanently)
}

func checkToken(t *jwt.Token) (interface{}, error) {
	return t, nil
}