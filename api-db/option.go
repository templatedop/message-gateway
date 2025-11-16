package db

import "time"

type DBConfig struct {
	DBUsername        string        `mapstructure:"username"`
	DBPassword        string        `mapstructure:"password"`
	DBHost            string        `mapstructure:"host"`
	DBPort            string        `mapstructure:"port"`
	DBDatabase        string        `mapstructure:"database"`
	Schema            string        `mapstructure:"schema"`
	MaxConns          int32         `mapstructure:"maxconns"`
	MinConns          int32         `mapstructure:"minconns"`
	MaxConnLifetime   time.Duration `mapstructure:"maxconnlifetime"`
	MaxConnIdleTime   time.Duration `mapstructure:"maxconnidletime"`
	HealthCheckPeriod time.Duration `mapstructure:"healthcheckperiod"`
	AppName           string        `mapstructure:"appname"`
	SSLMode           string        `mapstructure:"sslmode"`
	Trace             bool          `mapstructure:"trace"`
}
