package middlewares

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	//"net/http"
	"strconv"
	//"strings"
	"sync"
	"time"

	"github.com/gibson042/canonicaljson-go"
	"github.com/gin-gonic/gin"

	b64 "encoding/base64"

	"golang.org/x/crypto/blake2b"
)

func DecryptMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPut && c.Request.Method != http.MethodPost {
			c.Next()
			return
		}

		sig := c.Request.Header.Get("sig")
		if sig == "" {
			c.AbortWithStatusJSON(400, gin.H{"error": "Missing signature header"})
			return
		}

		buf := bodyBufPool.Get().(*bytes.Buffer)
		buf.Reset()
		defer bodyBufPool.Put(buf)

		tee := io.TeeReader(c.Request.Body, buf)

		if !VerifyJSON(tee, sig) {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid signature"})
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))

		c.Next()

	}
}

var bodyBufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func VerifyJSON(r io.Reader, sigHeader string) bool {
	//pubKeyBase64 := "WL5FN2QzDZ/YPPpLM0NK9fHPI/UQ8r2owKevcCyWKj4="
	pubKeyBase64 := "N3Iu59pOFhQ4EcH0jRx3vkhOxUHb6Gh5USHyApvWccM="
	// Parse input JSON into object
	var obj interface{}
	if err := json.NewDecoder(r).Decode(&obj); err != nil {
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

	now := time.Now().Unix()
	if now > expiry {
		return false
	}

	if expiry != ts+3600 {
	}

	if now > expiry || expiry != ts+3600 {
		return false
	}

	// Construct message hash
	message := fmt.Sprintf("%d%s%d", ts, canonJSON, expiry)
	hash := blake2b.Sum256([]byte(message))

	// Decode public key and verify
	pubBytes, err := b64.StdEncoding.DecodeString(pubKeyBase64)
	if err != nil {
		return false
	}

	sigDecoded, err := b64.StdEncoding.DecodeString(sigPayload.Signature)
	if err != nil {
		return false
	}

	return ed25519.Verify(pubBytes, hash[:], sigDecoded)
}

func ResponseSignatureMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPut && c.Request.Method != http.MethodPost {
			c.Next()
			return
		}
		respBody := &bytes.Buffer{}
		writer := &bodyCaptureWriter{ResponseWriter: c.Writer, body: respBody}
		c.Writer = writer

		c.Next()
		if c.Writer.Status() < http.StatusOK || c.Writer.Status() >= http.StatusMultipleChoices {
			writer.ResponseWriter.WriteHeader(c.Writer.Status())
			writer.ResponseWriter.Write(respBody.Bytes())
			return
		}
		// if c.Writer.Status() != http.StatusOK {
		// 	return
		// }

		contentType := c.Writer.Header().Get("Content-Type")
		if !strings.HasPrefix(contentType, "application/json") {
			return
		}

		sigHeader, err := SignJSON(respBody.Bytes(), 3600)
		if err != nil {
			return
		}

		writer.Header().Set("sig", sigHeader)

		writer.ResponseWriter.Write(respBody.Bytes())

	}
}

type bodyCaptureWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyCaptureWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func SignJSON(jsonData []byte, validitySeconds int64) (sigHeader string, err error) {

	privKeyBase64 := "/wnL6WJmFKG4X14zzvNrZog8+dHtaNoD30rEdwxIbf3p+U4BH83F5SrRvIC8M/0Qi6zorAydLGk+/bE8KsBtFA=="
	//privKeyBase64 := "0ATUfSSxF/qtIYu14AF3u3oSed40ZPCoBbQ01H5D79dYvkU3ZDMNn9g8+kszQ0r18c8j9RDyvajAp69wLJYqPg=="
	var obj interface{}
	if err := json.Unmarshal(jsonData, &obj); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	canonJSON, err := canonicaljson.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("canonicalization failed: %w", err)
	}

	currentTime := time.Now().Unix()
	expiry := currentTime + validitySeconds

	message := fmt.Sprintf("%d%s%d", currentTime, canonJSON, expiry)
	hash := blake2b.Sum256([]byte(message))

	privBytes, err := b64.StdEncoding.DecodeString(privKeyBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode private key: %w", err)
	}

	sig := ed25519.Sign(privBytes, hash[:])
	sigEncoded := b64.StdEncoding.EncodeToString(sig)

	payload := SignaturePayload{
		Timestamp: fmt.Sprintf("%d", currentTime),
		Signature: sigEncoded,
		Expiry:    fmt.Sprintf("%d", expiry),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("signature header encode failed: %w", err)
	}

	return b64.StdEncoding.EncodeToString(jsonPayload), nil
}

type SignaturePayload struct {
	Timestamp string `json:"t"`
	Signature string `json:"s"`
	Expiry    string `json:"e"`
}
