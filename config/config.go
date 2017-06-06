package config

import (
	"github.com/kudinovdenis/acronis-gd/utils"
	"encoding/json"
)

type Config struct {
	GoogleCheckOauth2TokenURL string `json:"google_check_oauth2_token_url"`
	ServerURL string `json:"server_url"`
	ClientID string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Oauth2CallbackPath string `json:"oauth2_callback_path"`
	UseLocalServer bool `json:"use_local_server"`
	Port string `json:"port"`
}

var Cfg *Config = &Config{}

func (config *Config) OauthCallbackURL() string {
	return config.ServerURL + config.Oauth2CallbackPath
}

func PopulateConfigWithFile(path string) error {
	reader, err := utils.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.NewDecoder(reader).Decode(Cfg)
	if err != nil {
		return err
	}

	return nil
}