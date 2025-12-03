package main

import (
	"fmt"
	"os"

	config "MgApplication/api-config"
	"MgApplication/api-server/swagger"
)

// generate-swagger is a build-time tool to pre-generate swagger documentation
// Usage: go run cmd/generate-swagger/main.go
// This generates ./docs/pregenerated_swagger.json which can be loaded at runtime for faster startup

func main() {
	fmt.Println("=== Swagger Build-Time Generator ===")
	fmt.Println("This tool generates swagger documentation at build time for faster application startup")
	fmt.Println()

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		fmt.Println("Attempting to continue with minimal config...")
		cfg = config.New()
	}

	// TODO: This tool needs to be integrated with your application's controller registration
	// For now, you need to manually provide endpoint definitions or import your controllers

	fmt.Println("⚠️  IMPORTANT: You need to customize this tool to load your application's controllers")
	fmt.Println()
	fmt.Println("Integration steps:")
	fmt.Println("1. Import your application's controller packages")
	fmt.Println("2. Call GetSwaggerDefs with your registered controllers")
	fmt.Println("3. Run this tool during build: go run cmd/generate-swagger/main.go")
	fmt.Println("4. Set SWAGGER_GENERATION_MODE=build in your environment")
	fmt.Println("5. Deploy with ./docs/pregenerated_swagger.json")
	fmt.Println()

	// Example of how to use (requires your controllers to be imported):
	//
	// import (
	//     router "MgApplication/api-server"
	//     // Import all your controller packages here
	//     _ "MgApplication/your-controller-package"
	// )
	//
	// registries := router.ParseControllers(
	//     // your controllers here
	// )
	// eds := router.GetSwaggerDefs(registries)

	// For demonstration, create empty endpoint definitions
	eds := []swagger.EndpointDef{}

	if len(eds) == 0 {
		fmt.Println("❌ No endpoint definitions found")
		fmt.Println("   Please customize this tool to include your controllers")
		os.Exit(1)
	}

	fmt.Printf("Found %d endpoints to document\n", len(eds))

	// Generate swagger documentation
	fmt.Println("Generating swagger documentation...")
	v3Doc := buildDocsForCLI(eds, cfg)
	if v3Doc == nil {
		fmt.Println("❌ Failed to generate swagger documentation")
		os.Exit(1)
	}

	// Save to pre-generated file
	fmt.Println("Saving to pre-generated file...")
	if err := swagger.SavePreGeneratedSwagger(v3Doc); err != nil {
		fmt.Printf("❌ Error saving swagger: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✅ Success! Swagger documentation generated")
	fmt.Printf("📄 File: %s\n", swagger.PreGeneratedSwaggerFile)
	fmt.Println()
	fmt.Println("To use build mode:")
	fmt.Println("1. Set environment variable: SWAGGER_GENERATION_MODE=build")
	fmt.Println("2. Or call swagger.SetGenerationMode(swagger.BuildMode) in your code")
	fmt.Println()
}

func loadConfig() (*config.Config, error) {
	// Try to load from standard locations
	for _, path := range []string{"config.yaml", "config.yml", "../config.yaml", "../../config.yaml"} {
		if _, err := os.Stat(path); err == nil {
			return config.Load(path)
		}
	}
	return nil, fmt.Errorf("config file not found")
}

// buildDocsForCLI is a wrapper around internal swagger generation for CLI use
func buildDocsForCLI(eds []swagger.EndpointDef, cfg *config.Config) *swagger.T {
	// This calls internal swagger generation functions
	// You may need to expose these or copy the logic here
	return nil // Placeholder - implement based on your swagger package structure
}
