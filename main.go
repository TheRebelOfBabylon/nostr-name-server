package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

var (
	pathToCfg = flag.String("config", "config.json", "path to the JSON configuration file")
)

type Response struct {
	Names  map[string]string   `json:"names"`
	Relays map[string][]string `json:"relays,omitempty"`
}

type Names struct {
	Pubkey string   `json:"pubkey"`
	Relays []string `json:"relays,omitempty"`
}

type Config struct {
	Names map[string]*Names `json:"names"`
	Port  int               `json:"port"`
}

type AuthHandler struct {
	cfg *Config
}

func (a *AuthHandler) AuthUser(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid user"))
		return
	}
	userInfo, ok := a.cfg.Names[name]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid user"))
		return
	}
	resp := Response{
		Names: map[string]string{
			name: userInfo.Pubkey,
		},
	}
	if userInfo.Relays != nil && len(userInfo.Relays) != 0 {
		resp.Relays = map[string][]string{
			userInfo.Pubkey: userInfo.Relays[:],
		}
	}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("failed to JSON marshal response: %v", err)))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(respBytes)
}

func main() {
	flag.Parse()
	// read config file
	cfgFileBytes, err := os.ReadFile(*pathToCfg)
	if err != nil {
		log.Fatalf("failed to read config file at %s: %v", *pathToCfg, err)
	}
	var cfg Config
	if err = json.Unmarshal(cfgFileBytes, &cfg); err != nil {
		log.Fatalf("failed to parse config file: %v", err)
	}
	authHandler := AuthHandler{
		cfg: &cfg,
	}
	http.HandleFunc("/.well-known/nostr.json", authHandler.AuthUser)

	if err := http.ListenAndServe(fmt.Sprintf(":%v", cfg.Port), nil); err != nil {
		log.Fatalf("failed to listen on port %v: %v", cfg.Port, err)
	}
}
