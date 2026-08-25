package conf

import (
	"fmt"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	EnvLocal      = "local"
	EnvProd       = "prod"
	EnvPseudoProd = "pseudo_prod"
)

type ZoomConfig struct {
	ClientID     string `env:"ZOOM_CLIENT_ID"`
	ClientSecret string `env:"ZOOM_CLIENT_SECRET"`
	RedirectURI  string `env:"ZOOM_REDIRECT_URI"`
	AuthorizeURL string `env:"ZOOM_AUTHORIZE_URL"`
}

type MeetingConfig struct {
	Service string     `env:"MEETING_SERVICE" env-default:"zoom"`
	Zoom    ZoomConfig
}

type Config struct {
	BasePath          string
	AppEnv            string `env:"APP_ENV" env-required:""`
	Port              int    `env:"PORT" env-default:"8080"`
	SuperuserUsername string `env:"SUPERUSER_USERNAME" env-required:""`
	SuperuserPassword string `env:"SUPERUSER_PASSWORD" env-required:""`
	Secret            string `env:"SECRET" env-required:""`
	Meeting           MeetingConfig
}

var cfg *Config
var once sync.Once

func Conf() *Config {
	once.Do(func() {
		var co Config
		if err := cleanenv.ReadConfig(".env", &co); err != nil {
			fmt.Printf("Error reading config: %v\n", err)
		}
		switch co.AppEnv {
		case EnvLocal, EnvProd, EnvPseudoProd:
		default:
			panic(fmt.Sprintf("Invalid APP_ENV. Got '%s'", co.AppEnv))
		}
		if co.Port < 1 || co.Port > 65535 {
			panic(fmt.Sprintf("Invalid PORT. Got '%d'", co.Port))
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
	return c.AppEnv == EnvProd
}

func (c *Config) IsLocal() bool {
	return c.AppEnv == EnvLocal
}
