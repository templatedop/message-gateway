package route

import (
	// validation "MgApplication/api-validation" - removed, using govalid in handlers
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"MgApplication/api-server/util/diutil/typlect"
	// perror "apilocalgin/errors"
	// "apilocalgin/errors/ecode"
	apierrors "MgApplication/api-errors"

	log "MgApplication/api-log"
	"MgApplication/api-server/response"
)

type Context struct {
	Ctx    context.Context
	cancel context.CancelFunc
	Log    *log.Logger
}

// Deadline implements context.Context by delegating to the underlying context.
func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.Ctx.Deadline()
}

// Done implements context.Context by delegating to the underlying context.
func (c *Context) Done() <-chan struct{} {
	return c.Ctx.Done()
}

// Err implements context.Context by delegating to the underlying context.
func (c *Context) Err() error {
	return c.Ctx.Err()
}

// Value implements context.Context by delegating to the underlying context.
func (c *Context) Value(key any) any {
	return c.Ctx.Value(key)
}

func (c *Context) fromGinCtx(ginCtx *gin.Context) {
	cc := ginCtx.Request.Context()

	// Use the existing request context which may already have timeout, tracing, etc.
	// from middlewares. Don't create a new timeout context here to avoid conflicts.
	// If timeout middleware is enabled, it already set up the timeout.
	// We just need a cancel context to allow early cancellation if needed.
	ctx, cancel := context.WithCancel(cc)
	c.Ctx = ctx
	c.cancel = cancel

	// Don't replace ginCtx.Request.Context() as it may contain values from middlewares
	// The handler will use c.Ctx for operations and ginCtx.Request.Context() is already set
}

type NoParam = typlect.NoParam

type HandlerFunc[Req, Res any] func(*Context, Req) (Res, error)

type Route interface {
	Meta() Meta
	Desc(s string) Route
	Name(s string) Route
	AddMiddlewares(mws ...gin.HandlerFunc) Route
}

type route[Req, Res any] struct {
	meta Meta
}

func New[Req, Res any](method, path string, h HandlerFunc[Req, Res], ds ...int) Route {
	return newRoute[Req, Res](method, path, buildImproved(h, ds...))
}

func newRoute[Req, Res any](method, path string, h gin.HandlerFunc) Route {
	return &route[Req, Res]{
		meta: Meta{
			Method: method,
			Path:   path,
			Func:   h,
			Req:    typlect.GetType[Req](),
			Res:    typlect.GetType[Res](),
		},
	}
}

func (h *route[Req, Res]) AddMiddlewares(mws ...gin.HandlerFunc) Route {
	h.meta.Middlewares = append(h.meta.Middlewares, mws...)
	return h
}

func (h *route[Req, Res]) Meta() Meta {
	return h.meta
}

func (h *route[Req, Res]) Desc(d string) Route {
	h.meta.Desc = d
	return h
}

func (h *route[Req, Res]) Name(d string) Route {
	h.meta.Name = d
	return h
}

// FileConsumer optionally implemented by request DTOs that want direct access to file headers.
type FileConsumer interface {
	AcceptFiles(map[string][]*multipart.FileHeader) error
}

// build is the legacy request handler builder kept for backward compatibility.
// The production code now uses buildImproved (see route_improved.go) which includes
// sync.Pool optimizations for better performance and reduced GC pressure.
func build[Req, Res any](f HandlerFunc[Req, Res], defaultStatus ...int) gin.HandlerFunc {
	ds := http.StatusOK
	if len(defaultStatus) == 1 {
		ds = defaultStatus[0]
	}

	hasInput := typlect.GetType[Req]() != typlect.TypeNoParam

	return func(c *gin.Context) {
		ctx := &Context{}
		ctx.fromGinCtx(c)
		defer ctx.cancel()

		var req Req
		startDeserialize := time.Now()

		if hasInput {
			// Path params
			if len(c.Params) > 0 {
				if err := c.ShouldBindUri(&req); err != nil {
					apierrors.HandleBindingError(c, err)
					return
				}
			}
			// Query params
			if len(c.Request.URL.Query()) > 0 {
				if err := c.ShouldBindQuery(&req); err != nil {
					apierrors.HandleBindingError(c, err)
					return
				}
			}

			// Body (non-GET usually has body) – content-type aware similar to provided generic body() example
			if c.Request.Method != http.MethodGet && c.Request.Body != nil {
				rawCT := c.GetHeader("Content-Type")

				// Require Content-Type header for non-GET requests with body
				if rawCT == "" {
					msg := "Content-Type header is required for requests with body"
					codeAndMsg := apierrors.NewHTTPStatsuCodeAndMessage(http.StatusBadRequest, msg)
					apierrors.ErrorResponseWithStatusCodeAndMessage(c, codeAndMsg, msg, nil)
					return
				}

				mediaType, _, err := mime.ParseMediaType(rawCT)
				if err != nil {
					msg := "Invalid Content-Type header"
					codeAndMsg := apierrors.NewHTTPStatsuCodeAndMessage(http.StatusBadRequest, msg)
					apierrors.ErrorResponseWithStatusCodeAndMessage(c, codeAndMsg,
						fmt.Sprintf("Failed to parse Content-Type '%s': %v", rawCT, err), err)
					return
				}
				mediaType = strings.ToLower(mediaType)
				if mediaType != "" { // proceed only if client declared a content type
					switch {
					case mediaType == "application/json":
						if err := c.ShouldBindJSON(&req); err != nil {
							log.Error(ctx.Ctx, "JSON bind failed: %v", err)
							apierrors.HandleBindingError(c, err)
							return
						}
					case mediaType == "application/xml" || mediaType == "text/xml":
						if reqbuf, ok := any(&req).(*bytes.Buffer); ok {
							_, err := io.Copy(reqbuf, c.Request.Body)
							if err != nil {
								log.Error(ctx, err)
								c.XML(500, gin.H{"error copying": err})
								return
							}

							req = any(reqbuf).(Req)
						}
						if err := c.ShouldBindXML(&req); err != nil {
							log.Error(ctx.Ctx, "XML bind failed: %v", err)
							apierrors.HandleBindingError(c, err)
							return
						}
					case mediaType == "application/x-www-form-urlencoded":
						if err := c.ShouldBind(&req); err != nil {
							log.Debug(ctx.Ctx, "Form bind failed: %v", err)
							apierrors.HandleBindingError(c, err)
							return
						}
					case mediaType == "text/plain":
						data, err := io.ReadAll(c.Request.Body)
						if err != nil {
							log.Error(ctx.Ctx, "text/plain read failed: %v", err)
							apierrors.HandleBindingError(c, err)
							return
						}
						switch v := any(&req).(type) {
						case *string:
							*v = string(data)
						case *[]byte:
							*v = data
						default:
							// unsupported target type for plain text
							err := errors.New("text/plain body can bind only into string or []byte request type")
							log.Error(ctx.Ctx, "text/plain bind failed: %v", err)
							apierrors.HandleBindingError(c, err)
							return
						}
					case strings.HasPrefix(mediaType, "multipart/form-data"):

						// Bind non-file form fields into request struct.
						if err := c.ShouldBind(&req); err != nil {
							log.Debug(ctx.Ctx, "Multipart form bind failed: %v", err)
							apierrors.HandleBindingError(c, err)
							return
						}

					case mediaType == "application/x-yaml" || mediaType == "application/yaml" || strings.HasPrefix(mediaType, "text/yaml"):
						// Attempt YAML binding via gin
						if err := c.ShouldBindBodyWith(&req, binding.YAML); err != nil {
							log.Error(ctx.Ctx, "YAML bind failed: %v", err)
							apierrors.HandleBindingError(c, err)
							return
						}
					default:
						// Provide a unified unsupported content type message listing allowed types
						supported := []string{"application/json", "application/xml", "application/x-www-form-urlencoded", "multipart/form-data", "text/plain", "application/yaml"}
						err := fmt.Errorf("unsupported content type '%s'. Supported: %s", mediaType, strings.Join(supported, ", "))
						msg := "Unsupported content type"
						codeAndMsg := apierrors.NewHTTPStatsuCodeAndMessage(http.StatusUnsupportedMediaType, msg)
						apierrors.ErrorResponseWithStatusCodeAndMessage(c, codeAndMsg, err.Error(), err)
						return
					}
				}

			}

			// Validation after all sources bound
			// Validation is now done in individual handlers using govalid
			// if err := validation.ValidateStruct(req); err != nil {
			// 	apierrors.HandleValidationError(c, err)
			// 	return
			// }
		}

		// Add simple timing header for deserialization (append if existing)
		duration := time.Since(startDeserialize)
		timingVal := fmt.Sprintf("deserialize;dur=%d", duration.Milliseconds())
		if existing := c.Writer.Header().Get("Server-Timing"); existing != "" {
			c.Writer.Header().Set("Server-Timing", existing+", "+timingVal)
		} else {
			c.Writer.Header().Set("Server-Timing", timingVal)
		}

		res, err := f(ctx, req)
		if err != nil {
			// apierrors.HandleCommonError(c, err)
			c.Error(err)
			return
		}
		handleResponse(c, res, ds)
	}
}

func handleResponse(c *gin.Context, res any, ds int) {
	if st, ok := any(res).(response.Stature); ok {
		status := st.Status()
		if status == 0 {
			if ds != 0 {
				status = ds
			} else {
				status = http.StatusOK
			}
		}
		responseType := st.ResponseType()

		if responseType == "file" {
			contentType := st.GetContentType()
			contentDisposition := st.GetContentDisposition()

			c.Writer.Header().Set("Content-Disposition", contentDisposition)
			c.Writer.Header().Set("Content-Type", contentType)

			// Stream if implementation provides a Stream method
			if streamer, ok2 := any(res).(response.Streamer); ok2 {
				if err := streamer.Stream(c.Writer); err != nil {
					log.Error(c.Request.Context(), "Failed to stream file response: %v", err)
					// Don't send body if headers already sent
					if !c.Writer.Written() {
						c.JSON(http.StatusInternalServerError, gin.H{
							"success": false,
							"message": "Failed to stream file",
						})
					}
					return
				}
				c.Status(status)
				return
			}
			// fallback to Data method
			c.Data(status, contentType, st.Object())
			return
		}

		// Standard JSON response
		c.JSON(status, res)
		return
	}

	// Response doesn't implement Stature interface - this should be rare
	// Log detailed warning to help identify which handler needs updating
	log.Warn(c.Request.Context(),
		"Response type %T does not implement Stature interface for %s %s - using default 200 OK. "+
			"Consider wrapping response in response.Response[T] for consistent API responses",
		res, c.Request.Method, c.Request.URL.Path)
	c.JSON(http.StatusOK, res)
}

func isStructEmpty(v interface{}) bool {
	val := reflect.ValueOf(v)
	for i := 0; i < val.NumField(); i++ {
		if !val.Field(i).IsZero() {
			return false
		}
	}
	return true
}

func extractFieldNameFromError(errorMessage string) string {
	errorMessage = strings.ReplaceAll(errorMessage, "\\n", "")
	errorMessage = strings.ReplaceAll(errorMessage, "\\t", "")
	errorMessage = strings.ReplaceAll(errorMessage, "\\", "")
	re := regexp.MustCompile(`Mismatch type (\w+) with value (\w+) "at index \d+: mismatched type with value"(\w+)":`)
	d := re.FindStringSubmatch(errorMessage)

	if len(d) == 0 {
		rm := regexp.MustCompile(`Mismatch type (\w+) with value (\w+)`)
		d = rm.FindStringSubmatch(errorMessage)
		if len(d) == 3 {
			expectedType := d[1]
			actualType := d[2]
			return fmt.Sprintf("One of the field expects is '%s' but sent '%s'", expectedType, actualType)
		}
	}

	if len(d) == 4 {
		expectedType := d[1]
		actualType := d[2]
		fieldName := d[3]
		return fmt.Sprintf("send %s for %s instead of %s", expectedType, fieldName, actualType)
	}

	return "unknown error format"
}
