package conf

import (
	"fmt"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	BasePath          string
	AppEnv            string `env:"APP_ENV" env-required:""`
	SuperuserUsername string `env:"SUPERUSER_USERNAME" env-required:""`
	SuperuserPassword string `env:"SUPERUSER_PASSWORD" env-required:""`
	Secret            string `env:"SECRET" env-required:""`
}

var cfg *Config
var once sync.Once

func Conf() *Config {
	once.Do(func() {
		var co Config
		if err := cleanenv.ReadConfig(".env", &co); err != nil {
			fmt.Printf("Error reading config: %v\n", err)
		}
		if co.AppEnv != "local" && co.AppEnv != "prod" {
			panic(fmt.Sprintf("Invalid APP_ENV. Got '%s'", co.AppEnv))
		}
		if co.Secret == "" {
			panic("SECRET is required")
		}
		if co.SuperuserUsername == "" || co.SuperuserPassword == "" {
			panic("SUPERUSER_USERNAME and SUPERUSER_PASSWORD are required")
		}

		cfg = &co
	})
	return cfg
}

func (c *Config) IsProd() bool {
	return c.AppEnv == "prod"
}
