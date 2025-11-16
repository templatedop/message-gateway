package route

import (
	"bytes"
	"strings"
	"sync"
)

// ============================================================================
// SYNC.POOL IMPLEMENTATIONS FOR PERFORMANCE OPTIMIZATION
// ============================================================================

// contextPool is a sync.Pool for Context objects to reduce allocations
var contextPool = sync.Pool{
	New: func() interface{} {
		return &Context{}
	},
}

// getContext retrieves a Context from the pool
func getContext() *Context {
	return contextPool.Get().(*Context)
}

// putContext returns a Context to the pool for reuse
func putContext(ctx *Context) {
	// Reset context fields before returning to pool
	ctx.Ctx = nil
	ctx.cancel = nil
	ctx.Log = nil
	contextPool.Put(ctx)
}

// bufferPool is a sync.Pool for byte buffers to reduce allocations
var bufferPool = sync.Pool{
	New: func() interface{} {
		// Pre-allocate 4KB buffer (typical request size)
		return bytes.NewBuffer(make([]byte, 0, 4096))
	},
}

// getBuffer retrieves a buffer from the pool
func getBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

// putBuffer returns a buffer to the pool for reuse
func putBuffer(buf *bytes.Buffer) {
	// Reset buffer before returning to pool
	buf.Reset()
	// Only return buffers with reasonable capacity to pool (avoid memory leak)
	if buf.Cap() <= 65536 { // 64KB max
		bufferPool.Put(buf)
	}
}

// stringBuilderPool is a sync.Pool for strings.Builder to reduce allocations
var stringBuilderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

// getStringBuilder retrieves a strings.Builder from the pool
func getStringBuilder() *strings.Builder {
	return stringBuilderPool.Get().(*strings.Builder)
}

// putStringBuilder returns a strings.Builder to the pool for reuse
func putStringBuilder(sb *strings.Builder) {
	// Reset before returning to pool
	sb.Reset()
	stringBuilderPool.Put(sb)
}

// ============================================================================
// MEDIA TYPE CACHE FOR FASTER LOOKUPS
// ============================================================================

// Common media types pre-computed in lowercase
const (
	mediaTypeJSON      = "application/json"
	mediaTypeXML       = "application/xml"
	mediaTypeTextXML   = "text/xml"
	mediaTypeForm      = "application/x-www-form-urlencoded"
	mediaTypeMultipart = "multipart/form-data"
	mediaTypePlainText = "text/plain"
	mediaTypeYAML      = "application/yaml"
	mediaTypeXYAML     = "application/x-yaml"
	mediaTypeTextYAML  = "text/yaml"
)

// mediaTypeHandlers maps media types to their handling priority
// This allows O(1) lookup instead of sequential switch cases
var mediaTypeHandlers = map[string]int{
	mediaTypeJSON:      1,
	mediaTypeXML:       2,
	mediaTypeTextXML:   2,
	mediaTypeForm:      3,
	mediaTypePlainText: 4,
	mediaTypeYAML:      5,
	mediaTypeXYAML:     5,
	mediaTypeTextYAML:  5,
}

// isMultipartFormData checks if media type is multipart/form-data
func isMultipartFormData(mediaType string) bool {
	return strings.HasPrefix(mediaType, mediaTypeMultipart)
}

// isYAMLMediaType checks if media type is any YAML variant
func isYAMLMediaType(mediaType string) bool {
	return mediaType == mediaTypeYAML ||
		mediaType == mediaTypeXYAML ||
		strings.HasPrefix(mediaType, mediaTypeTextYAML)
}

// ============================================================================
// POOL STATISTICS (FOR MONITORING)
// ============================================================================

// PoolStats provides statistics about pool usage
type PoolStats struct {
	ContextGets       int64
	ContextPuts       int64
	BufferGets        int64
	BufferPuts        int64
	StringBuilderGets int64
	StringBuilderPuts int64
}

var poolStats PoolStats
var poolStatsMutex sync.RWMutex

// GetPoolStats returns current pool statistics
func GetPoolStats() PoolStats {
	poolStatsMutex.RLock()
	defer poolStatsMutex.RUnlock()
	return poolStats
}

// ResetPoolStats resets pool statistics
func ResetPoolStats() {
	poolStatsMutex.Lock()
	defer poolStatsMutex.Unlock()
	poolStats = PoolStats{}
}
