package config

import (
	"os"

	contact_info "github.com/Newlode/forcepoint-ngfw-licenses/ngfw-licenses/contact-info"
	"github.com/mbndr/logo"
	"github.com/snwfdhmp/errlog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	Logger *logo.Logger

	Loggers []*logo.Logger
)

type Config struct {
	Silent            bool
	Debug             bool                      `mapstructure:"debug"`
	Verbose           bool                      `mapstructure:"verbose"`
	ConcurrentWorkers int                       `mapstructure:"concurrent_workers"`
	LicensesOutputDir string                    `mapstructure:"licenses_output_dir"`
	ContactInfo       *contact_info.ContactInfo `mapstructure:"contact_info"`
	Reseller          string                    `mapstructure:"resseller"`
	Binding           string                    `mapstructure:"binding"`
}

//=================================================================
// Config

var (
	ConfigFile string
	Cfg        = Config{}
)

func init() {
	cobra.OnInitialize(initConfig)
	Loggers = make([]*logo.Logger, 0)
}

func initConfig() {
	if ConfigFile != "" {
		viper.SetConfigFile(ConfigFile)
	} else {
		viper.SetConfigFile("./config.yml")
	}

	readConfig()

	if Cfg.Verbose {
		Logger.SetLevel(logo.INFO)
		SetLogLevel(logo.INFO)
		// ngfwlicenses.Logger.SetLevel(logo.INFO)
	}

	if Cfg.Debug {
		Logger.SetLevel(logo.DEBUG)
		SetLogLevel(logo.DEBUG)
		// ngfwlicenses.Logger.SetLevel(logo.DEBUG)
	}

	if Cfg.ContactInfo != nil {
		if err := Cfg.ContactInfo.Validate(); err != nil {
			Logger.Fatal(err)
		}
	}
}

func readConfig() {
	// viper.SetConfigName("config.yml")
	viper.SetConfigType("yaml")

	viper.SetDefault("contact_info", nil)

	viper.ReadInConfig()

	err := viper.Unmarshal(&Cfg)
	if errlog.Debug(err) {
		Logger.Fatalf("Unable to read config file: %s\n", err)
	}
}

func GetNewLogger(prefix string) *logo.Logger {
	logger := logo.NewSimpleLogger(os.Stderr, logo.WARN, prefix, true)
	Loggers = append(Loggers, logger)
	return logger
}

func SetLogLevel(level logo.Level) {
	for _, logger := range Loggers {
		logger.SetLevel(level)
	}
}
