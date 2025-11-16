package router

import (
	"context"
	"embed"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	config "MgApplication/api-config"
	apierrors "MgApplication/api-errors"

	"github.com/arl/statsviz"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/goccy/go-json"
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
		currentConns := atomic.LoadInt64(&activeConnections)
		if currentConns >= s.MaxConnections {
			log.GetBaseLoggerInstance().ToZerolog().Warn().
				Int64("activeConnections", currentConns).
				Int64("maxConnections", s.MaxConnections).
				Msg("Max connections reached")
			conn.Close()
			return
		}
		atomic.AddInt64(&activeConnections, 1)
	case http.StateClosed, http.StateHijacked:
		atomic.AddInt64(&activeConnections, -1)
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

type CustomJSONBinding struct{}

func (CustomJSONBinding) Name() string {
	return "json"
}
func (CustomJSONBinding) Bind(req *http.Request, obj interface{}) error {
	if req == nil || req.Body == nil {
		return fmt.Errorf("missing request body")
	}
	decoder := json.NewDecoder(req.Body)
	return decoder.Decode(obj)
}

func (CustomJSONBinding) BindBody(body []byte, obj interface{}) error {
	return json.Unmarshal(body, obj)
}

// func Defaultgin(cfg *config.Config, osdktrace *otelsdktrace.TracerProvider, MetricsRegistry *prometheus.Registry, Checker *healthcheck.Checker) *Router {
func Defaultgin(cfg *config.Config, osdktrace *otelsdktrace.TracerProvider, MetricsRegistry *prometheus.Registry, registries []*registry) *Router {

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
	binding.JSON = CustomJSONBinding{}

	//add all no method, no path, no handler routes here
	app := gin.New()
	serverCfg, err := cfg.Of("server")
	if err != nil {
		log.Error(nil, "Error in getting server config")
	}

	corsCfg, err := serverCfg.Of("cors")
	if err != nil {
		log.Error(nil, "Error in getting cors config")
	}

	var defaultsizelimit int64

	defaultsizelimit = 2 * 1024 * 1024

	var sizelimit int64 = 0
	sizelimit = cfg.GetInt64("server.bodylimit")

	if sizelimit == 0 {
		sizelimit = int64(defaultsizelimit)
	}
	app.Use(
		middlewares.BodyLimiter(sizelimit),
		middlewares.BodyLimitErrorHandler())

	if cfg.Exists("server.ratelimit") {
		ratelimit := cfg.GetString("server.ratelimit")
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
	} else {
		globalBucket = rate.NewLeakyBucket(DefaultRate, DefaultCapacity)
	}
	app.Use(middlewares.RateMiddleware(globalBucket))
	ratelimiter.InitMetrics(globalBucket, MetricsRegistry)

	encryptenabled := false
	if cfg.Exists("server.encrypt") {
		encryptenabled = cfg.GetBool("server.encrypt")
	}

	app.Use(
		middlewares.CORSMiddleware(corsCfg),
		middlewares.Recover(cfg),
		//middlewares.CustomJSON(),

		middlewares.ErrorHandler(),

		//middlewares.TimeoutMiddleware(time.Second*50),
	)

	if encryptenabled {
		app.Use(middlewares.DecryptMiddleware())
		app.Use(middlewares.ResponseSignatureMiddleware())
	}

	if cfg.GetBool("trace.enabled") {
		// var token string

		// if cfg.Exists("trace.token") {
		// 	token = cfg.GetString("trace.token")
		// }

		app.Use(middlewares.RequestTracerMiddleware(
			cfg.AppName(),
			middlewares.RequestTracerMiddlewareConfig{
				TracerProvider: AnnotateTracerProvider(osdktrace),
			},
			// token,
		))
	}

	app.Use(middlewares.SetCtxLoggerMiddleware(),
		middlewares.RequestResponseLoggerMiddleware())

	// Register global routes: healthz, NoRoute, NoMethod
	Setup(app)

	if cfg.GetBool("metrics.collect.routes") {
		var buckets []float64
		if bucketsConfig := cfg.GetString("metrics.buckets"); bucketsConfig != "" {
			for _, s := range Split(bucketsConfig) {
				f, err := strconv.ParseFloat(s, 64)
				if err == nil {
					buckets = append(buckets, f)
				}
			}
		}

		metricsMiddlewareConfig := middlewares.RequestMetricsMiddlewareConfig{
			Registry: MetricsRegistry,
			// Namespace:               Sanitize(cfg.GetString("appname")),
			Namespace:               "",
			Subsystem:               Sanitize("router"),
			Buckets:                 buckets,
			NormalizeRequestPath:    true,
			NormalizeResponseStatus: true,
		}

		app.Use(middlewares.RequestMetricsMiddlewareWithConfig(metricsMiddlewareConfig))
	}

	dashboardEnabled := cfg.GetBool("server.dashboard.enabled")
	pprofExpose := cfg.GetBool("server.debug.pprof.expose")
	//startupExpose := cfg.GetBool("server.healthcheck.expose")
	metricsExpose := cfg.GetBool("metrics.expose")
	statsExpose := cfg.GetBool("server.debug.stats.expose")
	metricsPath := DefaultMetricsPath
	pprofPath := cfg.GetString("server.debug.pprof.path")
	statsPath := cfg.GetString("server.debug.stats.path")
	//startupPath := cfg.GetString("server.healthcheck.path")

	if metricsPath == "" {
		metricsPath = "/metrics"
	}

	if metricsExpose {
		app.GET(metricsPath, gin.WrapH(promhttp.HandlerFor(MetricsRegistry, promhttp.HandlerOpts{})))
		log.GetBaseLoggerInstance().ToZerolog().Debug().Msg("registered metrics handler")
	}

	if pprofExpose {
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

	if dashboardEnabled {
		renderer, err := NewDashboardRenderer(templatesFS, "templates/dashboard.html")
		if err != nil {
			panic(err)
		}
		// theme
		app.POST("/theme", func(c *gin.Context) {
			themeCookie := &http.Cookie{
				Name: "theme",
			}

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

		// render
		app.GET("/dashboard", func(c *gin.Context) {
			var theme string
			themeCookie, err := c.Cookie("theme")
			if err == nil {
				switch themeCookie {
				case ThemeDark:
					theme = ThemeDark
				case ThemeLight:
					theme = ThemeLight
				default:
					theme = ThemeLight
				}
			} else {
				theme = ThemeLight
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

	if statsExpose {
		if statsPath == "" {
			statsPath = DefaultDebugStatsPath
		}

		srv, err := statsviz.NewServer()
		if err != nil {
			panic(err)
		}

		debug := app.Group(statsPath)
		{
			debug.GET("/*filepath", func(c *gin.Context) {
				if c.Param("filepath") == "/ws" {
					srv.Ws()(c.Writer, c.Request)
					return
				}
				srv.Index()(c.Writer, c.Request)
			})
		}

	}

	// if startupExpose {
	// 	if startupPath == "" {
	// 		startupPath = DefaultHealthCheckStartupPath
	// 	}

	// 	// app.GET(startupPath,
	// 	// 	func(ctx *gin.Context) {
	// 	// 		 results := Checker.Check(ctx.Request.Context(), healthcheck.Startup)
	// 	// 		 overallHealthy := true
	// 	//     healthStatus := gin.H{}
	// 	// 	for name, result := range results.ProbesResults {
	// 	// 		if !result.Success {
	// 	// 			overallHealthy = false
	// 	// 		}
	// 	// 		healthStatus[name] = gin.H{
	// 	//         "ok":      result.Success,
	// 	//         "details": result.Message,
	// 	//     }

	// 	// 	}
	// 	// 	if overallHealthy {
	// 	// 		ctx.JSON(http.StatusOK, gin.H{
	// 	// 			"status": "healthy",
	// 	// 			"probes": healthStatus,
	// 	// 		})
	// 	// 	} else {
	// 	// 		ctx.JSON(http.StatusInternalServerError, gin.H{
	// 	// 			"status": "unhealthy",
	// 	// 			"probes": healthStatus,
	// 	// 		})
	// 	// 	}

	// 	// },
	// 	// )

	// //	app.GET(startupPath, health.MultipleHealthCheckHandler(Checker, healthcheck.Startup))
	// 	log.GetBaseLoggerInstance().ToZerolog().Debug().Msg("registered healthcheck startup handler")
	// }

	r := NewRouter(app, cfg, registries)
	r.RegisterRoutes()
	r.MaxConnections = defaultMaxConnections
	if r.cfg.Exists("server.maxConnections") {
		r.MaxConnections = r.cfg.GetInt64("server.maxConnections")
	}
	if r.cfg.Exists("server.addr") {
		r.Addr = r.cfg.GetString("server.addr")
	} else {
		r.Addr = ":8080"
	}

	// if r.cfg.Exists("server.readTimeout") {
	// 	r.ReadTimeout = r.cfg.GetDuration("server.readTimeout")
	// } else {
	// 	r.ReadTimeout = 90 * time.Second
	// }

	// if r.cfg.Exists("server.writeTimeout") {
	// 	r.WriteTimeout = r.cfg.GetDuration("server.writeTimeout")
	// } else {
	// 	r.WriteTimeout = 90 * time.Second
	// }

	r.ReadTimeout = 90 * time.Second
	r.WriteTimeout = 90 * time.Second
	r.IdleTimeout = 90 * time.Second
	r.ReadHeaderTimeout = 90 * time.Second
	// if r.cfg.Exists("server.idleTimeout") {
	// 	r.IdleTimeout = r.cfg.GetDuration("server.idleTimeout")
	// } else {
	// 	r.IdleTimeout = 90 * time.Second
	// }

	// if r.cfg.Exists("server.readHeaderTimeout") {
	// 	r.IdleTimeout = r.cfg.GetDuration("server.readHeaderTimeout")
	// } else {
	// 	r.IdleTimeout = 30 * time.Second
	// }

	return r

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
