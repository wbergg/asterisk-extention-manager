package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

type Config struct {
	ListenAddr    string `json:"listen_addr"`
	DBPath        string `json:"db_path"`
	JWTSecret     string `json:"jwt_secret"`
	PJSIPConfPath string `json:"pjsip_conf_path"`
	AsteriskCmd   string `json:"asterisk_cmd"`
	AdminUser     string `json:"admin_user"`
	AdminPass     string `json:"admin_pass"`
	CDRLogPath    string `json:"cdr_log_path"`
	TLSDomain     string `json:"tls_domain"`
	TLSCertDir    string `json:"tls_cert_dir"`
	TLSCertFile    string `json:"tls_cert_file"`
	TLSKeyFile     string `json:"tls_key_file"`
	FaxStoragePath  string `json:"fax_storage_path"`
	FaxSpoolPath    string `json:"fax_spool_path"`
	AsteriskLogPath string `json:"asterisk_log_path"`
}

var Loaded *Config

func Load() *Config {
	configPath := flag.String("config", "config/config.json", "path to config file")
	flag.Parse()

	data, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("Failed to read config file %s: %v", *configPath, err)
	}

	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		log.Fatalf("Failed to parse config file %s: %v", *configPath, err)
	}

	Loaded = cfg
	return cfg
}
