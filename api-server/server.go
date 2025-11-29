package router

import (
	"context"
	"embed"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	config "MgApplication/api-config"
	apierrors "MgApplication/api-errors"

	"github.com/arl/statsviz"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	//healthcheck "MgApplication/api-healthcheck"
	log "MgApplication/api-log"
	//health "MgApplication/api-server/health"
	"MgApplication/api-server/middlewares"
	prof "MgApplication/api-server/pprof"
	"MgApplication/api-server/ratelimiter"
	rate "MgApplication/api-server/ratelimiter"
	"MgApplication/api-server/route"
	"MgApplication/api-server/util/slc"

	otelsdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var (
	activeConnections int64
	//go:embed templates/*
	templatesFS  embed.FS
	globalBucket *rate.LeakyBucket
)

const (
	defaultMaxConnections         = 1000
	DefaultDebugPProfPath         = "/debug/pprof"
	ThemeLight                    = "light"
	ThemeDark                     = "dark"
	DefaultDebugStatsPath         = "/debug/statsviz"
	DefaultHealthCheckStartupPath = "/healthzz"
	DefaultMetricsPath            = "/metrics"
	DefaultRate                   = 300
	DefaultCapacity               = 700
)

type DashboardTheme struct {
	Theme string `form:"theme" json:"theme"`
}
type Router struct {
	ctx               context.Context
	app               *gin.Engine
	cfg               *config.Config
	Addr              string
	MaxConnections    int64
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	ConnState         func(net.Conn, http.ConnState)
	RequestTimeout    time.Duration
	registries        []*registry
}

func (s *Router) handleConnState(conn net.Conn, state http.ConnState) {
	if s.ConnState != nil {
		s.ConnState(conn, state)
	}

	switch state {
	case http.StateNew:
		// Atomically increment first, then check - prevents TOCTOU race
		newCount := atomic.AddInt64(&activeConnections, 1)

		if newCount > s.MaxConnections {
			// Over limit - decrement back and reject
			atomic.AddInt64(&activeConnections, -1)
			IncRejectedConnections()

			log.GetBaseLoggerInstance().ToZerolog().Warn().
				Int64("activeConnections", newCount-1).
				Int64("maxConnections", s.MaxConnections).
				Msg("Max connections reached - connection rejected")
			conn.Close()
			return
		}

		// Under limit - connection accepted
		IncActiveConnections()

	case http.StateClosed, http.StateHijacked:
		atomic.AddInt64(&activeConnections, -1)
		DecActiveConnections()
	}
}

var ginserver *http.Server

func (s *Router) RegisterRoutes() {
	slc.ForEach(s.registries, func(r *registry) {
		metas := slc.Map(r.routes, r.toMeta)

		slc.ForEach(metas, func(m route.Meta) {
			handlers := []gin.HandlerFunc{}

			// Add middlewares from registry
			for _, mw := range r.mws {
				handlers = append(handlers, gin.HandlerFunc(mw))
			}

			// Add route-specific middlewares
			for _, mw := range m.Middlewares {
				handlers = append(handlers, gin.HandlerFunc(mw))
			}

			// Add the main route handler function
			handlers = append(handlers, gin.HandlerFunc(m.Func))

			// Add route to Gin router
			s.app.Handle(m.Method, m.Path, handlers...)
		})
	})
}

func (s *Router) Start() error {
	//register routes
	ginserver = &http.Server{
		Addr:              s.Addr,
		Handler:           s.app,
		ReadTimeout:       s.ReadTimeout,
		ReadHeaderTimeout: s.ReadHeaderTimeout,
		WriteTimeout:      s.WriteTimeout,
		IdleTimeout:       s.IdleTimeout,
		ConnState:         s.handleConnState,
		// BaseContext provides the signal-aware context to all HTTP handlers
		// This allows handlers to detect shutdown signals via req.Context()
		BaseContext: func(net.Listener) context.Context {
			return s.ctx
		},
	}

	return ginserver.ListenAndServe()

}

func (s *Router) Shutdown(ctx context.Context) error {
	return ginserver.Shutdown(ctx)

}

func NewRouter(app *gin.Engine, cfg *config.Config, registries []*registry) *Router {
	return &Router{
		app:        app,
		cfg:        cfg,
		registries: registries,
	}
}

// ============================================================================
// HELPER FUNCTIONS FOR SERVER INITIALIZATION
// ============================================================================

// configureGinMode sets the Gin framework mode based on configuration
func configureGinMode(cfg *config.Config) {
	if cfg.Exists("server.env") {
		switch cfg.GetString("server.env") {
		case "production":
			gin.SetMode(gin.ReleaseMode)
		case "test":
			gin.SetMode(gin.TestMode)
		case "debug":
			gin.SetMode(gin.DebugMode)
		default:
			gin.SetMode(gin.ReleaseMode)
		}
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
}

// configureRateLimiting sets up rate limiting middleware based on configuration
func configureRateLimiting(app *gin.Engine, cfg *config.Config, metricsRegistry *prometheus.Registry) {
	ratelimit := "medium" // default
	if cfg.Exists("server.ratelimit") {
		ratelimit = cfg.GetString("server.ratelimit")
	}

	switch ratelimit {
	case "verylow":
		globalBucket = rate.NewLeakyBucket(100, 300)
	case "low":
		globalBucket = rate.NewLeakyBucket(200, 450)
	case "medium":
		globalBucket = rate.NewLeakyBucket(DefaultRate, DefaultCapacity)
	case "high":
		globalBucket = rate.NewLeakyBucket(400, 900)
	case "veryhigh":
		globalBucket = rate.NewLeakyBucket(500, 1100)
	default:
		globalBucket = rate.NewLeakyBucket(DefaultRate, DefaultCapacity)
	}

	app.Use(middlewares.RateMiddleware(globalBucket))
	ratelimiter.InitMetrics(globalBucket, metricsRegistry)
}

// registerCoreMiddlewares adds body limiter, rate limiter, CORS, recovery, and error handler
func registerCoreMiddlewares(app *gin.Engine, cfg *config.Config, metricsRegistry *prometheus.Registry) {
	// Get server config with fallback
	serverCfg, err := cfg.Of("server")
	if err != nil {
		log.Error(nil, "Failed to get server config, using root config: %v", err)
		serverCfg = cfg
	}

	// Get CORS config
	corsCfg, err := serverCfg.Of("cors")
	if err != nil {
		log.Warn(nil, "Failed to get CORS config, CORS middleware will use empty defaults: %v", err)
	}

	// Configure body size limit
	defaultsizelimit := int64(2 * 1024 * 1024)
	sizelimit := cfg.GetInt64("server.bodylimit")
	if sizelimit == 0 {
		sizelimit = defaultsizelimit
	}

	app.Use(
		middlewares.BodyLimiter(sizelimit),
		middlewares.BodyLimitErrorHandler())

	// Configure rate limiting
	configureRateLimiting(app, cfg, metricsRegistry)

	// Add core middlewares
	app.Use(
		middlewares.CORSMiddleware(corsCfg),
		middlewares.Recover(cfg),
		middlewares.ErrorHandler(),
	)
}

// registerSecurityMiddlewares adds encryption/decryption middleware if enabled
func registerSecurityMiddlewares(app *gin.Engine, cfg *config.Config) {
	encryptenabled := false
	if cfg.Exists("server.encrypt") {
		encryptenabled = cfg.GetBool("server.encrypt")
	}

	if encryptenabled {
		app.Use(middlewares.DecryptMiddleware())
		app.Use(middlewares.ResponseSignatureMiddleware())
	}
}

// parseMetricBuckets parses metric bucket configuration from config string
func parseMetricBuckets(cfg *config.Config) []float64 {
	var buckets []float64
	if bucketsConfig := cfg.GetString("metrics.buckets"); bucketsConfig != "" {
		for _, s := range Split(bucketsConfig) {
			f, err := strconv.ParseFloat(s, 64)
			if err == nil {
				buckets = append(buckets, f)
			}
		}
	}
	return buckets
}

// registerObservabilityMiddlewares adds tracing, logging, and metrics middleware
func registerObservabilityMiddlewares(app *gin.Engine, cfg *config.Config,
	osdktrace *otelsdktrace.TracerProvider, metricsRegistry *prometheus.Registry) {

	// Configure tracing
	if cfg.GetBool("trace.enabled") {
		app.Use(middlewares.RequestTracerMiddleware(
			cfg.AppName(),
			middlewares.RequestTracerMiddlewareConfig{
				TracerProvider: AnnotateTracerProvider(osdktrace),
			},
		))
	}

	// Configure logging
	app.Use(
		middlewares.SetCtxLoggerMiddleware(),
		middlewares.RequestResponseLoggerMiddleware())

	// Configure metrics
	if cfg.GetBool("metrics.collect.routes") {
		buckets := parseMetricBuckets(cfg)
		metricsMiddlewareConfig := middlewares.RequestMetricsMiddlewareConfig{
			Registry:                metricsRegistry,
			Namespace:               "",
			Subsystem:               Sanitize("router"),
			Buckets:                 buckets,
			NormalizeRequestPath:    true,
			NormalizeResponseStatus: true,
		}
		app.Use(middlewares.RequestMetricsMiddlewareWithConfig(metricsMiddlewareConfig))
	}
}

// registerPprofEndpoints registers performance profiling endpoints
func registerPprofEndpoints(app *gin.Engine, cfg *config.Config) {
	pprofPath := cfg.GetString("server.debug.pprof.path")
	if pprofPath == "" {
		pprofPath = DefaultDebugPProfPath
	}

	pprofGroup := app.Group(pprofPath)
	pprofGroup.GET("/", prof.PprofIndexHandler())
	pprofGroup.GET("/allocs", prof.PprofAllocsHandler())
	pprofGroup.GET("/block", prof.PprofBlockHandler())
	pprofGroup.GET("/cmdline", prof.PprofCmdlineHandler())
	pprofGroup.GET("/goroutine", prof.PprofGoroutineHandler())
	pprofGroup.GET("/heap", prof.PprofHeapHandler())
	pprofGroup.GET("/mutex", prof.PprofMutexHandler())
	pprofGroup.GET("/profile", prof.PprofProfileHandler())
	pprofGroup.GET("/symbol", prof.PprofSymbolHandler())
	pprofGroup.POST("/symbol", prof.PprofSymbolHandler())
	pprofGroup.GET("/threadcreate", prof.PprofThreadCreateHandler())
	pprofGroup.GET("/trace", prof.PprofTraceHandler())

	log.GetBaseLoggerInstance().ToZerolog().Debug().Msg("registered debug pprof handlers")
}

// registerDashboardEndpoints registers dashboard UI endpoints
func registerDashboardEndpoints(app *gin.Engine, cfg *config.Config) {
	renderer, err := NewDashboardRenderer(templatesFS, "templates/dashboard.html")
	if err != nil {
		panic(err)
	}

	// Get config values once
	statsExpose := cfg.GetBool("server.debug.stats.expose")
	statsPath := cfg.GetString("server.debug.stats.path")
	metricsExpose := cfg.GetBool("metrics.expose")
	metricsPath := cfg.GetString("metrics.path")
	if metricsPath == "" {
		metricsPath = DefaultMetricsPath
	}
	pprofExpose := cfg.GetBool("server.debug.pprof.expose")
	pprofPath := cfg.GetString("server.debug.pprof.path")

	// Theme switching endpoint
	app.POST("/theme", func(c *gin.Context) {
		themeCookie := &http.Cookie{Name: "theme"}

		var theme DashboardTheme
		if err := c.ShouldBind(&theme); err != nil {
			themeCookie.Value = ThemeLight
		} else {
			switch theme.Theme {
			case ThemeDark:
				themeCookie.Value = ThemeDark
			case ThemeLight:
				themeCookie.Value = ThemeLight
			default:
				themeCookie.Value = ThemeLight
			}
		}

		http.SetCookie(c.Writer, themeCookie)
		c.Redirect(http.StatusMovedPermanently, "/")
	})

	log.GetBaseLoggerInstance().ToZerolog().Debug().Msg("registered dashboard theme handler")

	// Dashboard rendering endpoint
	app.GET("/dashboard", func(c *gin.Context) {
		theme := ThemeLight
		if themeCookie, err := c.Cookie("theme"); err == nil {
			switch themeCookie {
			case ThemeDark:
				theme = ThemeDark
			case ThemeLight:
				theme = ThemeLight
			}
		}

		renderer.Render(c, "dashboard.html", gin.H{
			"statsExpose":   statsExpose,
			"statsPath":     statsPath,
			"metricsExpose": metricsExpose,
			"metricsPath":   metricsPath,
			"pprofExpose":   pprofExpose,
			"pprofPath":     pprofPath,
			"theme":         theme,
		})
	})

	log.GetBaseLoggerInstance().ToZerolog().Debug().Msg("registered dashboard handler")
}

// registerStatsvizEndpoints registers runtime statistics visualization endpoints
func registerStatsvizEndpoints(app *gin.Engine, cfg *config.Config) {
	statsPath := cfg.GetString("server.debug.stats.path")
	if statsPath == "" {
		statsPath = DefaultDebugStatsPath
	}

	srv, err := statsviz.NewServer()
	if err != nil {
		panic(err)
	}

	debug := app.Group(statsPath)
	debug.GET("/*filepath", func(c *gin.Context) {
		if c.Param("filepath") == "/ws" {
			srv.Ws()(c.Writer, c.Request)
			return
		}
		srv.Index()(c.Writer, c.Request)
	})
}

// registerDebugEndpoints registers pprof, dashboard, statsviz, and metrics endpoints
func registerDebugEndpoints(app *gin.Engine, cfg *config.Config, metricsRegistry *prometheus.Registry) {
	// Metrics endpoint
	if cfg.GetBool("metrics.expose") {
		metricsPath := DefaultMetricsPath
		if metricsPath == "" {
			metricsPath = "/metrics"
		}
		app.GET(metricsPath, gin.WrapH(promhttp.HandlerFor(metricsRegistry, promhttp.HandlerOpts{})))
		log.GetBaseLoggerInstance().ToZerolog().Debug().Msg("registered metrics handler")
	}

	// Pprof endpoints
	if cfg.GetBool("server.debug.pprof.expose") {
		registerPprofEndpoints(app, cfg)
	}

	// Dashboard endpoints
	if cfg.GetBool("server.dashboard.enabled") {
		registerDashboardEndpoints(app, cfg)
	}

	// Statsviz endpoints
	if cfg.GetBool("server.debug.stats.expose") {
		registerStatsvizEndpoints(app, cfg)
	}
}

// createAndConfigureRouter creates router and configures connection limits, timeouts, and metrics
func createAndConfigureRouter(ctx context.Context, app *gin.Engine, cfg *config.Config,
	registries []*registry, metricsRegistry *prometheus.Registry) *Router {

	r := NewRouter(app, cfg, registries)
	r.ctx = ctx // Set the signal-aware context
	r.RegisterRoutes()

	// Configure max connections
	r.MaxConnections = defaultMaxConnections
	if cfg.Exists("server.maxConnections") {
		r.MaxConnections = cfg.GetInt64("server.maxConnections")
	}

	// Initialize connection metrics
	InitConnectionMetrics(metricsRegistry)
	SetMaxConnections(r.MaxConnections)

	// Configure server address
	if cfg.Exists("server.addr") {
		r.Addr = cfg.GetString("server.addr")
	} else {
		r.Addr = ":8080"
	}

	// Configure timeouts
	r.ReadTimeout = 90 * time.Second
	r.WriteTimeout = 90 * time.Second
	r.IdleTimeout = 90 * time.Second
	r.ReadHeaderTimeout = 90 * time.Second

	return r
}

// ============================================================================
// MAIN SERVER INITIALIZATION FUNCTION
// ============================================================================

// func Defaultgin(cfg *config.Config, osdktrace *otelsdktrace.TracerProvider, MetricsRegistry *prometheus.Registry, Checker *healthcheck.Checker) *Router {
func Defaultgin(ctx context.Context, cfg *config.Config, osdktrace *otelsdktrace.TracerProvider, MetricsRegistry *prometheus.Registry, registries []*registry) *Router {
	// Configure Gin mode based on environment
	configureGinMode(cfg)

	// Create Gin engine
	// Note: Custom JSON binding with goccy/go-json is set up automatically
	// via init() function in api-server/route/route_improved.go
	app := gin.New()

	// Register middlewares in order
	registerCoreMiddlewares(app, cfg, MetricsRegistry)
	registerSecurityMiddlewares(app, cfg)
	registerObservabilityMiddlewares(app, cfg, osdktrace, MetricsRegistry)

	// Register global routes: healthz, NoRoute, NoMethod
	Setup(app)

	// Register debug and monitoring endpoints
	registerDebugEndpoints(app, cfg, MetricsRegistry)

	// Create and configure router with timeouts and connection limits
	return createAndConfigureRouter(ctx, app, cfg, registries, MetricsRegistry)
}

var isShuttingDown atomic.Value

func init() {
	isShuttingDown.Store(false)
}

// SetIsShuttingDown is an exported function that allows other packages to update the isShuttingDown value
func SetIsShuttingDown(shuttingDown bool) {
	isShuttingDown.Store(shuttingDown)
}

func HealthCheckHandler(c *gin.Context) {
	shuttingDown := isShuttingDown.Load().(bool)
	if shuttingDown {
		// If the server is shutting down, respond with Service Unavailable
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy"})
		return
	}
	// If the server is not shutting down, respond with OK
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

// Setup registers global error handlers and health endpoint on the provided Gin router.
func Setup(router *gin.Engine) {
	router.NoRoute(func(c *gin.Context) {
		apierrors.HandleNoRouteError(c)
	})

	router.NoMethod(func(c *gin.Context) {
		apierrors.HandleNoMethodError(c)
	})

	router.GET("/healthz", HealthCheckHandler)
}
