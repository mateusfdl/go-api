package http

type Controller interface {
	RegisterRoutes()
}

// Register all routes for the given controllers
func RegisterRoutes(controllers ...Controller) {
	for _, controller := range controllers {
		controller.RegisterRoutes()
	}
}
