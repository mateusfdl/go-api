package http

type Controller interface {
	RegisterRoutes()
}

func RegisterRoutes(controllers ...Controller) {
	for _, controller := range controllers {
		controller.RegisterRoutes()
	}
}
