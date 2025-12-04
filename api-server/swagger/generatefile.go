package swagger

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/getkin/kin-openapi/openapi3"
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func generatejson(v3 *openapi3.T) {
	// This function is optional post-processing that creates resolved_swagger.json
	// It's safe to skip if v3Doc.json doesn't exist yet

	// Check if v3Doc.json exists (it's created asynchronously by buildDocs)
	if _, err := os.Stat("./docs/v3Doc.json"); os.IsNotExist(err) {
		// File doesn't exist yet - this is expected on first run or in build mode
		// The async write in buildDocs will create it later
		fmt.Println("Swagger post-processing skipped: v3Doc.json not yet available")
		return
	}

	// Load the Swagger JSON file
	file, err := os.Open("./docs/v3Doc.json")
	if err != nil {
		// Non-fatal: log and return instead of crashing
		fmt.Printf("Warning: Failed to open v3Doc.json for post-processing: %v\n", err)
		return
	}
	defer file.Close()

	// Read the file content
	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("Warning: Failed to read v3Doc.json: %v\n", err)
		return
	}

	// Validate that we have content
	if len(data) == 0 {
		fmt.Println("Warning: v3Doc.json is empty, skipping post-processing")
		return
	}

	// Parse the JSON into a Gabs container
	jsonParsed, err := gabs.ParseJSON(data)
	if err != nil {
		fmt.Printf("Warning: Failed to parse v3Doc.json: %v (file may be corrupted or incomplete)\n", err)
		fmt.Printf("First 200 bytes of file: %s\n", string(data[:min(200, len(data))]))
		return
	}

	// Start by processing components.schemas
	//components := jsonParsed.Path("components.schemas").ChildrenMap()

	// Traverse the entire document to replace $refs with actual schemas
	traverseAndReplaceRefs(jsonParsed, jsonParsed)
	replaceDataType(jsonParsed, "NullString", "string")
	wrap200Responses(jsonParsed)
	nullStringPath := "components.schemas.NullString"
	if jsonParsed.ExistsP(nullStringPath) {
		jsonParsed.DeleteP(nullStringPath)
	}

	// Create docs folder if not available
	if _, err := os.Stat("docs"); os.IsNotExist(err) {
		if err := os.Mkdir("docs", os.ModePerm); err != nil {
			fmt.Printf("Warning: Failed to create docs directory: %v\n", err)
			return
		}
	}

	err = os.WriteFile("./docs/resolved_swagger.json", []byte(jsonParsed.StringIndent("", "  ")), 0644)
	if err != nil {
		fmt.Printf("Warning: Failed to write resolved_swagger.json: %v\n", err)
		return
	}

	fmt.Println("Swagger post-processing completed: resolved_swagger.json created")
}


func replaceDataType(container *gabs.Container, targetType, newType string) {
	//fmt.Println("container details: ", container)	

// if container.Exists("schemas"){
// 		fmt.Println("schema is: ",container.Path("schema"))
// 	}
		


	//get components.schemas	
	// if container.Exists("components") {
	// 	componentsSchemas := container.Path("components.schemas")
	// 	fmt.Println("componentsSchemas:", componentsSchemas)
	// }

	// if container.Exists("components") {
	// 	fmt.Println("checking components schemas nullstring")

	// 	//componentsSchemas := container.Path("components.schemas")
	// 	//fmt.Println("componentsSchemas:", componentsSchemas)
	// 	//replaceDataType(componentsSchemas, targetType, newType)
	// }
	// Traverse the JSON tree
	children,_ := container.ChildrenMap()

	for key, child := range children {
		// Check if this is an object with a NullString property
		if key == "properties" && child.Exists(targetType) {
			// Replace the entire object with a new type
			container.Delete("properties") // Remove "properties" key
			container.Set(newType, "type") // Set "type" to the new type (e.g., "string")
		} else if key == targetType {
			// If "NullString" exists, replace it with the new type
			container.Set(newType, key)
		} else if key == "type" && child.Data().(string) == targetType {
			// If the type is exactly the targetType (e.g., "NullString"), replace it
			container.Set(newType, "type")
		} else if child != nil {
			// }else if key == "schemas" && child.Data().(string) == targetType {
			// 	// If the type is exactly the targetType (e.g., "NullString"), replace it
			// 	fmt.Println("schemas nullstring: ",container.Data())
			// 	//container.Set(newType, "type")
			// }

			// Recurse into nested structures
			replaceDataType(child, targetType, newType)
		}
	}
}

func traverseAndReplaceRefs(container, root *gabs.Container) {
	children,_ := container.ChildrenMap()
	//fmt.Println("children", children)

	for key, child := range children {
		// Check for $ref and replace it with the actual schema
		if key == "$ref" {
			//fmt.Println("came inside $ref")
			refPath := child.Data().(string)
			//fmt.Println("refPath",refPath)
			if schema := resolveSchema(refPath, root); schema != nil {
				// Replace the $ref with the actual schema
				container.Merge(schema)
				container.Delete("$ref")
			}
		} else if child != nil {
			// Recurse into nested structures
			traverseAndReplaceRefs(child, root)
		}
	}
}

func resolveSchema(refPath string, root *gabs.Container) *gabs.Container {
	if strings.HasPrefix(refPath, "#/components/schemas/") {
		// Extract the schema name
		schemaName := refPath[len("#/components/schemas/"):]
		resolved := root.Path(fmt.Sprintf("components.schemas.%s", schemaName))
		if resolved.Exists() {
			return resolved
		}
	}
	return nil
}

func wrap200Responses(container *gabs.Container) {
	paths := container.Path("paths")
	if paths == nil {
		//fmt.Println("No paths found in the Swagger document.")
		return
	}

	// Traverse all paths
	pathsMap,_ := paths.ChildrenMap()
	for _, pathData := range pathsMap {
		methods,_ := pathData.ChildrenMap()

		// Traverse all methods (e.g., GET, POST, etc.)
		for _, methodData := range methods {
			responses := methodData.Path("responses")
			if responses != nil {
				// Check for 200 response
				response200 := responses.Path("200")
				if response200.Exists() {
					// Wrap the existing schema in the SuccessResponse
					existingSchema := response200.Path("content.application/json.schema")
					if existingSchema.Exists() {
						//fmt.Printf("Wrapping 200 response for %s %s\n", method, path)

						// Create the new SuccessResponse wrapper
						successResponse := map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"success": map[string]interface{}{
									"type":    "boolean",
									"example": true,
								},
								"message": map[string]interface{}{
									"type":    "string",
									"example": "success",
								},
								"data": existingSchema.Data(),
							},
						}

						// Replace the original schema with the new SuccessResponse structure
						response200.Set(successResponse, "content", "application/json", "schema")
					}
				}
			}
		}
	}
}