package route

import (
	"fmt"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	apierrors "MgApplication/api-errors"
	"MgApplication/api-server/util/diutil/typlect"
	validation "MgApplication/api-validation"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/goccy/go-json"
)

// ============================================================================
// CUSTOM JSON BINDING WITH GOCCY/GO-JSON
// ============================================================================

// CustomJSONBinding uses goccy/go-json for high-performance JSON encoding/decoding
// goccy/go-json provides 2-8x better performance compared to encoding/json in this environment
// while maintaining 100% compatibility. See API_SERVER_JSON_PERFORMANCE.md for detailed benchmarks.
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

// init sets up goccy/go-json as the default JSON binding for Gin
// This runs automatically when the route package is imported
func init() {
	binding.JSON = CustomJSONBinding{}
}

// ============================================================================
// IMPROVED BUILD FUNCTION WITH SYNC.POOL OPTIMIZATIONS
// ============================================================================

// buildImproved is an optimized version of build with sync.Pool usage
// and cleaner code organization
func buildImproved[Req, Res any](f HandlerFunc[Req, Res], defaultStatus ...int) gin.HandlerFunc {
	ds := http.StatusOK
	if len(defaultStatus) == 1 {
		ds = defaultStatus[0]
	}

	hasInput := typlect.GetType[Req]() != typlect.TypeNoParam

	return func(c *gin.Context) {
		// Get pooled context to reduce allocations
		ctx := getContext()
		defer func() {
			if ctx.cancel != nil {
				ctx.cancel()
			}
			putContext(ctx)
		}()

		ctx.fromGinCtx(c)

		var req Req
		startDeserialize := time.Now()

		if hasInput {
			// Bind URI parameters
			if len(c.Params) > 0 {
				if err := c.ShouldBindUri(&req); err != nil {
					apierrors.HandleBindingError(c, err)
					return
				}
			}

			// Bind query parameters
			if len(c.Request.URL.Query()) > 0 {
				if err := c.ShouldBindQuery(&req); err != nil {
					apierrors.HandleBindingError(c, err)
					return
				}
			}

			// Bind body based on content type
			if c.Request.Method != http.MethodGet && c.Request.Body != nil {
				if err := bindRequestBody(c, ctx, &req); err != nil {
					return // Error already handled by bindRequestBody
				}
			}

			// Validate the bound request
			if err := validation.ValidateStruct(req); err != nil {
				apierrors.HandleValidationError(c, err)
				return
			}
		}

		// Add Server-Timing header using pooled string builder
		setServerTimingHeader(c, startDeserialize)

		// Execute handler
		res, err := f(ctx, req)
		if err != nil {
			c.Error(err)
			return
		}

		// Handle response
		handleResponse(c, res, ds)
	}
}

// bindRequestBody binds the request body based on Content-Type header
func bindRequestBody[Req any](c *gin.Context, ctx *Context, req *Req) error {
	rawCT := c.GetHeader("Content-Type")

	// Require Content-Type header
	if rawCT == "" {
		msg := "Content-Type header is required for requests with body"
		codeAndMsg := apierrors.NewHTTPStatsuCodeAndMessage(http.StatusBadRequest, msg)
		apierrors.ErrorResponseWithStatusCodeAndMessage(c, codeAndMsg, msg, nil)
		return fmt.Errorf("missing content-type")
	}

	// Parse media type
	mediaType, _, err := mime.ParseMediaType(rawCT)
	if err != nil {
		msg := "Invalid Content-Type header"
		codeAndMsg := apierrors.NewHTTPStatsuCodeAndMessage(http.StatusBadRequest, msg)
		apierrors.ErrorResponseWithStatusCodeAndMessage(c, codeAndMsg,
			fmt.Sprintf("Failed to parse Content-Type '%s': %v", rawCT, err), err)
		return err
	}

	// Normalize to lowercase for comparison
	mediaType = strings.ToLower(mediaType)

	// Route to appropriate binder
	return routeToContentTypeBinder(c, ctx, req, mediaType)
}

// routeToContentTypeBinder routes request to appropriate content-type handler
func routeToContentTypeBinder[Req any](c *gin.Context, ctx *Context, req *Req, mediaType string) error {
	// Handle multipart first (prefix check)
	if isMultipartFormData(mediaType) {
		return bindMultipartForm(c, ctx, req)
	}

	// Handle YAML variants
	if isYAMLMediaType(mediaType) {
		return bindYAML(c, ctx, req)
	}

	// Use optimized map lookup for exact matches
	switch mediaType {
	case mediaTypeJSON:
		return bindJSON(c, ctx, req)
	case mediaTypeXML, mediaTypeTextXML:
		return bindXML(c, ctx, req)
	case mediaTypeForm:
		return bindForm(c, ctx, req)
	case mediaTypePlainText:
		return bindPlainText(c, ctx, req)
	default:
		// Unsupported content type
		supported := []string{
			mediaTypeJSON,
			mediaTypeXML,
			mediaTypeForm,
			mediaTypeMultipart,
			mediaTypePlainText,
			mediaTypeYAML,
		}
		err := fmt.Errorf("unsupported content type '%s'. Supported: %s",
			mediaType, strings.Join(supported, ", "))
		msg := "Unsupported content type"
		codeAndMsg := apierrors.NewHTTPStatsuCodeAndMessage(http.StatusUnsupportedMediaType, msg)
		apierrors.ErrorResponseWithStatusCodeAndMessage(c, codeAndMsg, err.Error(), err)
		return err
	}
}

// setServerTimingHeader adds Server-Timing header using pooled string builder
func setServerTimingHeader(c *gin.Context, startTime time.Time) {
	duration := time.Since(startTime)

	// Use pooled string builder for efficient string construction
	sb := getStringBuilder()
	defer putStringBuilder(sb)

	// Check if there's an existing Server-Timing header
	if existing := c.Writer.Header().Get("Server-Timing"); existing != "" {
		sb.WriteString(existing)
		sb.WriteString(", ")
	}

	// Add deserialization timing (using strconv to avoid allocation)
	sb.WriteString("deserialize;dur=")
	sb.WriteString(strconv.FormatInt(duration.Milliseconds(), 10))

	c.Writer.Header().Set("Server-Timing", sb.String())
}

// ============================================================================
// ORIGINAL BUILD FUNCTION (KEPT FOR BACKWARD COMPATIBILITY)
// ============================================================================
// The original build function remains unchanged and can be used if needed.
// To switch to the improved version, replace `build` with `buildImproved`
// in the New() function in route.go
