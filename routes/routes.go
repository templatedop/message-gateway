package routes

// var isShuttingDown atomic.Value

// func init() {
// 	isShuttingDown.Store(false)
// }

// // SetIsShuttingDown is an exported function that allows other packages to update the isShuttingDown value
// func SetIsShuttingDown(shuttingDown bool) {
// 	isShuttingDown.Store(shuttingDown)
// }

// func HealthCheckHandler(c *gin.Context) {
// 	shuttingDown := isShuttingDown.Load().(bool)
// 	if shuttingDown {
// 		// If the server is shutting down, respond with Service Unavailable
// 		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy"})
// 		return
// 	}
// 	// If the server is not shutting down, respond with OK
// 	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
// }

// func Routes(router *r.Router,
// 	// msgappHandler *handler.MgApplicationHandler,
// 	appHandler *handler.ApplicationHandler,
// 	// providerHandler *handler.ProviderHandler,
// 	// templateHandler *handler.TemplateHandler,
// 	// reportsHandler *handler.ReportsHandler,
// ) {

// 	router.NoRoute(func(c *gin.Context) {
// 		apierrors.HandleNoRouteError(c)
// 	})

// 	router.NoMethod(func(c *gin.Context) {
// 		apierrors.HandleNoMethodError(c)
// 	})

// 	//add subroutes.

// 	//healthz URL
// 	router.GET("/healthz", HealthCheckHandler)

// 	v1 := router.Group("/v1")
// 	{
// 		// Swagger
// 		v1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

// 		Application := v1.Group("/applications") //.Use(authMiddleware(token))
// 		{
// 			// Application.POST("", appHandler.CreateMessageApplicationHandler)
// 			Application.GET("", appHandler.ListMessageApplicationsHandler)
// 			// Application.GET("/:application-id", appHandler.FetchApplicationHandler)
// 			// Application.PUT("/:application-id", appHandler.UpdateMessageApplicationHandler)
// 			// Application.PUT("/:application-id/status", appHandler.ToggleApplicationStatusHandler)
// 		}

// 		// Provider := v1.Group("/sms-providers")
// 		// {
// 		// 	Provider.POST("", providerHandler.CreateMessageProviderHandler)
// 		// 	Provider.GET("", providerHandler.ListMessageProvidersHandler)
// 		// 	Provider.GET("/:provider-id", providerHandler.FetchMessageProviderHandler)
// 		// 	Provider.PUT("/:provider-id", providerHandler.UpdateMessageProviderHandler)
// 		// 	Provider.PUT("/:provider-id/status", providerHandler.ToggleMessageProviderStatusHandler)
// 		// }

// 		// Template := v1.Group("/sms-templates")
// 		// {
// 		// 	Template.POST("", templateHandler.CreateTemplateHandler)
// 		// 	Template.GET("", templateHandler.ListTemplatesHandler)
// 		// 	Template.GET("/:template-local-id", templateHandler.FetchTemplateHandler)
// 		// 	Template.GET("/name", templateHandler.FetchTemplateByApplicationHandler) //by appID query param
// 		// 	Template.GET("/details", templateHandler.FetchTemplateDetailsHandler)    //takes query param, by template-format is yet to be tested
// 		// 	Template.PUT("/:template-local-id/status", templateHandler.ToggleTemplateStatusHandler)
// 		// 	Template.PUT("/:template-local-id", templateHandler.UpdateTemplateHandler)
// 		// }

// 		// v1.POST("/msgrequest/create", msgappHandler.CreateSMSRequestHandler)
// 		// v1.POST("/sms-request", msgappHandler.CreateSMSRequestHandler)
// 		// v1.POST("/sms-request-kafka", msgappHandler.CreateSMSRequestHandlerKafka)

// 		// v1.POST("/test-sms-request", msgappHandler.CreateTestSMSHandler)

// 		// v1.GET("/sms-delivery-status", msgappHandler.FetchCDACSMSDeliveryStatusHandler) //CDAC Delivery report

// 		// //reports
// 		// v1.GET("/sms-dashboard", reportsHandler.SMSDashboardHandler)
// 		// v1.GET("/sms-sent-status-report", reportsHandler.SentSMSStatusReportHandler)
// 		// v1.GET("/aggregate-sms-report", reportsHandler.AggregateSMSUsageReportHandler)

// 		// v1.POST("/bulk-sms-initiate", msgappHandler.InitiateBulkSMSHandler)
// 		// v1.GET("/bulk-sms-validate-otp", msgappHandler.ValidateTestSMSHandler)
// 		// v1.POST("/bulk-sms", msgappHandler.SendBulkSMSHandler)
// 	}
// }
