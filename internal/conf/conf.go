package conf

import (
	"fmt"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	AppEnv        string `env:"APP_ENV" env-required:""`
	AdminUsername string `env:"ADMIN_USERNAME" env-required:""`
	AdminPassword string `env:"ADMIN_PASSWORD" env-required:""`
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
		cfg = &co
	})
	return cfg
}

func (c *Config) IsProd() bool {
	return c.AppEnv == "prod"
}
