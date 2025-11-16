package bootstrap

import (
	v1 "MgApplication/gen/smsrequest/v1/MgApplicationconnect"
	handler "MgApplication/handler"
	repo "MgApplication/repo/postgres"

	g "MgApplication/grpc-server"

	server "MgApplication/api-server"
	serverHandler "MgApplication/api-server/handler"

	"go.uber.org/fx"
)

// NewValidatorService add it as part of fx invoke
var Fxvalidator = fx.Module(
	"validator",
	fx.Invoke(handler.NewValidatorService),
)

/*
func HourValidate(f1 validator.FieldLevel) bool {
	timeRegex := regexp.MustCompile(`^([01]\d|2[0-3]):([0-5]\d)$`)
	return timeRegex.MatchString(f1.Field().String())
}
*/

var FxRepo = fx.Module(
	"Repomodule",
	fx.Provide(
		// repo.NewUserRepository,
		// repo.NewMgApplicationRepository,
		repo.NewApplicationRepository,
		// repo.NewProviderRepository,
		// repo.NewTemplateRepository,
		// repo.NewReportsRepository,
	),
)

// var FxHandler = fx.Module(
// 	"Handlermodule",
// 	fx.Provide(
// 		// handler.NewUserHandler,
// 		// handler.NewMgApplicationHandler,
// 		handler.NewApplicationHandler,
// 		// handler.NewProviderHandler,
// 		// handler.NewTemplateHandler,
// 		// handler.NewReportsHandler,
// 		// handler.NewMgApplicationHandlergrpc,
// 	),
// )

// Group tag used for aggregating server handlers via Fx's group injection.
// Fx expects tags in the form key:"value" (e.g., group:"name").
const serverControllersGroupTag = `group:"servercontrollers"`

var FxHandler = fx.Module(
	"Handlermodule",
	// fx.Provide(
	// 	// handler.NewUserHandler,
	// 	// handler.NewMgApplicationHandler,
	// 	handler.NewApplicationHandler,
	// 	// handler.NewProviderHandler,
	// 	// handler.NewTemplateHandler,
	// 	// handler.NewReportsHandler,
	// 	// handler.NewMgApplicationHandlergrpc,
	// ),
	fx.Provide(
		fx.Annotate(
			handler.NewApplicationHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverControllersGroupTag),
		),
	),
)

var FxParseController = fx.Module(
	"ParseControllermodule",
	fx.Provide(
		server.ParseControllers,
	),
)

func AddHandlers(registry *g.HandlerRegistry, msgapplicationhandler *handler.MgApplicationHandlergrpc) {
	registry.AddHandlers([]g.HandlerDefinition{
		{
			Constructor: g.Wrap(v1.NewSMSRequestServiceHandler),
			Server:      msgapplicationhandler,
		},
	})
}
