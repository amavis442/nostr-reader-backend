package config

import (
	"amavis442/nostr-reader/internal/db"
	"amavis442/nostr-reader/internal/http"
	wrapper "amavis442/nostr-reader/internal/nostr"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

/**
 * Used to store the config.json file and some database related stuff for easy access
 *
 */
type Config struct {
	Database *db.DbConfig
	Server   *http.ServerConfig
	Env      string
	Interval uint
	Nostr    *wrapper.WrapperConfig
}

func configDir() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		dir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(dir, ".config"), nil
	default:
		return os.UserConfigDir()
	}
}

/**
 * Get the content of config.json file
 */
func LoadConfig() (*Config, error) {
	var cfg Config

	dir, err := configDir()
	if err != nil {
		return nil, err
	}
	dir = filepath.Join(dir, "nostr-reader")
	fp := filepath.Join(dir, "config.json")
	os.MkdirAll(filepath.Dir(fp), 0700)

	content, err := os.ReadFile(fp)
	if err != nil {
		fmt.Println("Done", err)
		slog.Error("Error when opening file", "error", err.Error())
		os.Exit(1)
	}

	err = json.Unmarshal(content, &cfg)
	if err != nil {
		slog.Error("Error during Unmarshal()", "error", err.Error())
		os.Exit(1)
	}

	if cfg.Nostr.PrivateKey == "" {
		slog.Info("You need to add your private key. This key will never be transmitted and stays local")
		os.Exit(0)
	}

	var pubKey string

	if cfg.Nostr.PrivateKey[:4] == "nsec" {
		if _, s, err := nip19.Decode(cfg.Nostr.PrivateKey); err == nil {
			if pubKey, err = nostr.GetPublicKey(s.(string)); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {

		pubKey, _ = nostr.GetPublicKey(cfg.Nostr.PrivateKey)
	}

	cfg.Nostr.PubKey = pubKey
	cfg.Nostr.Nsec, _ = nip19.EncodePrivateKey(cfg.Nostr.PrivateKey)
	cfg.Nostr.Npub, _ = nip19.EncodePublicKey(pubKey)

	return &cfg, nil
}
