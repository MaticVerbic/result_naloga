package config

import (
	"time"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config ...
type Config struct {
	Log  *logrus.Entry
	URLS []string
}

// New returns a new Config struct
func New() *Config {
	if err := godotenv.Load(); err != nil {
		panic(errors.Wrap(err, "failed to load env with error"))
	}

	viper.AutomaticEnv()

	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FORCE_COLORS", true)
	viper.SetDefault("LOG_FULL_TIMESTAMP", true)

	colors := viper.GetBool("LOG_FORCE_COLORS")
	level, err := logrus.ParseLevel(viper.GetString("LOG_LEVEL"))
	if err != nil {
		panic(errors.Wrap(err, "failed to parse log level"))
	}

	log := logger(level, viper.GetBool("LOG_FORCE_COLORS"), colors, !colors)
	urls := viper.GetStringSlice("URLS")

	return &Config{
		Log:  log,
		URLS: urls,
	}
}

func logger(logLevel logrus.Level, timestamp, forceColors, disableColors bool) *logrus.Entry {
	logrus.SetLevel(logLevel)
	formatter := &logrus.TextFormatter{
		TimestampFormat: "02-01-2006T15:04:05",
		FullTimestamp:   timestamp,
		ForceColors:     forceColors,
		DisableColors:   disableColors,
	}
	logrus.SetFormatter(formatter)
	return logrus.WithField("start", time.Now().Format("02-01-2006T15:04:05"))
}
