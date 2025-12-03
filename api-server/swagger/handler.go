package swagger

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"MgApplication/api-server/common"
	"MgApplication/api-server/swagger/files"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
)

// Cache for swagger JSON responses per host to avoid repeated marshaling
var (
	swaggerCache     sync.Map // map[string][]byte (host -> marshaled JSON)
	swaggerCacheLock sync.RWMutex
)

func ginWrapper(v3Doc *openapi3.T) common.GinAppWrapper {
	return func(r *gin.Engine) *gin.Engine {
		r.Use(
			gin.WrapH(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				redirectRules := map[string]string{
					"/":                        "/swagger/index.html",
					"/swagger":                 "/swagger/index.html",
					"/swagger.json":            "/docs/resolved_swagger.json",
					"/swagger/v1/swagger.json": "/swagger/docs.json",
				}

				if newPath, ok := redirectRules[req.URL.Path]; ok {
					http.Redirect(w, req, newPath, http.StatusMovedPermanently)
					return
				}

				r.ServeHTTP(w, req)
			})),
			gin.WrapH(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Path == "/swagger/docs.json" || req.URL.Path == "/swagger/docs.json/" {
					v3Doc = attachHostToV3Doc(v3Doc, req.Host)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(v3Doc)
					return
				}

				if strings.HasPrefix(req.URL.Path, "/swagger") {
					trimmedPath := strings.TrimPrefix(req.URL.Path, "/swagger")
					req.URL.Path = trimmedPath
					fsHandler := http.StripPrefix("/swagger", http.FileServer(http.FS(files.Files)))
					fsHandler.ServeHTTP(w, req)
					return
				}

				r.ServeHTTP(w, req)
			})),
		)
		return r
	}
}

func newRedirectMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		redirectRules := map[string]string{
			"/":                        "/swagger/index.html",
			"/swagger":                 "/swagger/index.html",
			"/swagger.json":            "/swagger/docs.json",
			"/swagger/v1/swagger.json": "/swagger/docs.json",
		}

		if newPath, ok := redirectRules[c.Request.URL.Path]; ok {
			c.Redirect(http.StatusMovedPermanently, newPath)
			c.Abort()
			return
		}

		c.Next()
	}
}

// Attach host to the OpenAPI document
func attachHostToV3Doc(doc *openapi3.T, host string) *openapi3.T {
	doc.Servers = []*openapi3.Server{
		{
			URL: fmt.Sprintf("http://%s", host),
		},
	}
	return doc
}

// Middleware function conversion for serving Swagger files
// Optimized with caching to avoid repeated JSON marshaling
func newMiddleware(v3Doc *openapi3.T) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/swagger/docs.json" || c.Request.URL.Path == "/swagger/docs.json/" {
			host := c.Request.Host

			// Check cache first
			if cached, ok := swaggerCache.Load(host); ok {
				c.Data(http.StatusOK, "application/json", cached.([]byte))
				return
			}

			// Cache miss - generate and cache
			docCopy := *v3Doc // Shallow copy to avoid mutating original
			docCopy.Servers = []*openapi3.Server{
				{URL: fmt.Sprintf("http://%s", host)},
			}

			data, err := json.Marshal(&docCopy)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal swagger"})
				return
			}

			// Store in cache
			swaggerCache.Store(host, data)

			c.Data(http.StatusOK, "application/json", data)
			return
		}

		if strings.HasPrefix(c.Request.URL.Path, "/swagger") {
			// Trim the prefix and serve the static files
			trimmedPath := strings.TrimPrefix(c.Request.URL.Path, "/swagger")
			c.Request.URL.Path = trimmedPath
			fsHandler := http.StripPrefix("/swagger", http.FileServer(http.FS(files.Files)))
			fsHandler.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}

		c.Next()
	}
}
