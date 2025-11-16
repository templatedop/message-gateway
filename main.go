package main

import (
	"MgApplication/bootstrap"
	"context"

	bootstrapper "MgApplication/api-bootstrapper"
)

// Swagger
//
//	@title			Message Gateway API
//	@version		1.0
//	@description	A comprehensive API for managing addresses, offering endpoints for creation, update, deletion, and retrieval of Message Gateway data
//	@termsOfService	http://cept.gov.in/terms
//	@contact.name	API Support Team
//	@contact.url	http://cept.gov.in/support
//	@contact.email	support_cept@indiapost.gov.in
//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html
//	@host			localhost:8080
//	@BasePath		/v1
//	@schemes		http https
func main() {
	// app := fx.New(
	app := bootstrapper.New().Options(
		// bootstrapper.Fxconfig,
		// bootstrapper.Fxlog,
		// bootstrapper.FxDB,
		// bootstrapper.Fxclient,
		// bootstrap.FxParseController,
		bootstrap.Fxvalidator,
		// bootstrapper.Fxrouter,
		bootstrap.FxHandler,
		bootstrap.FxRepo,
		// fx.Invoke(routes.Routes),
		// bootstrapper.FxGrpc,
		// fx.Invoke(bootstrap.AddHandlers),
	)

	// app.Run()
	app.WithContext(context.Background()).Run()

	// app := bootstrapper.New().Options(
	// 	bootstrapper.Fxclient,
	// 	bootstrap.Fxvalidator,
	// 	bootstrap.FxHandler,
	// 	bootstrap.FxRepo,
	// 	fx.Invoke(routes.Routes),
	// 	bootstrap.FxRepo,
	// 	bootstrapper.FxGrpc,
	// 	fx.Invoke(bootstrap.AddHandlers),

	// )

	// app.WithContext(context.Background()).Run()
}
