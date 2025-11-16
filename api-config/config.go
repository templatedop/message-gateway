package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

const (
	appEnvProd        = "prod"     // prod environment
	appEnvDev         = "dev"      // dev environment
	appEnvTest        = "test"     // test environment
	defaultAppName    = "app"      // default application name
	defaultAppVersion = "unknown"  // default application version
	appEnvPrepod      = "prepod"   // prepod environment
	appEnvTraining    = "training" // training environment

)

type Config struct {
	*viper.Viper
}

func NewConfig(v *viper.Viper) *Config {
	return &Config{v}
}

func (c *Config) Exists(key string) bool {
	return c.IsSet(key)
}
func (c *Config) Of(section string) (*Config, error) {
	subViper := c.Sub(section)
	if subViper == nil {
		return nil, fmt.Errorf("could not load config file for env %s", section)
	}
	return &Config{subViper}, nil
}

func ToStruct[T any](v *Config, root string, cfgStruct *T) error {
	subViper := v.Sub(root)
	if subViper == nil {
		return fmt.Errorf("no sub-config found for root key: %s", root)
	}
	return subViper.Unmarshal(cfgStruct)
}

func (c *Config) GetEnvVar(envVar string) string {
	return os.Getenv(envVar)
}

func (c *Config) AppName() string {
	return c.GetString("appname")
}

func (c *Config) AppEnv() string {
	return c.GetString("info.env")
}

func (c *Config) AppVersion() string {
	return c.GetString("info.version")
}

func (c *Config) AppDebug() bool {
	return c.GetBool("info.debug")
}

func (c *Config) IsProdEnv() bool {
	return c.AppEnv() == appEnvProd
}

func (c *Config) IsDevEnv() bool {
	return c.AppEnv() == appEnvDev
}

func (c *Config) IsTestEnv() bool {
	return c.AppEnv() == appEnvTest
}
func (c *Config) IsPreProdEnv() bool {
	return c.AppEnv() == appEnvPrepod
}
func (c *Config) IsTrainingEnv() bool {
	return c.AppEnv() == appEnvTraining
}
