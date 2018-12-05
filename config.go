package main

import (
	"encoding/json"
	"io/ioutil"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type gameserverConfig struct {
	RoundID        int    `json:"round_id"`
	Instance       string `json:"instance"`
	Hostname       string `json:"hostname"`
	Production     bool   `json:"production"`
	UserPictureURL string `json:"user_avatar_picture_url"`

	Log struct {
		Level   zapcore.Level `json:"level"`
		Config  string        `json:"config"`
		Colored bool          `json:"colored"`
		File    struct {
			Active bool   `json:"active"`
			Name   string `json:"name"`
		} `json:"file"`
		Sentry struct {
			Active bool   `json:"active"`
			DSN    string `json:"dsn"`
		} `json:"sentry"`
	} `json:"log"`

	Database struct {
		MariaDB struct {
			Host     string `json:"host"`
			User     string `json:"user"`
			Password string `json:"password"`
			Name     string `json:"name"`
		} `json:"mariadb"`
	} `json:"database"`

	Network struct {
		TCP struct {
			Addr string `json:"addr"`
		} `json:"tcp"`
		WS struct {
			Addr        string `json:"addr"`
			ValidOrigin string `json:"valid_origin"`
			TLS         struct {
				Active bool   `json:"active"`
				Cert   string `json:"cert"`
				PKey   string `json:"pkey"`
			} `json:"tls"`
		} `json:"ws"`
	} `json:"network"`

	Battle struct {
		RoundID int
		Users   []int
	} `json:"-"`
}

// loadConfig takes a path to a configfile and returns a
// pointer to a gameserverConfig
func loadConfig(path string) *gameserverConfig {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		logctx.Fatal("failed to load config",
			zap.Error(err),
			zap.String("path", path))
	}

	conf := &gameserverConfig{}
	err = json.Unmarshal([]byte(f), conf)
	if err != nil {
		logctx.Fatal("failed to load config",
			zap.Error(err),
			zap.String("path", path))
	}

	return conf
}
