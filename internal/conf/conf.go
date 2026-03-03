package conf

import (
	"fmt"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	AppEnv               string `env:"APP_ENV" env-required:""`
	AdminTeacherUsername string `env:"ADMIN_TEACHER_USERNAME" env-required:""`
	AdminTeacherPassword string `env:"ADMIN_TEACHER_PASSWORD" env-required:""`
	SuperuserUsername    string `env:"SUPERUSER_USERNAME" env-required:""`
	SuperuserPassword    string `env:"SUPERUSER_PASSWORD" env-required:""`
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

		if co.AdminTeacherUsername == co.SuperuserUsername {
			panic("admin teacher username and superuser username must be different")
		}

		cfg = &co
	})
	return cfg
}

func (c *Config) IsProd() bool {
	return c.AppEnv == "prod"
}
