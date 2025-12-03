package swagger

import (
	"database/sql"
	"reflect"
	"strings"
	"sync"

	errors "MgApplication/api-errors"
	"MgApplication/api-server/util/diutil/typlect"
	"MgApplication/api-server/util/slc"
)

// Type cache to avoid processing the same type multiple times
var (
	processedTypes sync.Map // map[reflect.Type]bool
)

func buildDefinitions(eds []EndpointDef) m {
	// Pre-size map based on endpoint count (heuristic: avg 2-3 types per endpoint)
	estimatedSize := len(eds) * 3
	defs := make(m, estimatedSize)

	// Reset type cache for new generation
	processedTypes = sync.Map{}

	for _, ed := range eds {
		buildModelDefinition(defs, ed.RequestType, true)
		buildModelDefinition(defs, ed.ResponseType, false)
		buildModelDefinition(defs, reflect.TypeOf(errors.APIErrorResponse{}), false)
	}

	// Ensure APIErrorResponse schema explicitly shows success=false in examples.
	errType := reflect.TypeOf(errors.APIErrorResponse{})
	if def, ok := defs[getNameFromType(errType)].(m); ok {
		if props, ok2 := def["properties"].(m); ok2 {
			if succ, ok3 := props["success"].(m); ok3 {
				succ["example"] = false
				succ["default"] = false
				props["success"] = succ
			}
			def["properties"] = props
			defs[getNameFromType(errType)] = def
		}
	}

	return defs
}

func buildModelDefinition(defs m, t reflect.Type, isReq bool) {
	if t == typlect.TypeNoParam {
		return
	}

	if t.Kind() == reflect.Slice {
		t = t.Elem()
	}

	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return
	}

	// Check type cache - skip if already processed
	if _, exists := processedTypes.Load(t); exists {
		return
	}
	processedTypes.Store(t, true)

	// Pre-size collections based on field count for better memory efficiency
	numFields := t.NumField()
	var smr []string
	smp := make(m, numFields) // Pre-allocate map with expected capacity

	// Cache fields to avoid repeated reflection calls
	for i := 0; i < numFields; i++ {
		f := t.Field(i)
		ft := f.Type
		// if basicType, ok := typeMapping[ft]; ok {
		//     fmt.Println("Converting special type: ", f)
		//     ft = basicType
		// }

		// Normalize certain special numeric / nullable types so nested struct traversal works as expected.
		if ft.Kind() == reflect.Uint64 {
			//fmt.Println("came inside uint64: ", f)
			ft = reflect.TypeOf(int(0))
		}
		if ft == reflect.TypeOf(sql.NullString{}) {
			//fmt.Println("came inside Nulstring: ", f)
			//ft = reflect.TypeOf(string)
		}

		// build subtype definitions (include pointer + slice + slice of pointer cases)
		if ft != typlect.TypeTime {
			// direct struct
			if ft.Kind() == reflect.Struct {
				buildModelDefinition(defs, ft, isReq)
			}
			// pointer to struct
			if ft.Kind() == reflect.Pointer && ft.Elem().Kind() == reflect.Struct {
				buildModelDefinition(defs, ft.Elem(), isReq)
			}
			// slice of structs
			if ft.Kind() == reflect.Slice {
				el := ft.Elem()
				if el.Kind() == reflect.Pointer {
					el = el.Elem()
				}
				if el.Kind() == reflect.Struct {
					buildModelDefinition(defs, el, isReq)
				}
			}
		}

		// Handle inline embedding: json:",inline" -> flatten properties of embedded struct
		if tag := f.Tag.Get("json"); strings.Contains(tag, ",inline") {
			// Dereference pointers
			inner := ft
			if inner.Kind() == reflect.Pointer {
				inner = inner.Elem()
			}
			if inner.Kind() == reflect.Struct {
				// Ensure inner definition built
				buildModelDefinition(defs, inner, isReq)
				if def, ok := defs[getNameFromType(inner)].(m); ok {
					if props, ok := def["properties"].(m); ok {
						for k, v := range props {
							smp[k] = v
						}
						if isReq {
							if reqs, ok := def["required"].([]string); ok {
								smr = append(smr, reqs...)
							}
						}
					}
				}
			}
			continue // Skip normal property addition
		}

		if !isReq || f.Tag.Get("json") != "" {
			//fmt.Println("FieldName: ", getFieldName(f))
			//fmt.Println("fname: ", f.Name)
			// fmt.Println("ftype: ",f.Type)
			// fmt.Println("ftype kind: ",f.Type.Kind())
			// fmt.Println("ftype name: ",f.Type.Name())
			//fmt.Println("Tag: ", f.Tag.Get("json"))
			if f.Tag.Get("json") == "-" {
				continue
			}
			//fmt.Println("f type:", f.Type)
			if f.Type == reflect.TypeOf(sql.NullString{}) {
				//fmt.Println("Name inside nullstring:", f.Name)
				//fmt.Println("came inside Nulstring: ", f)
				//f.Type = reflect.TypeOf("")
				//f.Name = "string"

				//fmt.Println("After changing type inside Nullstring: ", f)
			}

			smp[getFieldName(f)] = getPropertyField(f.Type)

			if vts, ok := f.Tag.Lookup("validate"); isReq && ok {
				if slc.Contains(strings.Split(vts, ","), "required") {
					smr = append(smr, getFieldName(f))
				}
			}
		}

		//fmt.Println("f:", f, "ft:", ft)
	}

	if len(smp) > 0 {
		mi := m{
			"type":       "object",
			"properties": smp,
		}

		if len(smr) > 0 {
			mi["required"] = smr
		}

		//fmt.Println("getNameFromType(t): ", getNameFromType(t))

		defs[getNameFromType(t)] = mi
	}
}

func getFieldName(f reflect.StructField) string {
	if tag := f.Tag.Get("json"); tag != "" {
		// Optimized: Use Index instead of Split to avoid allocation
		if idx := strings.Index(tag, ","); idx >= 0 {
			return tag[:idx]
		}
		return tag
	}

	return f.Name
}
