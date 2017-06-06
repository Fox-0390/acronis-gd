package main

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"github.com/kudinovdenis/logger"
	"net/url"
	"github.com/dgrijalva/jwt-go"
	"bytes"
	"fmt"
	"github.com/kudinovdenis/acronis-gd/config"
)

func authorizationHandler(rw http.ResponseWriter, r *http.Request) {
	logger.LogRequestToService(r, true)

	domain := r.URL.Query().Get("domain")

	if domain == "" {
		http.Error(rw, "Must provide domain name.", http.StatusBadRequest)
		return
	}

	redirectURL := "https://accounts.google.com/o/oauth2/auth?client_id=" +
		config.Cfg.ClientID +
		"&response_type=code&scope=openid%20email&redirect_uri=" +
		config.Cfg.OauthCallbackURL() +
		"&openid.realm=" +
		config.Cfg.OauthCallbackURL() + "&domain=" + domain

	logger.Logf(logger.LogLevelDefault, "Redirecting to %s", redirectURL)

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

	data := url.Values{}
	data.Set("code", code)
	data.Add("grant_type", "authorization_code")
	data.Add("client_id", config.Cfg.ClientID)
	data.Add("client_secret", config.Cfg.ClientSecret)
	data.Add("redirect_uri", config.Cfg.OauthCallbackURL())

	/// Verify code and get token

	req, err := http.NewRequest("POST", config.Cfg.GoogleCheckOauth2TokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

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

	tokenString := token.IDToken

	parsed_token, err := jwt.Parse(tokenString, nil)

	claims, ok := parsed_token.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Println(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return

	}

	logger.Logf(logger.LogLevelDefault, "CLAIMS! : %#v", claims)

	admin_email := claims["email"].(string)
	domain := claims["hd"].(string)

	logger.Logf(logger.LogLevelDefault, "Parsed from JWT: %s, %s", admin_email, domain)

	redirect_url := config.Cfg.ServerURL + "/client?domain=" + domain + "&admin_email=" + admin_email

	logger.Logf(logger.LogLevelDefault, "Redirecting to %s", redirect_url)

	http.Redirect(rw, r, redirect_url, http.StatusMovedPermanently)
}