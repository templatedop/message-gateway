package routeradapter

import (
	"fmt"
	"strings"
	"time"
)

// RouterType represents the type of router/framework to use
type RouterType string

const (
	// RouterTypeGin uses the Gin web framework (default)
	RouterTypeGin RouterType = "gin"

	// RouterTypeFiber uses the Fiber web framework
	RouterTypeFiber RouterType = "fiber"

	// RouterTypeEcho uses the Echo web framework
	RouterTypeEcho RouterType = "echo"

	// RouterTypeNetHTTP uses the standard library net/http
	RouterTypeNetHTTP RouterType = "nethttp"
)

// RouterConfig contains configuration for the router adapter
type RouterConfig struct {
	// Type specifies which framework to use (gin, fiber, echo, nethttp)
	// Defaults to gin if not specified
	Type RouterType `yaml:"type" json:"type"`

	// Port is the HTTP server port
	Port int `yaml:"port" json:"port"`

	// Gin-specific configuration
	Gin *GinConfig `yaml:"gin,omitempty" json:"gin,omitempty"`

	// Fiber-specific configuration
	Fiber *FiberConfig `yaml:"fiber,omitempty" json:"fiber,omitempty"`

	// Echo-specific configuration
	Echo *EchoConfig `yaml:"echo,omitempty" json:"echo,omitempty"`

	// NetHTTP-specific configuration
	NetHTTP *NetHTTPConfig `yaml:"nethttp,omitempty" json:"nethttp,omitempty"`

	// Common server configuration
	ReadTimeout       time.Duration `yaml:"readTimeout" json:"readTimeout"`
	WriteTimeout      time.Duration `yaml:"writeTimeout" json:"writeTimeout"`
	IdleTimeout       time.Duration `yaml:"idleTimeout" json:"idleTimeout"`
	ReadHeaderTimeout time.Duration `yaml:"readHeaderTimeout" json:"readHeaderTimeout"`
	MaxHeaderBytes    int           `yaml:"maxHeaderBytes" json:"maxHeaderBytes"`
}

// GinConfig contains Gin-specific configuration
type GinConfig struct {
	// Mode sets the Gin mode: "debug", "release", or "test"
	Mode string `yaml:"mode" json:"mode"`

	// TrustedProxies is a list of trusted proxy IP addresses
	TrustedProxies []string `yaml:"trustedProxies,omitempty" json:"trustedProxies,omitempty"`

	// ForwardedByClientIP enables parsing X-Forwarded-For and X-Real-IP headers
	ForwardedByClientIP bool `yaml:"forwardedByClientIP" json:"forwardedByClientIP"`

	// RemoveExtraSlash enables removing extra slashes from paths
	RemoveExtraSlash bool `yaml:"removeExtraSlash" json:"removeExtraSlash"`
}

// FiberConfig contains Fiber-specific configuration
type FiberConfig struct {
	// Prefork enables use of SO_REUSEPORT socket option
	// Allows spawning multiple Go processes listening on the same port
	Prefork bool `yaml:"prefork" json:"prefork"`

	// ServerHeader sets the Server HTTP header value
	ServerHeader string `yaml:"serverHeader,omitempty" json:"serverHeader,omitempty"`

	// StrictRouting enables strict routing (case-sensitive, trailing slash)
	StrictRouting bool `yaml:"strictRouting" json:"strictRouting"`

	// CaseSensitive enables case-sensitive routing
	CaseSensitive bool `yaml:"caseSensitive" json:"caseSensitive"`

	// ETag enables automatic ETag header generation
	ETag bool `yaml:"etag" json:"etag"`

	// BodyLimit sets the maximum request body size
	BodyLimit int `yaml:"bodyLimit" json:"bodyLimit"`

	// Concurrency sets the maximum number of concurrent connections
	Concurrency int `yaml:"concurrency" json:"concurrency"`

	// DisableKeepalive disables HTTP keep-alive
	DisableKeepalive bool `yaml:"disableKeepalive" json:"disableKeepalive"`
}

// EchoConfig contains Echo-specific configuration
type EchoConfig struct {
	// Debug enables debug mode
	Debug bool `yaml:"debug" json:"debug"`

	// HideBanner hides the Echo banner on startup
	HideBanner bool `yaml:"hideBanner" json:"hideBanner"`

	// HidePort hides the port number in the banner
	HidePort bool `yaml:"hidePort" json:"hidePort"`

	// DisableHTTP2 disables HTTP/2 support
	DisableHTTP2 bool `yaml:"disableHTTP2" json:"disableHTTP2"`
}

// NetHTTPConfig contains net/http-specific configuration
type NetHTTPConfig struct {
	// MaxMultipartMemory sets the maximum memory for multipart forms
	MaxMultipartMemory int64 `yaml:"maxMultipartMemory" json:"maxMultipartMemory"`

	// EnableHTTP2 enables HTTP/2 support
	EnableHTTP2 bool `yaml:"enableHTTP2" json:"enableHTTP2"`
}

// Validate validates the router configuration
func (c *RouterConfig) Validate() error {
	// Normalize router type
	c.Type = RouterType(strings.ToLower(string(c.Type)))

	// Default to Gin if not specified
	if c.Type == "" {
		c.Type = RouterTypeGin
	}

	// Validate router type
	switch c.Type {
	case RouterTypeGin, RouterTypeFiber, RouterTypeEcho, RouterTypeNetHTTP:
		// Valid types
	default:
		return fmt.Errorf("invalid router type: %s (must be gin, fiber, echo, or nethttp)", c.Type)
	}

	// Validate port
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be 1-65535)", c.Port)
	}

	// Set default timeouts if not specified
	if c.ReadTimeout == 0 {
		c.ReadTimeout = 30 * time.Second
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = 30 * time.Second
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = 120 * time.Second
	}
	if c.ReadHeaderTimeout == 0 {
		c.ReadHeaderTimeout = 10 * time.Second
	}

	// Validate framework-specific configs
	switch c.Type {
	case RouterTypeGin:
		if c.Gin == nil {
			c.Gin = &GinConfig{}
		}
		if err := c.Gin.Validate(); err != nil {
			return fmt.Errorf("invalid gin config: %w", err)
		}

	case RouterTypeFiber:
		if c.Fiber == nil {
			c.Fiber = &FiberConfig{}
		}
		if err := c.Fiber.Validate(); err != nil {
			return fmt.Errorf("invalid fiber config: %w", err)
		}

	case RouterTypeEcho:
		if c.Echo == nil {
			c.Echo = &EchoConfig{}
		}
		if err := c.Echo.Validate(); err != nil {
			return fmt.Errorf("invalid echo config: %w", err)
		}

	case RouterTypeNetHTTP:
		if c.NetHTTP == nil {
			c.NetHTTP = &NetHTTPConfig{}
		}
		if err := c.NetHTTP.Validate(); err != nil {
			return fmt.Errorf("invalid nethttp config: %w", err)
		}
	}

	return nil
}

// Validate validates GinConfig
func (c *GinConfig) Validate() error {
	// Set default mode
	if c.Mode == "" {
		c.Mode = "release"
	}

	// Validate mode
	switch c.Mode {
	case "debug", "release", "test":
		// Valid modes
	default:
		return fmt.Errorf("invalid gin mode: %s (must be debug, release, or test)", c.Mode)
	}

	return nil
}

// Validate validates FiberConfig
func (c *FiberConfig) Validate() error {
	// Set defaults
	if c.BodyLimit == 0 {
		c.BodyLimit = 4 * 1024 * 1024 // 4 MB default
	}

	if c.Concurrency == 0 {
		c.Concurrency = 256 * 1024 // 256k default
	}

	return nil
}

// Validate validates EchoConfig
func (c *EchoConfig) Validate() error {
	// Echo has minimal configuration requirements
	// Most settings are fine with zero values
	return nil
}

// Validate validates NetHTTPConfig
func (c *NetHTTPConfig) Validate() error {
	// Set defaults
	if c.MaxMultipartMemory == 0 {
		c.MaxMultipartMemory = 32 << 20 // 32 MB default
	}

	return nil
}

// DefaultRouterConfig returns a default router configuration
func DefaultRouterConfig() *RouterConfig {
	return &RouterConfig{
		Type:              RouterTypeGin,
		Port:              8080,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB
		Gin: &GinConfig{
			Mode:                "release",
			ForwardedByClientIP: true,
			RemoveExtraSlash:    true,
		},
	}
}
