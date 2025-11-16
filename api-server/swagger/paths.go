package swagger

import (
	//"fmt"
	"reflect"
	"strings"

	errors "MgApplication/api-errors"
	"MgApplication/api-server/util/slc"
)

func buildPaths(eds []EndpointDef) m {
	p := make(m)
	for _, ed := range eds {
		if strings.HasPrefix(ed.Endpoint, "/__") {
			continue
		}

		desc := m{
			"tags":        []string{ed.Group},
			"summary":     ed.Name,
			"description": "",
			"consumes": []string{
				"application/json",
			},
			"produces": []string{
				"application/json",
			},
			// "externalDocs": m{},
		}

		params := getParameters(ed.RequestType)
		// Ensure path params present for each :segment in endpoint
		missing := map[string]struct{}{}
		if matches := pathRegexp.FindAllString(ed.Endpoint, -1); len(matches) > 0 {
			existing := map[string]struct{}{}
			for _, p := range params {
				if name, ok := p["name"].(string); ok {
					existing[name] = struct{}{}
				}
			}
			for _, mseg := range matches {
				seg := strings.TrimPrefix(mseg, ":")
				if _, ok := existing[seg]; !ok {
					missing[seg] = struct{}{}
				}
			}
		}
		for seg := range missing {
			params = append(params, m{
				"in":       "path",
				"name":     seg,
				"required": true,
				"type":     "string",
			})
		}
		if len(params) > 0 {
			desc["parameters"] = params
		}

		// r200 := getPropertyField(ed.ResponseType)
		// r400 := getPropertyField(reflect.TypeOf(response.ResponseError{}))
		// mergedSchema := mergeSchemas(r200, r400)
		//  fmt.Println("mergedSchema:", mergedSchema)

		// Base success response
		responses := m{
			"200": m{
				"description": "Successful Operation",
				"schema":      getPropertyField(ed.ResponseType),
			},
		}

		// Standard error schema reference (already added to definitions in buildDefinitions)
		errSchema := getPropertyField(reflect.TypeOf(errors.APIErrorResponse{}))
		// Default error codes (422 added for validation errors)
		defaultErrorCodes := []string{"400", "401", "403", "404", "422", "500"}
		for _, code := range defaultErrorCodes {
			responses[code] = m{
				"description": httpStatusDescription(code),
				"schema":      errSchema,
				// "examples": m{
				// 	"application/json": buildErrorExample(code),
				// },
			}
		}

		desc["responses"] = responses

		// desc["responses"] = m{
		// 	"200": m{
		// 		"description": "successful operation",
		// 		"schema":      mergedSchema,
		// 	},
		// 	//"schema":      mergedSchema,
		// }

		meth := strings.ToLower(ed.Method)
		swagp := toSwaggerPath(ed.Endpoint)

		existdesc, ok := p[swagp].(m)
		if !ok {
			existdesc = m{}
		}

		existdesc[meth] = desc
		p[swagp] = existdesc
	}

	return p
}

// httpStatusDescription returns a short description for common HTTP status codes used in auto-generated responses.
func httpStatusDescription(code string) string {
	switch code {
	case "400":
		return "Bad Request"
	case "401":
		return "Unauthorized"
	case "403":
		return "Forbidden"
	case "404":
		return "Not Found"
	case "422":
		return "Validation Error"
	case "500":
		return "Internal Server Error"
	default:
		return "Error"
	}
}

// getParameters inspects the given request type (struct) and extracts Swagger parameter definitions
// based on struct tags: `param` or `uri` for path params, `query` for query params, `form` for form data,
// and `json` for body (if any json tags present). It supports embedded structs and validation tags.
func getParameters(t reflect.Type) []m {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice { // treat slice as body
		mi := arrayProperty(t)
		mi["name"], mi["in"] = "body", "body"
		return []m{mi}
	}

	var params []m
	var hasBody bool

	var walk func(rt reflect.Type)
	walk = func(rt reflect.Type) {
		if rt.Kind() == reflect.Pointer {
			rt = rt.Elem()
		}
		if rt.Kind() != reflect.Struct {
			return
		}
		for i := 0; i < rt.NumField(); i++ {
			f := rt.Field(i)
			ft := f.Type

			// Inline / embedded struct (anonymous) or json:",inline"
			if (f.Anonymous && ft.Kind() == reflect.Struct) || strings.Contains(f.Tag.Get("json"), ",inline") {
				walk(ft)
				continue
			}

			required := false
			if vts, ok := f.Tag.Lookup("validate"); ok {
				if slc.Contains(strings.Split(vts, ","), "required") {
					required = true
				}
			}

			// path param: support `param` or `uri` tag
			if n := firstNonEmpty(f.Tag.Get("param"), f.Tag.Get("uri")); n != "" {
				pi := getPropertyField(ft)
				pi["in"], pi["name"], pi["description"], pi["required"] = "path", n, "", true
				params = append(params, pi)
			}

			// explicit query tag
			if n := f.Tag.Get("query"); n != "" {
				pi := getPropertyField(ft)
				pi["in"], pi["name"], pi["description"] = "query", n, ""
				if required {
					pi["required"] = true
				}
				params = append(params, pi)
			}

			// form tag (treat as query)
			if raw := f.Tag.Get("form"); raw != "" {
				parts := strings.Split(raw, ",")
				name := parts[0]
				if name != "" { // ignore default or other options after comma
					pi := getPropertyField(ft)
					pi["in"], pi["name"], pi["description"] = "query", name, ""
					if required {
						pi["required"] = true
					}
					params = append(params, pi)
				}
			}

			if f.Tag.Get("json") != "" {
				hasBody = true
			}
		}
	}

	walk(t)

	if hasBody {
		params = append(params, m{
			"in": "body", "name": "body", "description": "", "required": true,
			"schema": m{refKey: withDefinitionPrefix(getNameFromType(t))},
		})
	}
	return params
}

// helper to pick first non-empty string
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// buildErrorExample constructs an example JSON object for a given HTTP error status code.
// The structure mirrors APIErrorResponse with success=false and a placeholder error payload.
func buildErrorExample(code string) m {
	msg := httpStatusDescription(code)
	errObj := m{"code": code, "message": strings.ToLower(msg), "id": "ERR-EXAMPLE-ID"}
	if code == "422" {
		// Place field_errors directly under error per requirement
		errObj["message"] = "validation error"
		errObj["field_errors"] = []m{
			{"field": "valid_from", "value": "", "message": "valid_from is a required field"},
		}
	}
	return m{
		"status_code": strToInt(code),
		"message":     msg,
		"success":     false,
		"error":       errObj,
	}
}

// strToInt converts numeric status code strings to int (fallback 0 if parse fails).
func strToInt(s string) int {
	var n int
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}
