package configurer

import (
	"strings"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Configurations struct {
	Server   ServerConfigurations   `koanf:"server"`
	Database DatabaseConfigurations `koanf:"database"`
	Iam      IamConfigurations      `koanf:"iam"`
	Email    EmailConfigurations    `koanf:"email"`
}

type ServerConfigurations struct {
	Port string `koanf:"port"`
}

type DatabaseConfigurations struct {
	Enabled    bool   `koanf:"enabled"`
	Host       string `koanf:"host"`
	Port       int    `koanf:"port"`
	DbName     string `koanf:"db-name"`
	Username   string `koanf:"user"`
	Password   string `koanf:"password"` // nolint:gosec
	PoolSize   int    `koanf:"pool-size"`
	LogQueries bool   `koanf:"log-queries"`
}

type IamConfigurations struct {
	BaseURL    string `koanf:"base-url"`
	HmacSecret string `koanf:"hmac-secret"`
}

type EmailConfigurations struct {
	Enabled bool   `koanf:"enabled"`
	From    string `koanf:"from"`
	ApiKey  string `koanf:"api-key"`
}

// LoadConfigurations Loads configurations depending upon the environment
func LoadConfigurations(path string) (*Configurations, error) {
	k := koanf.New(".")
	err := k.Load(file.Provider(path), toml.Parser())
	if err != nil {
		return nil, err
	}

	// Searches for env variables and will transform them into koanf format
	// e.g. SERVER_PORT variable will be server.port: value
	err = k.Load(env.Provider("", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(s), "_", ".")
	}), nil)
	if err != nil {
		return nil, err
	}

	var configuration Configurations

	err = k.Unmarshal("", &configuration)
	if err != nil {
		return nil, err
	}

	return &configuration, nil
}
