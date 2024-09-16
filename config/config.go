package config

import (
	"errors"
	"github.com/betterde/cdns/internal/journal"
	"github.com/spf13/viper"
	"os"
	"strings"
)

const TLSModeACME = "acme"
const TLSModeFile = "file"

var Conf *Config

type Config struct {
	NS        NS        `yaml:"ns" mapstructure:"NS"`
	SOA       SOA       `yaml:"soa" mapstructure:"SOA"`
	DNS       DNS       `yaml:"dns" mapstructure:"DNS"`
	HTTP      HTTP      `yaml:"http" mapstructure:"HTTP"`
	Ingress   Ingress   `yaml:"ingress" mapstructure:"INGRESS"`
	Logging   Logging   `yaml:"logging" mapstructure:"LOGGING"`
	Providers Providers `yaml:"providers" mapstructure:"PROVIDERS"`
}

type NS struct {
	IP string `yaml:"ip" mapstructure:"IP"`
}

type Logging struct {
	Level string `yaml:"level" mapstructure:"LEVEL"`
}

type DNS struct {
	Admin    string            `yaml:"admin" mapstructure:"ADMIN"`
	Listen   string            `yaml:"listen" mapstructure:"LISTEN"`
	NSName   string            `yaml:"nsname" mapstructure:"NSNAME"`
	Records  map[string]Record `yaml:"records" mapstructure:"RECORDS"`
	Protocol string            `yaml:"protocol" mapstructure:"PROTOCOL"`
}

type HTTP struct {
	TLS    TLS    `yaml:"tls" mapstructure:"TLS"`
	Domain string `yaml:"domain" mapstructure:"DOMAIN"`
	Listen string `yaml:"listen" mapstructure:"LISTEN"`
}

type Record struct {
	Type  string `yaml:"type" mapstructure:"TYPE"`
	Value string `yaml:"value" mapstructure:"VALUE"`
}

type Ingress struct {
	IP string `yaml:"ip" mapstructure:"IP"`
}

type Providers struct {
	ACME ACME `yaml:"acme" mapstructure:"ACME"`
	File File `yaml:"file" mapstructure:"FILE"`
}

type TLS struct {
	Mode string `yaml:"mode" mapstructure:"MODE"`
}

type SOA struct {
	Domain string `yaml:"domain" mapstructure:"DOMAIN"`
}

type ACME struct {
	Email   string `yaml:"email" mapstructure:"EMAIL"`
	Server  string `yaml:"server" mapstructure:"SERVER"`
	Storage string `yaml:"storage" mapstructure:"STORAGE"`
}

type File struct {
	TLSKey  string `yaml:"tlsKey" mapstructure:"TLSKEY"`
	TLSCert string `yaml:"tlsCert" mapstructure:"TLSCERT"`
}

func Parse(file string) {
	if file != "" {
		viper.SetConfigFile(file)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".cdns")
		viper.AddConfigPath("/etc/cdns")
	}

	// read in environment variables that match
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("CDNS")

	var notFoundError viper.ConfigFileNotFoundError

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil && errors.As(err, &notFoundError) {
		viper.SetDefault("DNS.LISTEN", "0.0.0.0:53")
		viper.SetDefault("DNS.PROTOCOL", "both")
		viper.SetDefault("HTTP.LISTEN", "0.0.0.0:443")
		viper.SetDefault("LOGGING.LEVEL", "DEBUG")

		err = viper.BindEnv("NS.IP", "CDNS_NS_IP")
		if err != nil {
			journal.Logger.Sugar().Error(err)
		}

		err = viper.BindEnv("DNS.ADMIN", "CDNS_DNS_ADMIN")
		if err != nil {
			journal.Logger.Sugar().Error(err)
		}

		err = viper.BindEnv("DNS.NSNAME", "CDNS_DNS_NSNAME")
		if err != nil {
			journal.Logger.Sugar().Error(err)
		}

		err = viper.BindEnv("DNS.LISTEN", "CDNS_DNS_LISTEN")
		if err != nil {
			journal.Logger.Sugar().Error(err)
		}

		err = viper.BindEnv("DNS.PROTOCOL", "CDNS_DNS_PROTOCOL")
		if err != nil {
			journal.Logger.Sugar().Error(err)
		}

		err = viper.BindEnv("SOA.DOMAIN", "CDNS_SOA_DOMAIN")
		if err != nil {
			journal.Logger.Sugar().Error(err)
		}

		err = viper.BindEnv("HTTP.TLS.MODE", "CDNS_HTTP_TLS_MODE")
		if err != nil {
			journal.Logger.Sugar().Error(err)
		}

		err = viper.BindEnv("HTTP.DOMAIN", "CDNS_HTTP_DOMAIN")
		if err != nil {
			journal.Logger.Sugar().Error(err)
		}

		err = viper.BindEnv("HTTP.LISTEN", "CDNS_HTTP_LISTEN")
		if err != nil {
			journal.Logger.Sugar().Error(err)
		}

		err = viper.BindEnv("INGRESS.IP", "CDNS_INGRESS_IP")
		if err != nil {
			journal.Logger.Sugar().Error(err)
		}

		err = viper.BindEnv("LOGGING.LEVEL", "CDNS_LOGGING_LEVEL")
		if err != nil {
			journal.Logger.Sugar().Error(err)
		}

		err = viper.BindEnv("PROVIDERS.ACME.EMAIL", "CDNS_PROVIDERS_ACME_EMAIL")
		if err != nil {
			journal.Logger.Sugar().Error(err)
		}

		err = viper.BindEnv("PROVIDERS.ACME.SERVER", "CDNS_PROVIDERS_ACME_SERVER")
		if err != nil {
			journal.Logger.Sugar().Error(err)
		}

		err = viper.BindEnv("PROVIDERS.ACME.STORAGE", "CDNS_PROVIDERS_ACME_STORAGE")
		if err != nil {
			journal.Logger.Sugar().Error(err)
		}

		err = viper.BindEnv("PROVIDERS.FILE.TLSKEY", "CDNS_PROVIDERS_FILE_TLSKEY")
		if err != nil {
			journal.Logger.Sugar().Error(err)
		}

		err = viper.BindEnv("PROVIDERS.FILE.TLSCERT", "CDNS_PROVIDERS_FILE_TLSCERT")
		if err != nil {
			journal.Logger.Sugar().Error(err)
		}
	}

	// read in environment variables that match
	viper.AutomaticEnv()

	err := viper.Unmarshal(&Conf)
	if err != nil {
		journal.Logger.Sugar().Errorf("Unable to decode into config struct, %v", err)
		os.Exit(1)
	}
}
