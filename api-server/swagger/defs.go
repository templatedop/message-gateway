package swagger

import (
	"database/sql"
	"reflect"
	"strings"

	errors "MgApplication/api-errors"
	"MgApplication/api-server/util/diutil/typlect"
	"MgApplication/api-server/util/slc"
)

func buildDefinitions(eds []EndpointDef) m {
	defs := make(m)

	for _, ed := range eds {

		buildModelDefinition(defs, ed.RequestType, true)
		buildModelDefinition(defs, ed.ResponseType, false)
		buildModelDefinition(defs, reflect.TypeOf(errors.APIErrorResponse{}), false)
		// buildModelDefinition(d, ed.ResponseType, false)
		// buildModelDefinition(d1, reflect.TypeOf(response.ResponseError{}), false)
		// mm:=mergeMaps(d,d1)

		// fmt.Println("d value is:", mm)
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
	// fmt.Println("Starting of buildModelDefinition")
	// fmt.Println("defs: ", defs)
	// fmt.Println("t: ", t)
	// fmt.Println("t kind: ", t.Kind())
	// fmt.Println("t Name: ", t.Name())

	// fmt.Println("isReq: ", isReq)
	// fmt.Println("Ends of buildModelDefinition")

	if t == typlect.TypeNoParam {
		return
	}

	//fmt.Println("t NumIn: ", t.NumIn())
	//fmt.Println("t Numout: ", t.NumOut())

	if t.Kind() == reflect.Slice {
		//fmt.Println("t elem: ", t.Elem())
		t = t.Elem()
	}

	if t.Kind() == reflect.Pointer {
		//fmt.Println("t elem: ", t.Elem())
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return
	}

	// typeMapping := map[reflect.Type]reflect.Type{
	//     reflect.TypeOf(sql.NullString{}): reflect.TypeOf(""),
	//     reflect.TypeOf(sql.NullInt64{}):  reflect.TypeOf(0),
	//     reflect.TypeOf(sql.NullFloat64{}): reflect.TypeOf(0.0),
	//     reflect.TypeOf(sql.NullBool{}):   reflect.TypeOf(true),
	// }

	var smr []string
	smp := m{}
	for i := 0; i < t.NumField(); i++ {

		//fmt.Println("i: ", i)
		//fmt.Println("t.NumField(): ", t.NumField())

		var (
			f = t.Field(i)

			ft = f.Type
		)
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
		return strings.Split(tag, ",")[0] // ignore ',omitempty'
	}

	return f.Name
}
