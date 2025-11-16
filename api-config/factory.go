package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type ConfigFactory interface {
	Create(options ...ConfigOption) (*Config, error)
}

type DefaultConfigFactory struct{}

func NewDefaultConfigFactory() ConfigFactory {
	return &DefaultConfigFactory{}
}

func (f *DefaultConfigFactory) Create(options ...ConfigOption) (*Config, error) {
	appliedOptions := DefaultConfigOptions()
	for _, opt := range options {
		opt(&appliedOptions)
	}

	v := viper.New()

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	if appliedOptions.AppEnv != "" {
		v.SetConfigName(fmt.Sprintf("%s.%s", appliedOptions.FileName, appliedOptions.AppEnv))
	} else {
		v.SetConfigName(appliedOptions.FileName)
	}

	for _, path := range appliedOptions.FilePaths {
		v.AddConfigPath(path)
	}
	v.SetConfigType("yaml")

	f.setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	sensitiveKeys := []string{
		"db.username",
		"db.password",
		"db.host",
		"db.port",
		"db.database",
		"db.schema",
		"db.maxconns",
		"db.minconns",
		"db.maxconnlifetime",
		"db.healthcheckperiod",
		"minio.url",
		"minio.accessKey",
		"minio.secretKey",
		"minio.bucketName",
	}

	for _, key := range v.AllKeys() {
		val := v.GetString(key)
		if strings.Contains(val, "$") {
			shouldExpand := true
			for _, sensitiveKey := range sensitiveKeys {
				if strings.EqualFold(key, sensitiveKey) {
					shouldExpand = false
					break
				}
			}
			if shouldExpand {
				expanded := os.ExpandEnv(val)
				v.Set(key, expanded)
			}
		}
	}

	return NewConfig(v), nil
}

func (f *DefaultConfigFactory) setDefaults(v *viper.Viper) {
	v.SetDefault("appname", defaultAppName)
	v.SetDefault("info.version", defaultAppVersion)
	v.SetDefault("info.debug", false)
}
