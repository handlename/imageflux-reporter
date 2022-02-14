package reporter

import (
	"fmt"

	"github.com/bloom42/rz-go"
	"github.com/bloom42/rz-go/log"
	"github.com/kayac/go-config"
)

type Config struct {
	Email    string   `yaml:"email,omitempty"`
	Password string   `yaml:"password,omitempty"`
	Origins  []Origin `yaml:"origins,omitempty"`
}

type Origin struct {
	Id       int    `yaml:"id,omitempty"`
	Project  string `yaml:"project,omitempty"`
	Endpoint string `yaml:"endpoint,omitempty"`
}

func LoadConfig(path string) (*Config, error) {
	c := &Config{}

	if err := config.LoadWithEnv(c, path); err != nil {
		log.Error("failed to load config", rz.String("path", path))
		return nil, err
	}
	log.Debug("config loaded", rz.String("path", path))
	log.Debug(fmt.Sprintf("%+v", c))

	return c, nil
}
