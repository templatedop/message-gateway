package middlewares

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"MgApplication/api-server/router-adapter"

	"crypto/ed25519"

	"github.com/gibson042/canonicaljson-go"
	"golang.org/x/crypto/blake2b"
)

// SignatureConfig configures the signature middleware
type SignatureConfig struct {
	// PublicKey is the Ed25519 public key for verifying request signatures (base64 encoded)
	PublicKey string

	// PrivateKey is the Ed25519 private key for signing responses (base64 encoded)
	PrivateKey string

	// ValiditySeconds is the validity period for signatures (default: 3600 seconds)
	ValiditySeconds int64

	// SkipMethods are HTTP methods to skip signature verification (e.g., GET, HEAD)
	SkipMethods []string

	// SignatureHeader is the HTTP header name for the signature (default: "sig")
	SignatureHeader string
}

// DefaultSignatureConfig returns default signature configuration
func DefaultSignatureConfig() SignatureConfig {
	return SignatureConfig{
		ValiditySeconds: 3600,
		SkipMethods:     []string{"GET", "HEAD", "OPTIONS"},
		SignatureHeader: "sig",
	}
}

// SignaturePayload represents the signature header payload
type SignaturePayload struct {
	Timestamp string `json:"t"` // Unix timestamp
	Signature string `json:"s"` // Base64 encoded signature
	Expiry    string `json:"e"` // Unix timestamp expiry
}

// bodyBufPool pools byte buffers for request/response bodies
var bodyBufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// RequestSignatureVerifier returns a middleware that verifies Ed25519 signatures on incoming requests
// This middleware:
// - Reads the "sig" header containing base64 encoded signature payload
// - Verifies the signature using Ed25519 + Blake2b
// - Checks timestamp and expiry for replay protection
// - Uses canonical JSON for consistent hashing
func RequestSignatureVerifier(config ...SignatureConfig) routeradapter.MiddlewareFunc {
	cfg := DefaultSignatureConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	// Validate config
	if cfg.PublicKey == "" {
		panic("RequestSignatureVerifier: PublicKey is required")
	}

	return func(ctx *routeradapter.RouterContext, next func() error) error {
		// Skip signature verification for specified methods
		for _, method := range cfg.SkipMethods {
			if ctx.Request.Method == method {
				return next()
			}
		}

		// Get signature header
		sigHeader := ctx.Request.Header.Get(cfg.SignatureHeader)
		if sigHeader == "" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Missing signature header",
			})
		}

		// Buffer request body for signature verification
		buf := bodyBufPool.Get().(*bytes.Buffer)
		buf.Reset()
		defer bodyBufPool.Put(buf)

		// Read body into buffer
		if ctx.Request.Body != nil {
			tee := io.TeeReader(ctx.Request.Body, buf)
			_, err := io.Copy(io.Discard, tee)
			if err != nil {
				return ctx.JSON(http.StatusBadRequest, map[string]string{
					"error": "Failed to read request body",
				})
			}
			ctx.Request.Body.Close()
		}

		// Verify signature
		if !verifyJSONSignature(buf.Bytes(), sigHeader, cfg.PublicKey, cfg.ValiditySeconds) {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Invalid signature",
			})
		}

		// Restore body for handlers
		ctx.Request.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))

		return next()
	}
}

// ResponseSigner returns a middleware that signs JSON responses with Ed25519
// This middleware:
// - Captures the response body
// - Signs JSON responses using Ed25519 + Blake2b
// - Adds "sig" header with base64 encoded signature payload
// - Uses canonical JSON for consistent hashing
func ResponseSigner(config ...SignatureConfig) routeradapter.MiddlewareFunc {
	cfg := DefaultSignatureConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	// Validate config
	if cfg.PrivateKey == "" {
		panic("ResponseSigner: PrivateKey is required")
	}

	return func(ctx *routeradapter.RouterContext, next func() error) error {
		// Skip signing for specified methods
		for _, method := range cfg.SkipMethods {
			if ctx.Request.Method == method {
				return next()
			}
		}

		// Buffer to capture response body
		respBuf := bodyBufPool.Get().(*bytes.Buffer)
		respBuf.Reset()
		defer bodyBufPool.Put(respBuf)

		// Wrap response writer to capture body
		originalWriter := ctx.Response
		captureWriter := &bodyCaptureWriter{
			ResponseWriter: originalWriter,
			body:           respBuf,
		}
		ctx.Response = captureWriter

		// Execute next middleware/handler
		err := next()

		// Get status code
		statusCode := ctx.StatusCode()
		if statusCode == 0 {
			statusCode = http.StatusOK
		}

		// Only sign successful JSON responses
		if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
			// Write response without signing
			originalWriter.WriteHeader(statusCode)
			originalWriter.Write(respBuf.Bytes())
			return err
		}

		// Check if response is JSON
		contentType := captureWriter.Header().Get("Content-Type")
		if !strings.HasPrefix(contentType, "application/json") {
			// Write response without signing
			originalWriter.WriteHeader(statusCode)
			originalWriter.Write(respBuf.Bytes())
			return err
		}

		// Sign the response
		sigHeader, signErr := signJSON(respBuf.Bytes(), cfg.PrivateKey, cfg.ValiditySeconds)
		if signErr != nil {
			// Failed to sign, but still write response
			originalWriter.WriteHeader(statusCode)
			originalWriter.Write(respBuf.Bytes())
			return err
		}

		// Add signature header
		originalWriter.Header().Set(cfg.SignatureHeader, sigHeader)

		// Write response with signature
		originalWriter.WriteHeader(statusCode)
		originalWriter.Write(respBuf.Bytes())

		return err
	}
}

// bodyCaptureWriter wraps http.ResponseWriter to capture response body
type bodyCaptureWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (w *bodyCaptureWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func (w *bodyCaptureWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	// Don't call underlying WriteHeader yet, we'll do it after signing
}

// verifyJSONSignature verifies an Ed25519 signature on JSON data
func verifyJSONSignature(jsonData []byte, sigHeader, pubKeyBase64 string, validitySeconds int64) bool {
	// Parse JSON object
	var obj interface{}
	if err := json.Unmarshal(jsonData, &obj); err != nil {
		return false
	}

	// Canonicalize JSON
	canonJSON, err := canonicaljson.Marshal(obj)
	if err != nil {
		return false
	}

	// Decode signature header
	sigBytes, err := b64.StdEncoding.DecodeString(sigHeader)
	if err != nil {
		return false
	}

	var sigPayload SignaturePayload
	if err := json.Unmarshal(sigBytes, &sigPayload); err != nil {
		return false
	}

	// Parse timestamp and expiry
	ts, err := strconv.ParseInt(sigPayload.Timestamp, 10, 64)
	if err != nil {
		return false
	}

	expiry, err := strconv.ParseInt(sigPayload.Expiry, 10, 64)
	if err != nil {
		return false
	}

	// Verify timestamp is not in the future
	now := time.Now().Unix()
	if ts > now {
		return false
	}

	// Verify not expired
	if now > expiry {
		return false
	}

	// Verify expiry matches expected validity period
	if expiry != ts+validitySeconds {
		return false
	}

	// Construct message hash
	message := fmt.Sprintf("%d%s%d", ts, canonJSON, expiry)
	hash := blake2b.Sum256([]byte(message))

	// Decode public key
	pubBytes, err := b64.StdEncoding.DecodeString(pubKeyBase64)
	if err != nil {
		return false
	}

	// Decode signature
	sigDecoded, err := b64.StdEncoding.DecodeString(sigPayload.Signature)
	if err != nil {
		return false
	}

	// Verify signature
	return ed25519.Verify(pubBytes, hash[:], sigDecoded)
}

// signJSON signs JSON data using Ed25519
func signJSON(jsonData []byte, privKeyBase64 string, validitySeconds int64) (string, error) {
	// Parse JSON object
	var obj interface{}
	if err := json.Unmarshal(jsonData, &obj); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	// Canonicalize JSON
	canonJSON, err := canonicaljson.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("canonicalization failed: %w", err)
	}

	// Generate timestamp and expiry
	currentTime := time.Now().Unix()
	expiry := currentTime + validitySeconds

	// Construct message hash
	message := fmt.Sprintf("%d%s%d", currentTime, canonJSON, expiry)
	hash := blake2b.Sum256([]byte(message))

	// Decode private key
	privBytes, err := b64.StdEncoding.DecodeString(privKeyBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode private key: %w", err)
	}

	// Sign hash
	sig := ed25519.Sign(privBytes, hash[:])
	sigEncoded := b64.StdEncoding.EncodeToString(sig)

	// Create signature payload
	payload := SignaturePayload{
		Timestamp: fmt.Sprintf("%d", currentTime),
		Signature: sigEncoded,
		Expiry:    fmt.Sprintf("%d", expiry),
	}

	// Encode payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("signature header encode failed: %w", err)
	}

	// Base64 encode the JSON payload
	return b64.StdEncoding.EncodeToString(jsonPayload), nil
}
