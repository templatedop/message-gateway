package db

import (
	"context"
	"fmt"
	"time"

	dbtracer "MgApplication/api-db/tracer"

	apierrors "MgApplication/api-errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	otelsdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// DBFactory interface to allow for extensibility
type DBFactory interface {
	NewPreparedDBConfig(input DBConfig) *DBConfig
	CreateConnection(dbConfig *DBConfig, osdktrace *otelsdktrace.TracerProvider, Registry *prometheus.Registry) (*DB, error)
	SetCollectorName(name string)
}

// DefaultDbFactory implements DBFactory interface
type DefaultDbFactory struct {
	CollectorName string
}

// SetCollectorName sets the CollectorName field
func (f *DefaultDbFactory) SetCollectorName(name string) {
	f.CollectorName = name
}

// NewDefaultDbFactory returns an instance of DefaultDbFactory
func NewDefaultDbFactory() DBFactory {
	return &DefaultDbFactory{
		CollectorName: "default_db_collector",
	}
}

// NewPreparedDBConfig creates and prepares the DBConfig from the input struct.
// This stage does not return an error, only prepares and validates the configuration.
func (f *DefaultDbFactory) NewPreparedDBConfig(input DBConfig) *DBConfig {

	// Initialize the DBConfig struct with values from the input
	dbConfig := &DBConfig{
		DBUsername:        input.DBUsername,
		DBPassword:        input.DBPassword,
		DBHost:            input.DBHost,
		DBPort:            input.DBPort,
		DBDatabase:        input.DBDatabase,
		Schema:            input.Schema,
		MaxConns:          input.MaxConns,
		MinConns:          input.MinConns,
		MaxConnLifetime:   time.Duration(input.MaxConnLifetime),
		MaxConnIdleTime:   time.Duration(input.MaxConnIdleTime),
		HealthCheckPeriod: time.Duration(input.HealthCheckPeriod),
		AppName:           input.AppName,
		SSLMode:           input.SSLMode,
		Trace:             input.Trace,
	}

	// Set defaults and validate the configuration
	validateDBConfig(dbConfig)

	return dbConfig
}

// CreateConnection uses the prepared DBConfig to establish a database connection.
func (f *DefaultDbFactory) CreateConnection(dbConfig *DBConfig, osdktrace *otelsdktrace.TracerProvider, Registry *prometheus.Registry) (*DB, error) {
	// Prepare the pgxpool.Config
	pgxConfig, err := Pgxconfig(dbConfig, osdktrace)
	if err != nil {
		appError := apierrors.NewAppError("pgxConfig Error", "500", err)
		return nil, &appError
	}

	// Create and return the DB connection
	conn, err := NewDB(dbConfig, pgxConfig, Registry, f.CollectorName)
	if err != nil {
		appError := apierrors.NewAppError("Error occurred while creating db connection", "500", err)
		return nil, &appError
	}

	return conn, nil
}

// Pgxconfig sets up the pgxpool configuration
func Pgxconfig(cfg *DBConfig, osdktrace *otelsdktrace.TracerProvider) (*pgxpool.Config, error) {
	dsn := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s search_path=%s sslmode=%s",
		cfg.DBUsername,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBDatabase,
		cfg.Schema,
		cfg.SSLMode,
	)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	if cfg.Trace {
		var tracer dbtracer.Tracer
		tracer, err = dbtracer.NewDBTracer(
			cfg.DBDatabase,
			dbtracer.WithTraceProvider(osdktrace),
		)
		if err != nil {
			return nil, err
		}

		if tracer != nil {
			config.ConnConfig.Tracer = tracer
		}
	}
	config.MaxConns = cfg.MaxConns
	config.MinConns = cfg.MinConns
	config.MaxConnLifetime = cfg.MaxConnLifetime * time.Minute
	config.MaxConnIdleTime = cfg.MaxConnIdleTime * time.Minute
	config.HealthCheckPeriod = cfg.HealthCheckPeriod * time.Minute
	config.ConnConfig.ConnectTimeout = 10 * time.Second
	config.ConnConfig.RuntimeParams = map[string]string{
		"application_name": cfg.AppName,
		"search_path":      cfg.Schema,
	}

	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheStatement
	config.ConnConfig.StatementCacheCapacity = 100
	config.ConnConfig.DescriptionCacheCapacity = 0
	return config, nil
}

// NewDB creates a new database connection with the given configuration
func NewDB(cfg *DBConfig, pcfg *pgxpool.Config, Registry *prometheus.Registry, collectorName string) (*DB, error) {

	ctx := context.Background()
	db, err := pgxpool.NewWithConfig(ctx, pcfg)
	if err != nil {
		return nil, err
	}

	collector := NewCollector(db, map[string]string{
		"db_name":        cfg.DBDatabase,
		"collector_name": collectorName,
	})
	Registry.MustRegister(collector)
	//	log.Info(nil, "collector in db:", collector)

	return &DB{
		db,
	}, nil
}

// validateDBConfig ensures critical fields are present and sets defaults for optional fields
func validateDBConfig(cfg *DBConfig) {

	// Validation for optional fields with defaults

	if cfg.MaxConns == 0 {
		cfg.MaxConns = 10 // Default max connections
	}

	if cfg.MinConns == 0 {
		cfg.MinConns = 1 // Default minimum connections
	}

	if cfg.MaxConnLifetime == 0 {
		cfg.MaxConnLifetime = 30 // Default 30 minutes
	}

	if cfg.MaxConnIdleTime == 0 {
		cfg.MaxConnIdleTime = 10 // Default 10 minutes
	}

	if cfg.HealthCheckPeriod == 0 {
		cfg.HealthCheckPeriod = 5 // Default 5 minutes
	}

}
