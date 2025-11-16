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
	refKey = "$ref"
)

// func buildDocs(eds []EndpointDef, cfg *config.Config) Docs {
func buildDocs(eds []EndpointDef, cfg *config.Config) *openapi3.T {
	// Load any nullable type override mappings from config before generating docs
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

	// // Attach success & error examples (overrides any missing examples)
	attachErrorExamples(v3Doc)

	// Persist generated v3 document to file (ignore error)
	err = storeV3DocToFile(v3Doc)
	if err != nil {
		fmt.Println("Error storing v3 doc to file:", err)
	}

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
	// Marshal v3Doc to JSON
	v3DocJSON, err := json.MarshalIndent(v3Doc, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling v3Doc to JSON: %w", err)
	}

	//create docs folder if not available
	if _, err := os.Stat("docs"); os.IsNotExist(err) {
		os.Mkdir("docs", os.ModePerm)
	}

	// Create or open a file
	file, err := os.Create("./docs/v3Doc.json")
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	// Write JSON data to the file
	if _, err := file.Write(v3DocJSON); err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}
	// if err := loadAndResolve(); err != nil {
	//     log.Fatal(err)
	// }

	fmt.Println("v3Doc has been successfully stored in v3Doc.json")
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

	//This is for examples....
	//fmt.Println("getPropertyField: ", t)
	if t == typlect.TypeNoParam {
		return m{"type": "string"}
	}
	/* This is for mapping special types */
	// if t == reflect.TypeOf(sql.NullString{}) {

	// 	return m{"type": "string"}
	// }
	/* This is for mapping special types */

	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	// Treat uploaded files as binary strings to avoid unresolved FileHeader schema refs
	if t == reflect.TypeOf(multipart.FileHeader{}) {
		return m{"type": "string", "format": "binary"}
	}

	// Map known nullable wrapper types (github.com/aarondl/null & database/sql) to primitives
	if v, ok := nullableTypeMapping(t); ok {
		return v
	}

	if t == typlect.TypeTime {
		b, _ := time.Now().MarshalJSON()
		return m{"type": "string", "example": strings.Trim(string(b), "\"")}
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
	s := strings.ReplaceAll(t.Name(), "]", "")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "*", "")
	return strings.ReplaceAll(s, "[", "__")
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
					errObj := map[string]any{
						"code":    code,
						"message": desc,
						"id":      "ERR-EXAMPLE-ID",
					}
					var ex map[string]any
					if code == "422" { // validation error format
						errObj["message"] = "validation error"
						ex = map[string]any{
							"status_code": status,
							"message":     "Validation Error",
							"success":     false,
							"error": map[string]any{
								"code":         code,
								"message":      "validation error",
								"field_errors": []any{map[string]any{"field": "string", "value": "", "message": "string"}},
							},
						}
					} else {
						ex = map[string]any{
							"status_code": status,
							"message":     desc,
							"success":     false,
							"error":       errObj,
						}
					}
					media.Example = ex
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
