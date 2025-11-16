package swagger

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"MgApplication/api-server/common"
	"MgApplication/api-server/swagger/files"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
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
func newMiddleware(v3Doc *openapi3.T) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/swagger/docs.json" || c.Request.URL.Path == "/swagger/docs.json/" {
			v3Doc = attachHostToV3Doc(v3Doc, c.Request.Host)
			c.JSON(http.StatusOK, v3Doc)
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
