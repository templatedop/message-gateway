package bootstrapper

import (
	"context"
	// "errors" // Temporarily commented - only used in commented FxGrpc module
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	db "MgApplication/api-db"
	log "MgApplication/api-log"
	"MgApplication/api-server/swagger"

	auth "MgApplication/api-authz"
	config "MgApplication/api-config"
	// g "MgApplication/grpc-server" // Commented out - grpc-server not implemented yet

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/prometheus/client_golang/prometheus"
	otelsdktrace "go.opentelemetry.io/otel/sdk/trace"

	router "MgApplication/api-server"
	routeradapter "MgApplication/api-server/router-adapter"
	// Temporarily commented for testing - uncomment after fixing adapter compilation errors
	// _ "MgApplication/api-server/router-adapter/echo"
	// _ "MgApplication/api-server/router-adapter/fiber"
	// _ "MgApplication/api-server/router-adapter/gin"
	// _ "MgApplication/api-server/router-adapter/nethttp"

	tclient "go.temporal.io/sdk/client"
	"go.uber.org/fx"
	"golang.org/x/sync/errgroup"

	fxhealthcheck "MgApplication/api-fxhealth"
	healthcheck "MgApplication/api-healthcheck"
	fxmetrics "MgApplication/api-metrics"
)

const (
	ReadDBProbeName      = "read-db-probe"
	WriteDBProbeName     = "write-db-probe"
	ReadDBCollectorName  = "read_db_collector"
	WriteDBCollectorName = "write_db_collector"
)

type Bootstrapper struct {
	context context.Context
	options []fx.Option
}

func New() *Bootstrapper {
	return &Bootstrapper{
		context: context.Background(),
		options: []fx.Option{
			fxconfig,
			fxlog,
			fxDB,
			fxRouterAdapter, // Router adapter system - supports gin, fiber, echo, nethttp
			// fxrouter,      // Old router module (Gin only) - kept for backward compatibility
			fxTrace,
			fxMetrics,
			//fxHealthcheck,
		},
	}
}

func (b *Bootstrapper) WithContext(ctx context.Context) *Bootstrapper {
	b.context = ctx

	return b
}

func (b *Bootstrapper) Options(options ...fx.Option) *Bootstrapper {
	b.options = append(b.options, options...)

	return b
}

func (b *Bootstrapper) BootstrapApp(options ...fx.Option) *fx.App {
	return fx.New(
		fx.Supply(fx.Annotate(b.context, fx.As(new(context.Context)))),
		fx.Options(b.options...),
		fx.Options(options...),
	)
}

func (b *Bootstrapper) Run(options ...fx.Option) {
	// Wrap the context with signal detection for graceful shutdown
	// Listen for SIGINT (Ctrl+C) and SIGTERM (kill command)
	ctx, cancel := signal.NotifyContext(b.context, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	// Update the bootstrapper context with signal-aware context
	b.context = ctx

	// Monitor context cancellation in a separate goroutine
	go func() {
		<-ctx.Done()
		if ctx.Err() == context.Canceled {
			log.GetBaseLoggerInstance().ToZerolog().Info().Msg("Shutdown signal received, initiating graceful shutdown...")
		}
	}()

	// Create and run the FX application
	app := b.BootstrapApp(options...)

	// Run the application with signal handling
	// When a signal is received, the context will be cancelled and fx will gracefully shutdown
	app.Run()

	log.GetBaseLoggerInstance().ToZerolog().Info().Msg("Application shutdown complete")
}

var fxHealthcheck = fx.Module(
	"healthcheck",

	fx.Provide(
		healthcheck.NewDefaultCheckerFactory,
		fxhealthcheck.NewFxCheckerProbeRegistry,
		fx.Annotated{
			Name:   "health_checker",
			Target: fxhealthcheck.NewFxChecker,
		},
		fxhealthcheck.NewFxChecker,
	),
)
var fxconfig = fx.Module(
	"configmodule",
	fx.Provide(
		config.NewDefaultConfigFactory,
		newFxConfig,
	),
)

type FxConfigParam struct {
	fx.In
	Factory config.ConfigFactory
}

func newFxConfig(p FxConfigParam) (*config.Config, error) {
	return p.Factory.Create(
		config.WithFileName("config"),
		config.WithAppEnv(os.Getenv("APP_ENV")),
		config.WithFilePaths(
			".",
			"./configs",
			os.Getenv("APP_CONFIG_PATH"),
		),
	)
}

var fxlog = fx.Module(
	"logmodule",
	fx.Provide(
		log.NewDefaultLoggerFactory,
	),
	fx.Invoke(newFxLogger),
)

type FxLogParam struct {
	fx.In
	Factory log.LoggerFactory
	Config  *config.Config
}

func newFxLogger(p FxLogParam) error {

	var version string
	if p.Config.Exists("info.version") {
		version = p.Config.GetString("info.version")
	}

	level := log.FetchLogLevel(p.Config.GetString("log.level"))
	err := p.Factory.Create(
		log.WithServiceName(p.Config.AppName()),
		log.WithLevel(level),
		log.WithOutputWriter(os.Stdout),
		log.WithVersion(version),
	)
	if err != nil {
		return err
	}

	return nil
}

func dbreadconfig(c *config.Config) db.DBConfig {

	var trace bool
	if c.Exists("db.trace.enabled") {
		trace = c.GetBool("db.trace.enabled")
	}

	dbconfig := db.DBConfig{

		DBUsername:        c.GetString("db.read.username"),
		DBPassword:        c.GetString("db.read.password"),
		DBHost:            c.GetString("db.read.host"),
		DBPort:            c.GetString("db.read.port"),
		DBDatabase:        c.GetString("db.read.database"),
		Schema:            c.GetString("db.read.schema"),
		MaxConns:          c.GetInt32("db.read.maxconns"),
		MinConns:          c.GetInt32("db.read.minconns"),
		MaxConnLifetime:   time.Duration(c.GetInt("db.read.maxconnlifetime")),
		MaxConnIdleTime:   time.Duration(c.GetInt("db.read.maxconnidletime")),
		HealthCheckPeriod: time.Duration(c.GetInt("db.read.healthcheckperiod")),
		Trace:             trace,
		AppName:           c.AppName(),
	}

	// return fx.Annotated{
	// 	Name:   "read_config",
	// 	Target: dbconfig,
	// }
	return dbconfig

}

var FxReadDB = fx.Module(
	"Read DBModule",
	fx.Provide(

		fx.Annotated{
			Name:   "read_config",
			Target: dbreadconfig},

		fx.Annotated{
			Name:   "read_prepared_config",
			Target: db.NewDefaultDbFactory().NewPreparedDBConfig,
		},

		fx.Annotated{
			Name: "read_db",
			Target: func(params struct {
				fx.In
				Config    db.DBConfig `name:"read_config"`
				Osdktrace *otelsdktrace.TracerProvider
				Registry  *prometheus.Registry
			}) (*db.DB, error) {
				factory := db.NewDefaultDbFactory()
				factory.SetCollectorName(ReadDBCollectorName)
				//factory.ReadDBCollectorName = ReadDBCollectorName
				return factory.CreateConnection(&params.Config, params.Osdktrace, params.Registry)
			},
			//Target: db.NewDefaultDbFactory().CreateConnection,
		},
		//db.NewDefaultDbFactory().CreateConnection,
	),
	fx.Invoke(readdblifecycle),
	fxhealthcheck.AsCheckerProbe(func(p readDBProbeParams) healthcheck.CheckerProbe {
		probe := db.NewSQLProbe(p.DB)
		probe.SetName(ReadDBProbeName)
		return probe
	}),
)

type readDBLifecycleParams struct {
	fx.In
	Ctx context.Context // Signal-aware context from bootstrapper
	DB  *db.DB          `name:"read_db"`
	LC  fx.Lifecycle
}

func readdblifecycle(p readDBLifecycleParams) {
	p.LC.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				log.GetBaseLoggerInstance().ToZerolog().Info().Str("module", "ReadDBModule").Msg("Starting read database module")

				// Use context-aware ping
				err := p.DB.PingContext(ctx)
				if err != nil {
					return err
				}

				log.GetBaseLoggerInstance().ToZerolog().Info().Msg("Successfully connected to read database")
				return nil
			},
			OnStop: func(ctx context.Context) error {
				logger := log.GetBaseLoggerInstance().ToZerolog()

				// Log connection stats before shutdown
				if count := p.DB.Stat(); count != nil {
					logger.Info().
						Int32("total_conns", count.TotalConns()).
						Int32("idle_conns", count.IdleConns()).
						Int32("acquired_conns", count.AcquiredConns()).
						Msg("Read database connection stats at shutdown start")
				}

				// Wait for active connections to drain with timeout
				// This allows in-flight HTTP requests to complete their DB operations
				drainTimeout := 5 * time.Second
				drainCtx, cancel := context.WithTimeout(context.Background(), drainTimeout)
				defer cancel()

				logger.Info().
					Dur("drain_timeout", drainTimeout).
					Msg("Waiting for read database connections to drain...")

				ticker := time.NewTicker(100 * time.Millisecond)
				defer ticker.Stop()

				for {
					select {
					case <-drainCtx.Done():
						// Timeout reached, force close
						if count := p.DB.Stat(); count != nil {
							logger.Warn().
								Int32("remaining_acquired", count.AcquiredConns()).
								Msg("Read DB drain timeout reached, forcing database closure")
						}
						goto closeDB

					case <-ticker.C:
						// Check if all connections are idle
						if count := p.DB.Stat(); count != nil {
							if count.AcquiredConns() == 0 {
								logger.Info().Msg("All read database connections drained successfully")
								goto closeDB
							}
						}
					}
				}

			closeDB:
				// Close the database connection pool
				p.DB.Close()

				// Log final stats
				if count := p.DB.Stat(); count != nil {
					logger.Info().
						Int32("final_total_conns", count.TotalConns()).
						Msg("Read database connection pool closed")
				}

				logger.Info().Msg("Read database shutdown complete")
				return nil
			},
		},
	)
}

func dbconfig(c *config.Config) db.DBConfig {

	var sslmode string
	if c.Exists("db.sslmode") {
		sslmode = c.GetString("db.sslmode")
	} else {
		sslmode = "disable"
	}
	var trace bool
	if c.Exists("trace.enabled") {
		trace = c.GetBool("trace.enabled")
		log.Info(nil, "DB trace is enabled!!")
	}

	dbconfig := db.DBConfig{

		DBUsername:        c.GetString("db.username"),
		DBPassword:        c.GetString("db.password"),
		DBHost:            c.GetString("db.host"),
		DBPort:            c.GetString("db.port"),
		DBDatabase:        c.GetString("db.database"),
		Schema:            c.GetString("db.schema"),
		MaxConns:          c.GetInt32("db.maxconns"),
		MinConns:          c.GetInt32("db.minconns"),
		MaxConnLifetime:   time.Duration(c.GetInt("db.maxconnlifetime")),
		MaxConnIdleTime:   time.Duration(c.GetInt("db.maxconnidletime")),
		HealthCheckPeriod: time.Duration(c.GetInt("db.healthcheckperiod")),
		SSLMode:           sslmode,
		Trace:             trace,
		AppName:           c.AppName(),
	}

	// return fx.Annotated{
	// 	Name:   "write_config",
	// 	Target: dbconfig,
	// }
	return dbconfig

}

var fxDB = fx.Module(
	"Write DBModule",
	fx.Provide(
		fx.Annotated{
			Name:   "write_config",
			Target: dbconfig},
		fx.Annotated{
			Name:   "write_prepared_config",
			Target: db.NewDefaultDbFactory().NewPreparedDBConfig,
		},
		fx.Annotated{
			Name: "write_db",
			Target: func(params struct {
				fx.In
				Config    db.DBConfig `name:"write_config"`
				Osdktrace *otelsdktrace.TracerProvider
				Registry  *prometheus.Registry
			}) (*db.DB, error) {
				factory := db.NewDefaultDbFactory()
				factory.SetCollectorName(WriteDBCollectorName)
				return factory.CreateConnection(&params.Config, params.Osdktrace, params.Registry)
			},
			// Target: db.NewDefaultDbFactory().CreateConnection,
		},
		// Bridge provider: expose the named write_db also as the default *db.DB so
		// constructors that request *db.DB without a name (repositories) receive it.
		func(p struct {
			fx.In
			Write *db.DB `name:"write_db"`
		}) *db.DB {
			return p.Write
		},
		//db.NewDefaultDbFactory().CreateConnection,
	),

	fx.Invoke(dblifecycle),
	fxhealthcheck.AsCheckerProbe(func(p writeDBProbeParams) healthcheck.CheckerProbe {
		probe := db.NewSQLProbe(p.DB)
		probe.SetName(WriteDBProbeName)
		return probe
	}),
)

type writeDBProbeParams struct {
	fx.In
	DB *db.DB `name:"write_db"`
}

type readDBProbeParams struct {
	fx.In
	DB *db.DB `name:"read_db"`
}

type writeDBLifecycleParams struct {
	fx.In
	Ctx context.Context // Signal-aware context from bootstrapper
	DB  *db.DB          `name:"write_db"`
	LC  fx.Lifecycle
}

func dblifecycle(p writeDBLifecycleParams) {
	p.LC.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				log.GetBaseLoggerInstance().ToZerolog().Info().Str("module", "DBModule").Msg("Starting fxdb module")

				// Use context-aware ping
				err := p.DB.PingContext(ctx)
				if err != nil {
					return err
				}

				log.GetBaseLoggerInstance().ToZerolog().Info().Msg("Successfully connected to the database")
				return nil
			},
			OnStop: func(ctx context.Context) error {
				logger := log.GetBaseLoggerInstance().ToZerolog()

				// Log connection stats before shutdown
				if count := p.DB.Stat(); count != nil {
					logger.Info().
						Int32("total_conns", count.TotalConns()).
						Int32("idle_conns", count.IdleConns()).
						Int32("acquired_conns", count.AcquiredConns()).
						Msg("Database connection stats at shutdown start")
				}

				// Wait for active connections to drain with timeout
				// This allows in-flight HTTP requests to complete their DB operations
				drainTimeout := 5 * time.Second
				drainCtx, cancel := context.WithTimeout(context.Background(), drainTimeout)
				defer cancel()

				logger.Info().
					Dur("drain_timeout", drainTimeout).
					Msg("Waiting for active database connections to drain...")

				ticker := time.NewTicker(100 * time.Millisecond)
				defer ticker.Stop()

				for {
					select {
					case <-drainCtx.Done():
						// Timeout reached, force close
						if count := p.DB.Stat(); count != nil {
							logger.Warn().
								Int32("remaining_acquired", count.AcquiredConns()).
								Msg("Drain timeout reached, forcing database closure")
						}
						goto closeDB

					case <-ticker.C:
						// Check if all connections are idle
						if count := p.DB.Stat(); count != nil {
							if count.AcquiredConns() == 0 {
								logger.Info().Msg("All database connections drained successfully")
								goto closeDB
							}
						}
					}
				}

			closeDB:
				// Close the database connection pool
				p.DB.Close()

				// Log final stats
				if count := p.DB.Stat(); count != nil {
					logger.Info().
						Int32("final_total_conns", count.TotalConns()).
						Msg("Database connection pool closed")
				}

				logger.Info().Msg("Database shutdown complete")
				return nil
			},
		},
	)
}

var Fxclient = fx.Module(
	"client",
	fx.Invoke(client),
)

func client() error {

	defaultTimeout := 10 * time.Second
	defaultMaxRetries := 3
	defaultRetryWait := 500 * time.Millisecond
	defaultMaxRetryWait := 3 * time.Second
	err := auth.Init(
		auth.ClientConfig{
			Timeout:      defaultTimeout,
			RetryWait:    defaultRetryWait,
			MaxRetryWait: defaultMaxRetryWait,
			MaxRetries:   defaultMaxRetries,
		},
	)
	if err != nil {
		return err
	}

	return nil

}

var fxrouter = fx.Module(
	"router",

	fx.Provide(router.ParseGroupedControllers, router.Defaultgin, router.GetSwaggerDefs),
	swagger.FxGenerateSwagger,
	fx.Invoke(startServer),
)

func startServer(lc fx.Lifecycle, sv *router.Router) {
	eg, _ := errgroup.WithContext(context.Background())

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			return sv.Shutdown(ctx)
		},
	})

	eg.Go(func() error {
		if err := sv.Start(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})

}

// fxRouterAdapter is the new FX module that uses router-adapter system
// This allows switching between different web frameworks (Gin, Fiber, Echo, net/http)
// via configuration instead of being hard-coded to Gin
var fxRouterAdapter = fx.Module(
	"router-adapter",
	fx.Provide(
		newRouterAdapter,
	),
	fx.Invoke(startRouterAdapter),
)

// routerAdapterParams holds the dependencies for creating a router adapter
type routerAdapterParams struct {
	fx.In
	Ctx       context.Context
	Config    *config.Config
	Osdktrace *otelsdktrace.TracerProvider
	Registry  *prometheus.Registry
}

// newRouterAdapter creates and configures a router adapter from config
func newRouterAdapter(p routerAdapterParams) (routeradapter.RouterAdapter, error) {
	// Adapter packages are imported at the top of file to register factories
	// This allows the factory registry to be populated during init()

	// Create router config from application config
	cfg := routeradapter.DefaultRouterConfig()

	// Determine router type from config (default to Gin)
	routerType := routeradapter.RouterTypeGin
	if p.Config.Exists("router.type") {
		routerType = routeradapter.RouterType(p.Config.GetString("router.type"))
	}
	cfg.Type = routerType

	// Set server configuration
	if p.Config.Exists("server.addr") {
		cfg.Port = p.Config.GetInt("server.port")
	}

	// Create the adapter
	adapter, err := routeradapter.NewRouterAdapter(cfg)
	if err != nil {
		return nil, err
	}

	// Set the signal-aware context
	adapter.SetContext(p.Ctx)

	// Note: Routes and middlewares will be registered from the application layer

	return adapter, nil
}

// routerAdapterLifecycleParams holds dependencies for router adapter lifecycle
type routerAdapterLifecycleParams struct {
	fx.In
	LC      fx.Lifecycle
	Adapter routeradapter.RouterAdapter
	Config  *config.Config
}

// startRouterAdapter manages the router adapter lifecycle
func startRouterAdapter(p routerAdapterLifecycleParams) {
	eg, _ := errgroup.WithContext(context.Background())

	p.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Get server address from config
			addr := ":8080"
			if p.Config.Exists("server.addr") {
				addr = p.Config.GetString("server.addr")
			}

			// Start server in background
			eg.Go(func() error {
				if err := p.Adapter.Start(addr); err != nil && err != http.ErrServerClosed {
					return err
				}
				return nil
			})

			log.GetBaseLoggerInstance().ToZerolog().Info().
				Str("adapter", string(p.Config.GetString("router.type"))).
				Str("address", addr).
				Msg("Router adapter started")

			return nil
		},
		OnStop: func(ctx context.Context) error {
			shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			if err := p.Adapter.Shutdown(shutdownCtx); err != nil {
				return err
			}

			log.GetBaseLoggerInstance().ToZerolog().Info().
				Msg("Router adapter shutdown complete")

			return nil
		},
	})
}

type FxMinioParam struct {
	fx.In
	Factory log.LoggerFactory
	Config  *config.Config
}

func newFxMinio(p FxMinioParam) {
	var err error
	var MinioClient *minio.Client

	MinioClient, err = minio.New(p.Config.GetString("minio.url"), &minio.Options{
		Creds:  credentials.NewStaticV4(p.Config.GetString("minio.AccessKey"), p.Config.GetString("minio.SecretKey"), ""),
		Secure: true})
	if err != nil {
		log.GetBaseLoggerInstance().ToZerolog().Error().Msg("Minio Client Error")
	}

	exists, errBucketExists := MinioClient.BucketExists(context.Background(), p.Config.GetString("minio.BucketName"))

	if errBucketExists != nil {
		log.GetBaseLoggerInstance().ToZerolog().Error().Msg("Error checking if bucket exists:")
	}

	if exists {
		log.GetBaseLoggerInstance().ToZerolog().Debug().Msg("Bucket found")
	} else {
		log.GetBaseLoggerInstance().ToZerolog().Error().Msg("Bucket does not exist")

	}

}

var FxMinIO = fx.Module(
	"MinIOModule",

	fx.Provide(func(p FxMinioParam) (*minio.Client, error) {
		return minio.New(p.Config.GetString("minio.url"), &minio.Options{
			Creds:  credentials.NewStaticV4(p.Config.GetString("minio.AccessKey"), p.Config.GetString("minio.SecretKey"), ""),
			Secure: true,
		})
	}),
	fx.Invoke(newFxMinio),
)

var Fxtemporal = fx.Module(
	"temporal",
	fx.Provide(
		temporalclient,
		//ProvideTemporalWorker,
	),
	fx.Invoke(temporallifecycle),
	// Temporal Client Initialization

)

func temporalclient(c *config.Config) (temporalclient tclient.Client, err error) {
	TemporalHost := c.GetString("temporal.host")
	TemporalPort := c.GetString("temporal.port")
	hostPort := TemporalHost + ":" + TemporalPort

	temporalClient, err := tclient.Dial(tclient.Options{
		HostPort: hostPort,
	})
	if err != nil {
		return nil, err
	}
	return temporalClient, nil

}
func temporallifecycle(lc fx.Lifecycle, temporalclient tclient.Client) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			return nil
		},
		OnStop: func(ctx context.Context) error {
			temporalclient.Close()
			return nil
		},
	})

}

// var compresskb connect.Option = connect.WithCompressMinBytes(1024)
var addr = ":8083"

// FxGrpc module - Commented out until grpc-server package is implemented
/*
var FxGrpc = fx.Module(
	"gRPCmodule",

	fx.Provide(
		g.NewHandlerRegistry,
	),
	fx.Invoke(func(lc fx.Lifecycle, srv *http.Server) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {

				go func() {
					if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
						log.Fatal(context.TODO(), "HTTP listen and serve: %v", err)
						return
					}
				}()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
				defer cancel()
				if err := srv.Shutdown(ctx); err != nil {
					log.Fatal(context.TODO(), "HTTP shutdown: %v", err)
					return err
				}
				return nil
			},
		})
	}),
)
*/

var fxMetrics = fx.Module(
	"metrics",
	fx.Provide(
		fxmetrics.NewDefaultMetricsRegistryFactory,
		fxmetrics.NewFxMetricsRegistry,
	),
)
