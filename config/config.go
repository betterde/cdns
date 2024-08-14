package config

import (
	"github.com/betterde/cdns/internal/journal"
	"github.com/spf13/viper"
	"os"
	"strings"
)

const TLSModeACME = "acme"
const TLSModeFile = "file"

var Conf *Config

type Config struct {
	NS        NS        `yaml:"ns"`
	SOA       SOA       `yaml:"soa"`
	DNS       DNS       `yaml:"dns"`
	HTTP      HTTP      `yaml:"http"`
	Ingress   Ingress   `yaml:"ingress"`
	Logging   Logging   `yaml:"logging"`
	Providers Providers `yaml:"providers"`
}

type NS struct {
	IP string `yaml:"ip"`
}

type Logging struct {
	Level string `yaml:"level"`
}

type DNS struct {
	Admin    string            `yaml:"admin"`
	Listen   string            `yaml:"listen"`
	NSName   string            `yaml:"nsname"`
	Records  map[string]Record `yaml:"records"`
	Protocol string            `yaml:"protocol"`
}

type HTTP struct {
	TLS      TLS    `yaml:"tls"`
	Domain   string `yaml:"domain"`
	Listen   string `yaml:"listen"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Record struct {
	Type  string `yaml:"type"`
	Value string `yaml:"value"`
}

type Ingress struct {
	IP string `yaml:"ip"`
}

type Providers struct {
	ACME ACME `yaml:"acme"`
	File File `yaml:"file"`
}

type TLS struct {
	Mode string `yaml:"mode"`
}

type SOA struct {
	Domain string `yaml:"domain"`
}

type ACME struct {
	Email   string `yaml:"email"`
	Server  string `yaml:"server"`
	Storage string `yaml:"storage"`
}

type File struct {
	TLSKey  string `yaml:"tlsKey"`
	TLSCert string `yaml:"tlsCert"`
}

func Parse(file string) {
	if file != "" {
		viper.SetConfigFile(file)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName(".cdns")
		viper.AddConfigPath("/etc/cdns")
	}

	// read in environment variables that match
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("CDNS")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		journal.Logger.Errorf("Failed to read configuration file: %s", err)
		os.Exit(1)
	}

	// read in environment variables that match
	viper.AutomaticEnv()

	err := viper.Unmarshal(&Conf)
	if err != nil {
		journal.Logger.Errorf("Unable to decode into config struct, %v", err)
		os.Exit(1)
	}
}
