package route

import (
	"bytes"
	"errors"
	"io"

	apierrors "MgApplication/api-errors"
	log "MgApplication/api-log"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// ============================================================================
// CONTENT-TYPE SPECIFIC BINDERS FOR CLEANER CODE
// ============================================================================

const (
	// Maximum size for text/plain body reading (10MB)
	maxPlainTextBodySize = 10 * 1024 * 1024
)

// bindJSON binds JSON request body to the request struct
func bindJSON[Req any](c *gin.Context, ctx *Context, req *Req) error {
	if err := c.ShouldBindJSON(req); err != nil {
		log.Error(ctx.Ctx, "JSON bind failed: %v", err)
		apierrors.HandleBindingError(c, err)
		return err
	}
	return nil
}

// bindXML binds XML request body to the request struct
func bindXML[Req any](c *gin.Context, ctx *Context, req *Req) error {
	// Handle special case for *bytes.Buffer
	if reqbuf, ok := any(req).(*bytes.Buffer); ok {
		_, err := io.Copy(reqbuf, c.Request.Body)
		if err != nil {
			log.Error(ctx.Ctx, "XML buffer copy failed: %v", err)
			apierrors.HandleBindingError(c, err)
			return err
		}
		*req = any(reqbuf).(Req)
		return nil
	}

	if err := c.ShouldBindXML(req); err != nil {
		log.Error(ctx.Ctx, "XML bind failed: %v", err)
		apierrors.HandleBindingError(c, err)
		return err
	}
	return nil
}

// bindForm binds form-urlencoded request body to the request struct
func bindForm[Req any](c *gin.Context, ctx *Context, req *Req) error {
	if err := c.ShouldBind(req); err != nil {
		log.Debug(ctx.Ctx, "Form bind failed: %v", err)
		apierrors.HandleBindingError(c, err)
		return err
	}
	return nil
}

// bindPlainText binds text/plain request body to string or []byte
func bindPlainText[Req any](c *gin.Context, ctx *Context, req *Req) error {
	// Use pooled buffer for reading
	buf := getBuffer()
	defer putBuffer(buf)

	// Limit the read to prevent DoS
	limited := io.LimitReader(c.Request.Body, maxPlainTextBodySize)

	_, err := buf.ReadFrom(limited)
	if err != nil {
		log.Error(ctx.Ctx, "text/plain read failed: %v", err)
		apierrors.HandleBindingError(c, err)
		return err
	}

	data := buf.Bytes()

	// Type assertion to determine target type
	switch v := any(req).(type) {
	case *string:
		*v = string(data)
	case *[]byte:
		// Make a copy since we're putting the buffer back to the pool
		dst := make([]byte, len(data))
		copy(dst, data)
		*v = dst
	default:
		err := errors.New("text/plain body can bind only into string or []byte request type")
		log.Error(ctx.Ctx, "text/plain bind failed: %v", err)
		apierrors.HandleBindingError(c, err)
		return err
	}

	return nil
}

// bindMultipartForm binds multipart/form-data request body to the request struct
func bindMultipartForm[Req any](c *gin.Context, ctx *Context, req *Req) error {
	if err := c.ShouldBind(req); err != nil {
		log.Debug(ctx.Ctx, "Multipart form bind failed: %v", err)
		apierrors.HandleBindingError(c, err)
		return err
	}
	return nil
}

// bindYAML binds YAML request body to the request struct
func bindYAML[Req any](c *gin.Context, ctx *Context, req *Req) error {
	if err := c.ShouldBindBodyWith(req, binding.YAML); err != nil {
		log.Error(ctx.Ctx, "YAML bind failed: %v", err)
		apierrors.HandleBindingError(c, err)
		return err
	}
	return nil
}
