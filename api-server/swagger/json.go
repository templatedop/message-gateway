package swagger

import (
	log "MgApplication/api-log"
	"database/sql"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	config "MgApplication/api-config"
	"MgApplication/api-server/util/diutil/typlect"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	null "github.com/volatiletech/null/v9"
)

type m map[string]any

type Docs m

func (d Docs) WithHost(h string) Docs {
	d["Host"] = h
	return d
}

const (
	refKey                  = "$ref"
	PreGeneratedSwaggerFile = "./docs/pregenerated_swagger.json"
)

// SwaggerGenerationMode determines when swagger is generated
type SwaggerGenerationMode int

const (
	// RuntimeMode generates swagger on application startup (default, slower startup)
	RuntimeMode SwaggerGenerationMode = iota
	// BuildMode loads pre-generated swagger from file (fast startup)
	BuildMode
)

// Configuration for swagger generation
var (
	generationMode     SwaggerGenerationMode
	generationModeLock sync.RWMutex
)

// Pre-computed type constants for performance
var (
	typeFileHeader = reflect.TypeOf(multipart.FileHeader{})
	timeExample    string
)

// Pre-computed error examples to avoid repeated allocation
var errorExamples = map[string]map[string]any{
	"400": {"status_code": 400, "message": "Bad Request", "success": false, "error": map[string]any{"code": "400", "message": "bad request", "id": "ERR-400"}},
	"401": {"status_code": 401, "message": "Unauthorized", "success": false, "error": map[string]any{"code": "401", "message": "unauthorized", "id": "ERR-401"}},
	"403": {"status_code": 403, "message": "Forbidden", "success": false, "error": map[string]any{"code": "403", "message": "forbidden", "id": "ERR-403"}},
	"404": {"status_code": 404, "message": "Not Found", "success": false, "error": map[string]any{"code": "404", "message": "not found", "id": "ERR-404"}},
	"422": {"status_code": 422, "message": "Validation Error", "success": false, "error": map[string]any{"code": "422", "message": "validation error", "field_errors": []any{map[string]any{"field": "string", "value": "", "message": "string"}}}},
	"500": {"status_code": 500, "message": "Internal Server Error", "success": false, "error": map[string]any{"code": "500", "message": "internal server error", "id": "ERR-500"}},
}

func init() {
	// Pre-compute time example once at package initialization
	b, _ := time.Now().MarshalJSON()
	timeExample = strings.Trim(string(b), "\"")
}

// loadPreGeneratedSwagger loads swagger documentation from pre-generated file
func loadPreGeneratedSwagger() *openapi3.T {
	data, err := os.ReadFile(PreGeneratedSwaggerFile)
	if err != nil {
		return nil
	}

	var v3Doc openapi3.T
	if err := json.Unmarshal(data, &v3Doc); err != nil {
		fmt.Printf("Error unmarshalling pre-generated swagger: %v\n", err)
		return nil
	}

	return &v3Doc
}

// SetGenerationMode configures how swagger documentation is generated
func SetGenerationMode(mode SwaggerGenerationMode) {
	generationModeLock.Lock()
	defer generationModeLock.Unlock()
	generationMode = mode
}

// GetGenerationMode returns the current generation mode
func GetGenerationMode() SwaggerGenerationMode {
	generationModeLock.RLock()
	defer generationModeLock.RUnlock()
	return generationMode
}

// buildDocs generates OpenAPI v3 documentation from endpoint definitions
// Supports two modes:
// - RuntimeMode: Generates swagger on startup (slower)
// - BuildMode: Loads pre-generated swagger from file (faster)
func buildDocs(eds []EndpointDef, cfg *config.Config) *openapi3.T {
	// Check if we should load from pre-generated file
	if GetGenerationMode() == BuildMode {
		if v3Doc := loadPreGeneratedSwagger(); v3Doc != nil {
			fmt.Println("Loaded pre-generated swagger documentation")
			return v3Doc
		}
		fmt.Println("Warning: BuildMode enabled but pre-generated file not found, falling back to runtime generation")
	}

	// Runtime generation
	loadNullableOverrides(cfg)
	dj := baseJSON(cfg)
	dj["definitions"] = buildDefinitions(eds)
	dj["paths"] = buildPaths(eds)

	var v2Doc openapi2.T
	data, err := json.Marshal(Docs(dj))
	if err != nil {
		return nil
	}
	if err := json.Unmarshal(data, &v2Doc); err != nil {
		return nil
	}
	v3Doc, err := openapi2conv.ToV3(&v2Doc)
	if err != nil {
		fmt.Println("Error converting to v3:", err)
		return nil
	}

	// Attach success & error examples (overrides any missing examples)
	attachErrorExamples(v3Doc)

	// Persist generated v3 document to file asynchronously (non-blocking)
	go func() {
		if err := storeV3DocToFile(v3Doc); err != nil {
			fmt.Println("Error storing v3 doc to file:", err)
		}
	}()

	return v3Doc

	// Populate servers for OpenAPI 3 so tools (Swagger UI/Editor) build correct curl / request URL.
	// Config precedence:
	// 1. swagger.serverUrls (comma separated full URLs)
	// 2. Derived from swagger.host + swagger.basePath + swagger.schemes (first scheme) or server.addr.
	// if len(v3Doc.Servers) == 0 { // only set if not already present
	// 	var serverURLs []string
	// 	if cfg.Exists("swagger.serverUrls") {
	// 		for _, u := range strings.Split(cfg.GetString("swagger.serverUrls"), ",") {
	// 			u = strings.TrimSpace(u)
	// 			if u != "" {
	// 				serverURLs = append(serverURLs, u)
	// 			}
	// 		}
	// 	}
	// 	if len(serverURLs) == 0 {
	// 		// derive host/basePath
	// 		host := cfg.GetString("swagger.host")
	// 		if host == "" {
	// 			// try server.addr like ":8080" or "0.0.0.0:8080"
	// 			if addr := cfg.GetString("server.addr"); addr != "" {
	// 				// normalize
	// 				if strings.HasPrefix(addr, ":") {
	// 					host = "localhost" + addr
	// 				} else {
	// 					host = addr
	// 				}
	// 			} else {
	// 				host = "localhost:8080"
	// 			}
	// 		}
	// 		basePath := cfg.GetString("swagger.basePath")
	// 		if basePath == "" {
	// 			basePath = "/"
	// 		}
	// 		if !strings.HasPrefix(basePath, "/") {
	// 			basePath = "/" + basePath
	// 		}
	// 		scheme := "http"
	// 		if cfg.Exists("swagger.schemes") {
	// 			schs := strings.Split(cfg.GetString("swagger.schemes"), ",")
	// 			if len(schs) > 0 && strings.TrimSpace(schs[0]) != "" {
	// 				scheme = strings.TrimSpace(schs[0])
	// 			}
	// 		}
	// 		// Force http for localhost / loopback unless explicitly forced via swagger.forceHTTPS=true
	// 		if (strings.Contains(host, "localhost") || strings.HasPrefix(host, "127.") || strings.HasPrefix(host, "0.0.0.0")) && !cfg.GetBool("swagger.forceHTTPS") {
	// 			scheme = "http"
	// 		}
	// 		// Downgrade to http if https chosen but server.tls.enabled not set/false (avoid broken curl URLs).
	// 		if scheme == "https" && !cfg.GetBool("server.tls.enabled") && !cfg.GetBool("swagger.forceHTTPS") {
	// 			scheme = "http"
	// 		}
	// 		serverURLs = []string{fmt.Sprintf("%s://%s%s", scheme, host, basePath)}
	// 	}
	// 	for _, u := range serverURLs {
	// 		v3Doc.Servers = append(v3Doc.Servers, &openapi3.Server{URL: u})
	// 	}
	// }

	/*to create json file*/
	// jsonData, err := json.Marshal(v3Doc)
	// if err != nil {
	// 	//fmt.Println("Error marshaling Docs to JSON:", err)
	// 	return nil
	// }

	// file, err := os.Create("v3Doc.json")
	// if err != nil {
	// 	fmt.Println("Error creating file:", err)
	// 	return nil
	// }
	// defer file.Close()
	// _, err = file.Write(jsonData)
	// if err != nil {
	// 	fmt.Println("Error writing JSON to file:", err)
	// 	return nil
	// }
	//fmt.Println("v3doc paths: ",v3Doc.Paths)
	/*to create json file*/
	// if v3Doc.Components.Schemas != nil {
	//     for name, schemaRef := range v3Doc.Components.Schemas {
	//         fmt.Printf("Schema name: %s\n", name)
	//         if schemaRef.Value != nil {
	//             fmt.Printf("Schema details: %+v\n", schemaRef.Value)
	//         }
	//     }
	// }

	//fmt.Println("v3Doc:", v3Doc)

}
func storeV3DocToFile(v3Doc *openapi3.T) error {
	// Optimized: Use json.Marshal instead of MarshalIndent for faster serialization
	// Indentation is not critical for machine-readable files
	v3DocJSON, err := json.Marshal(v3Doc)
	if err != nil {
		return fmt.Errorf("error marshaling v3Doc to JSON: %w", err)
	}

	// Create docs folder if not available
	if _, err := os.Stat("docs"); os.IsNotExist(err) {
		if err := os.Mkdir("docs", os.ModePerm); err != nil {
			return fmt.Errorf("error creating docs directory: %w", err)
		}
	}

	// Write file atomically
	if err := os.WriteFile("./docs/v3Doc.json", v3DocJSON, 0644); err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	fmt.Println("v3Doc has been successfully stored in v3Doc.json")
	return nil
}

// SavePreGeneratedSwagger saves swagger doc to the pre-generated file location
// This is used by the build-time CLI tool
func SavePreGeneratedSwagger(v3Doc *openapi3.T) error {
	v3DocJSON, err := json.Marshal(v3Doc)
	if err != nil {
		return fmt.Errorf("error marshaling v3Doc: %w", err)
	}

	if _, err := os.Stat("docs"); os.IsNotExist(err) {
		if err := os.Mkdir("docs", os.ModePerm); err != nil {
			return fmt.Errorf("error creating docs directory: %w", err)
		}
	}

	if err := os.WriteFile(PreGeneratedSwaggerFile, v3DocJSON, 0644); err != nil {
		return fmt.Errorf("error writing pre-generated swagger: %w", err)
	}

	fmt.Printf("Pre-generated swagger saved to %s\n", PreGeneratedSwaggerFile)
	return nil
}

// Allow hyphen in path param names (e.g., :application-id)
var pathRegexp = regexp.MustCompile(`(\:[A-Za-z0-9_\-]*)`)

func toSwaggerPath(s string) string {
	return pathRegexp.ReplaceAllStringFunc(s, func(s string) string {
		return fmt.Sprintf("{%s}", s[1:])
	})
}

func baseJSON(cfg *config.Config) m {

	cfg.SetDefault("info.description", "")
	cfg.SetDefault("info.version", "1.1.0")
	cfg.SetDefault("info.title", "Application")
	cfg.SetDefault("info.terms", "http://swagger.io/terms/")
	cfg.SetDefault("info.email", "")
	of, err := cfg.Of("info")
	//TODO
	if err != nil {
		fmt.Println("Error getting info:", err)
	}
	//fmt.Println("info value:", of.GetString("version"))
	// Host/basePath/schemes (Swagger 2) can be configured; fallback to sensible defaults.
	// host := cfg.GetString("swagger.host")
	// if host == "" {
	// 	if addr := cfg.GetString("server.addr"); addr != "" {
	// 		if strings.HasPrefix(addr, ":") {
	// 			host = "localhost" + addr
	// 		} else {
	// 			host = addr
	// 		}
	// 	}
	// }
	// basePath := cfg.GetString("swagger.basePath")
	// if basePath == "" {
	// 	basePath = "/"
	// }
	// if !strings.HasPrefix(basePath, "/") {
	// 	basePath = "/" + basePath
	// }
	// schemes := []string{}
	// if cfg.Exists("swagger.schemes") {
	// 	for _, s := range strings.Split(cfg.GetString("swagger.schemes"), ",") {
	// 		s = strings.TrimSpace(s)
	// 		if s != "" {
	// 			schemes = append(schemes, s)
	// 		}
	// 	}
	// }
	return m{
		"swagger": "2.0",
		"info": m{
			"description":    of.GetString("description"),
			"version":        of.GetString("version"),
			"title":          cfg.GetString("info.title"),
			"termsOfService": cfg.GetString("info.terms"),
			"contact":        m{"email": cfg.GetString("info.email")},
			"license":        m{"name": "Apache 2.0", "url": "http://www.apache.org/licenses/LICENSE-2.0.html"},
		},
		// "host":     host,
		// "basePath": basePath,
		// "schemes":  schemes,
		"host":     "",
		"basePath": "/",
		"schemes":  []string{},
	}
}

func withDefinitionPrefix(s string) string {
	//fmt.Println("withDefinitionPrefix: ", s)

	return fmt.Sprintf("#/definitions/%s", s)
}

func getPrimitiveType(t reflect.Type) m {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return m{"type": "integer", "format": fmt.Sprintf("int%d", t.Bits())}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return m{"type": "integer", "format": fmt.Sprintf("uint%d", t.Bits())}
	case reflect.Float32:
		return m{"type": "number", "format": "float"}
	case reflect.Float64:
		return m{"type": "number", "format": "double"}
	case reflect.Bool:
		return m{"type": "boolean"}
	case reflect.String:
		return m{"type": "string"}
	case reflect.Map, reflect.Interface, reflect.Struct:
		// Struct handled earlier normally, but if it falls through treat as object.
		return m{"type": "object"}
	case reflect.Func:
		// Represent function pointers as string for documentation purposes.
		return m{"type": "string", "format": "func"}
	case reflect.Chan:
		return m{"type": "string", "format": "channel"}
	case reflect.UnsafePointer:
		return m{"type": "string", "format": "pointer"}
	default:
		// Fallback to string to ensure only valid OpenAPI primitive types are emitted.
		return m{"type": "string"}
	}
}

func getPropertyField(t reflect.Type) m {
	if t == typlect.TypeNoParam {
		return m{"type": "string"}
	}

	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	// Optimized: Use pre-computed constant instead of reflect.TypeOf
	if t == typeFileHeader {
		return m{"type": "string", "format": "binary"}
	}

	// Map known nullable wrapper types (github.com/aarondl/null & database/sql) to primitives
	if v, ok := nullableTypeMapping(t); ok {
		return v
	}

	// Optimized: Use pre-computed timeExample instead of generating each time
	if t == typlect.TypeTime {
		return m{"type": "string", "example": timeExample}
	}

	if t.Kind() == reflect.Struct {
		//fmt.Println("t.Kind() == reflect.Struct: ", t)
		return m{
			refKey: withDefinitionPrefix(getNameFromType(t)),
		}
	}

	if t.Kind() == reflect.Slice {
		return arrayProperty(t)
	}

	return getPrimitiveType(t)
}

// nullableTypeMapping returns an OpenAPI schema mapping for supported nullable wrapper types.
func nullableTypeMapping(t reflect.Type) (m, bool) {
	// 1. Check dynamic overrides loaded from config (by reflect.Type string name)
	if ov, ok := nullableOverrides[t.String()]; ok {
		return ov, true
	}

	// 2. Fallback to static built‑ins (types are comparable keys)
	v, ok := builtinNullableTypeMap[t]
	return v, ok
}

// builtinNullableTypeMap holds the default mappings for well-known nullable types.
var builtinNullableTypeMap = map[reflect.Type]m{
	reflect.TypeOf(sql.NullString{}):  {"type": "string"},
	reflect.TypeOf(sql.NullInt64{}):   {"type": "integer", "format": "int64"},
	reflect.TypeOf(sql.NullBool{}):    {"type": "boolean"},
	reflect.TypeOf(sql.NullFloat64{}): {"type": "number", "format": "double"},
	reflect.TypeOf(sql.NullTime{}):    {"type": "string", "format": "date-time"},

	// github.com/aarondl/null/v9 (supports JSON marshalling similar to primitives)
	// reflect.TypeOf(null.String{}):  {"type": "string"},
	reflect.TypeOf(null.Int{}):     {"type": "integer", "format": "int64"},
	reflect.TypeOf(null.Int64{}):   {"type": "integer", "format": "int64"},
	reflect.TypeOf(null.Uint{}):    {"type": "integer", "format": "uint32"},
	reflect.TypeOf(null.Uint64{}):  {"type": "integer", "format": "uint64"},
	reflect.TypeOf(null.Bool{}):    {"type": "boolean"},
	reflect.TypeOf(null.Float32{}): {"type": "number", "format": "float"},
	reflect.TypeOf(null.Float64{}): {"type": "number", "format": "double"},
	reflect.TypeOf(null.Time{}):    {"type": "string", "format": "date-time"},
}

// nullableOverrides stores user-provided override mappings keyed by reflect.Type.String()
// (e.g. "sql.NullString", "null.String"). Values must be OpenAPI schema fragments.
var nullableOverrides = map[string]m{}

// loadNullableOverrides loads JSON (or inline YAML string) from config key
// swagger.nullableTypeMap. Expected format example (JSON string):
//
//	{
//	  "sql.NullString": {"type": "string"},
//	  "null.Uint64": {"type": "integer", "format": "uint64"}
//	}
//
// Any invalid JSON is logged and ignored. This function is idempotent; later calls
// replace previously loaded overrides.
func loadNullableOverrides(cfg *config.Config) {
	if cfg == nil || !cfg.Exists("swagger.nullableTypeMap") {
		return
	}
	raw := strings.TrimSpace(cfg.GetString("swagger.nullableTypeMap"))
	if raw == "" {
		return
	}
	// If the user provides YAML block with JSON inside, that's fine. Attempt unmarshal.
	var tmp map[string]map[string]any
	if err := json.Unmarshal([]byte(raw), &tmp); err != nil {
		log.Error(nil, "swagger: unable to parse swagger.nullableTypeMap JSON: %v", err)
		return
	}
	converted := make(map[string]m, len(tmp))
	for k, v := range tmp {
		converted[k] = v
	}
	nullableOverrides = converted
}

func arrayProperty(t reflect.Type) m {
	it := t.Elem()
	if it.Kind() == reflect.Pointer {
		it = it.Elem()
	}

	return m{
		"type":  "array",
		"items": getPropertyField(it),
	}
}

func getNameFromType(t reflect.Type) string {
	name := t.Name()
	// Optimized: Use strings.Builder to avoid multiple allocations
	var sb strings.Builder
	sb.Grow(len(name)) // Pre-allocate capacity
	for _, r := range name {
		switch r {
		case ']': // skip
		case '/':
			sb.WriteByte('_')
		case '*': // skip
		case '[':
			sb.WriteString("__")
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// attachErrorExamples walks the v3 doc and attaches an Example to each non-2xx response
// whose schema $ref points to APIErrorResponse (case-insensitive suffix match).

func attachErrorExamples(doc *openapi3.T) {
	if doc == nil || doc.Paths == nil {
		return
	}
	for _, item := range doc.Paths.Map() { // iterate path items
		if item == nil {
			continue
		}
		for _, op := range []*openapi3.Operation{item.Get, item.Put, item.Post, item.Delete, item.Options, item.Head, item.Patch, item.Trace} {
			if op == nil || op.Responses == nil {
				continue
			}
			for code, respRef := range op.Responses.Map() {
				if respRef == nil || respRef.Value == nil {
					continue
				}
				for cType, media := range respRef.Value.Content {
					if media == nil || media.Schema == nil {
						continue
					}
					if media.Example != nil {
						continue
					}
					status := parseStatus(code)
					desc := "Response"
					if respRef.Value.Description != nil && *respRef.Value.Description != "" {
						desc = *respRef.Value.Description
					}
					// Success example with data field
					if len(code) > 0 && code[0] == '2' {
						// Build full example from schema (expands refs)
						ex := buildSchemaExample(media.Schema, doc.Components, 0, map[string]struct{}{})
						// Ensure standard fields if present in schema
						if exObj, ok := ex.(map[string]any); ok {
							// override common wrapper fields if they exist
							if _, ok := exObj["status_code"]; ok {
								exObj["status_code"] = status
							}
							if _, ok := exObj["message"]; ok {
								exObj["message"] = desc
							}
							if _, ok := exObj["success"]; ok {
								exObj["success"] = true
							}
							media.Example = exObj
						} else {
							media.Example = ex
						}
						respRef.Value.Content[cType] = media
						continue
					}
					// Error example only if schema ref matches APIErrorResponse
					ref := media.Schema.Ref
					if ref == "" {
						continue
					}
					lr := strings.ToLower(ref)
					if !strings.HasSuffix(lr, "/apierrorresponse") && !strings.HasSuffix(lr, ".apierrorresponse") {
						continue
					}

					// Optimized: Use pre-computed error examples
					if ex, ok := errorExamples[code]; ok {
						media.Example = ex
					} else {
						// Fallback for non-standard error codes
						errObj := map[string]any{
							"code":    code,
							"message": desc,
							"id":      "ERR-EXAMPLE-ID",
						}
						media.Example = map[string]any{
							"status_code": status,
							"message":     desc,
							"success":     false,
							"error":       errObj,
						}
					}
					respRef.Value.Content[cType] = media
				}
			}
		}
	}
}

// inferSchemaExample produces a lightweight example for a SchemaRef.
func inferSchemaExample(sr *openapi3.SchemaRef) any {
	if sr == nil || sr.Value == nil {
		return nil
	}
	if sr.Ref != "" {
		return map[string]any{"$ref": sr.Ref}
	}
	s := sr.Value
	// kin-openapi may represent type as string or slice (Types). Normalize.
	var t string
	if s.Type == nil || len(*s.Type) == 0 {
		if len(s.Properties) > 0 {
			t = "object"
		} else {
			return nil
		}
	} else if len(*s.Type) == 1 {
		t = (*s.Type)[0]
	} else {
		for _, cand := range *s.Type {
			if cand != "null" {
				t = cand
				break
			}
		}
		if t == "" {
			t = (*s.Type)[0]
		}
	}
	switch t {
	case "string":
		return "string"
	case "integer":
		return 0
	case "number":
		return 0.0
	case "boolean":
		return true
	case "array":
		return []any{inferSchemaExample(s.Items)}
	case "object":
		obj := map[string]any{}
		for name, prop := range s.Properties {
			obj[name] = inferSchemaExample(prop)
		}
		return obj
	default:
		return nil
	}
}

// buildSchemaExample recursively expands a schema (resolving $refs) into a representative example value.
// depth is limited to prevent infinite recursion on self-referential schemas.
func buildSchemaExample(sr *openapi3.SchemaRef, comp *openapi3.Components, depth int, seen map[string]struct{}) any {
	if sr == nil {
		return nil
	}
	if depth > 6 {
		return nil
	}
	if sr.Ref != "" {
		refName := sr.Ref
		if i := strings.LastIndex(refName, "/"); i >= 0 {
			refName = refName[i+1:]
		}
		if _, cyc := seen[refName]; cyc {
			return nil
		}
		if comp != nil {
			if target, ok := comp.Schemas[refName]; ok {
				seen[refName] = struct{}{}
				return buildSchemaExample(target, comp, depth+1, seen)
			}
		}
		return map[string]any{"$ref": sr.Ref}
	}
	if sr.Value == nil {
		return nil
	}
	s := sr.Value
	// Determine type similarly to inferSchemaExample
	var t string
	if s.Type == nil || len(*s.Type) == 0 {
		if len(s.Properties) > 0 {
			t = "object"
		} else {
			return nil
		}
	} else {
		t = (*s.Type)[0]
		if t == "null" && len(*s.Type) > 1 {
			t = (*s.Type)[1]
		}
	}
	switch t {
	case "string":
		if s.Format == "date-time" {
			return time.Now().UTC().Format(time.RFC3339)
		}
		return "string"
	case "integer":
		return 0
	case "number":
		return 0.0
	case "boolean":
		return true
	case "array":
		return []any{buildSchemaExample(s.Items, comp, depth+1, seen)}
	case "object":
		obj := map[string]any{}
		for name, prop := range s.Properties {
			obj[name] = buildSchemaExample(prop, comp, depth+1, seen)
		}
		return obj
	default:
		return nil
	}
}

func parseStatus(code string) int {
	if n, err := strconv.Atoi(code); err == nil {
		return n
	}
	return 0
}
