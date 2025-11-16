package swagger

import (
	"MgApplication/api-server/util/module"

	"go.uber.org/fx"
)

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
