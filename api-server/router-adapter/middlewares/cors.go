package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"MgApplication/api-server/router-adapter"
)

// CORSConfig configures CORS middleware
type CORSConfig struct {
	// AllowOrigins is a list of allowed origins
	// Use "*" to allow all origins
	AllowOrigins []string

	// AllowMethods is a list of allowed HTTP methods
	// Default: GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS
	AllowMethods []string

	// AllowHeaders is a list of allowed headers
	// Default: Origin, Content-Type, Accept, Authorization
	AllowHeaders []string

	// ExposeHeaders is a list of headers exposed to the client
	ExposeHeaders []string

	// AllowCredentials indicates whether credentials are allowed
	// Default: false
	AllowCredentials bool

	// MaxAge indicates how long preflight results can be cached (in seconds)
	// Default: 86400 (24 hours)
	MaxAge int
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
		},
		ExposeHeaders:    []string{},
		AllowCredentials: false,
		MaxAge:           86400, // 24 hours
	}
}

// CORS returns a middleware that handles CORS
func CORS(config ...CORSConfig) routeradapter.MiddlewareFunc {
	cfg := DefaultCORSConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	// Pre-join methods and headers for performance
	allowMethods := strings.Join(cfg.AllowMethods, ", ")
	allowHeaders := strings.Join(cfg.AllowHeaders, ", ")
	exposeHeaders := strings.Join(cfg.ExposeHeaders, ", ")

	return func(ctx *routeradapter.RouterContext, next func() error) error {
		origin := ctx.Header("Origin")

		// Check if origin is allowed
		originAllowed := false
		for _, allowed := range cfg.AllowOrigins {
			if allowed == "*" {
				ctx.SetHeader("Access-Control-Allow-Origin", "*")
				originAllowed = true
				break
			}
			if origin == allowed {
				ctx.SetHeader("Access-Control-Allow-Origin", origin)
				if cfg.AllowCredentials {
					ctx.SetHeader("Access-Control-Allow-Credentials", "true")
				}
				originAllowed = true
				break
			}
		}

		// Only set CORS headers if origin is allowed
		if originAllowed {
			ctx.SetHeader("Access-Control-Allow-Methods", allowMethods)
			ctx.SetHeader("Access-Control-Allow-Headers", allowHeaders)

			if len(exposeHeaders) > 0 {
				ctx.SetHeader("Access-Control-Expose-Headers", exposeHeaders)
			}

			if cfg.MaxAge > 0 {
				ctx.SetHeader("Access-Control-Max-Age", formatInt(cfg.MaxAge))
			}
		}

		// Handle preflight OPTIONS request
		if ctx.Request.Method == http.MethodOptions {
			if originAllowed {
				return ctx.NoContent(http.StatusNoContent)
			}
			return ctx.NoContent(http.StatusForbidden)
		}

		// Continue to next middleware/handler
		return next()
	}
}

// formatInt formats an integer to string
func formatInt(i int) string {
	return fmt.Sprintf("%d", i)
}
