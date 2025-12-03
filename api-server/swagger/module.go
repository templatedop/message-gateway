package swagger

import (
	"fmt"
	"os"
	"strings"

	"MgApplication/api-server/util/module"

	"go.uber.org/fx"
)

func init() {
	// Check environment variable to determine swagger generation mode
	// SWAGGER_GENERATION_MODE=build     -> BuildMode (load pre-generated)
	// SWAGGER_GENERATION_MODE=runtime   -> RuntimeMode (generate on startup)
	// Default: RuntimeMode
	mode := os.Getenv("SWAGGER_GENERATION_MODE")
	switch strings.ToLower(mode) {
	case "build":
		SetGenerationMode(BuildMode)
		fmt.Println("Swagger: Build mode enabled (loading pre-generated documentation)")
	case "runtime", "":
		SetGenerationMode(RuntimeMode)
		if mode != "" {
			fmt.Println("Swagger: Runtime mode enabled (generating documentation on startup)")
		}
	default:
		fmt.Printf("Warning: Unknown SWAGGER_GENERATION_MODE '%s', using runtime mode\n", mode)
		SetGenerationMode(RuntimeMode)
	}
}

func Module() *module.Module {
	m := module.New("swagger")

	m.Provide(
		buildDocs,
		ginWrapper,
	)
	m.Invoke(generatejson)

	return m
}

var FxGenerateSwagger = fx.Module(
	"swagger",
	fx.Provide(
		buildDocs,
		ginWrapper,
	),
	fx.Invoke(generatejson),
)
