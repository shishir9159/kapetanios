package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/swagger"
	"net/http"
)

var (
	// TODO: or should it be klovercloud with additional service accounts?
	certRenewalNamespace  = "default"
	minorUpgradeNamespace = "default"
)

func certRenewal(c *fiber.Ctx) error {

	go Cert(certRenewalNamespace)

	return c.JSON(fiber.Map{"status": http.StatusOK})
}

func cleanup(c *fiber.Ctx) error {

	go Cleanup(certRenewalNamespace)

	return c.JSON(fiber.Map{"status": http.StatusOK})
}

func expiration(c *fiber.Ctx) error {

	go Expiration(certRenewalNamespace)

	return c.JSON(fiber.Map{"status": http.StatusOK})
}

func minorUpgrade(c *fiber.Ctx) error {

	go MinorUpgradeFirstRun(minorUpgradeNamespace)

	//go MinorUpgradeFirstRun(minorUpgradeNamespace)

	return c.JSON(fiber.Map{"status": http.StatusOK})
}

func rollback(c *fiber.Ctx) error {

	go Rollback(certRenewalNamespace)

	return c.JSON(fiber.Map{"status": http.StatusOK})
}

func setupRoutes(app *fiber.App) {

	//api := app.Group("/cert", logger.New())
	//minorUpgrade := app.Group("minor-upgrade")

	// Provide a minimal config for liveness check
	//app.Get(healthcheck.DefaultLivenessEndpoint, healthcheck.NewHealthChecker())
	//// Provide a minimal config for readiness check
	//app.Get(healthcheck.DefaultReadinessEndpoint, healthcheck.NewHealthChecker())
	//// Provide a minimal config for startup check
	//app.Get(healthcheck.DefaultStartupEndpoint, healthcheck.NewHealthChecker())
	// Provide a minimal config for check with custom endpoint
	app.Get("/live", healthcheck.New())

	//// Or extend your config for customization
	//app.Get(healthcheck.DefaultLivenessEndpoint, healthcheck.NewHealthChecker(healthcheck.Config{
	//	Probe: func(c fiber.Ctx) bool {
	//		return true
	//	},
	//}))
	//// And it works the same for readiness, just change the route
	//app.Get(healthcheck.DefaultReadinessEndpoint, healthcheck.NewHealthChecker(healthcheck.Config{
	//	Probe: func(c fiber.Ctx) bool {
	//		return true
	//	},
	//}))
	//// And it works the same for startup, just change the route
	//app.Get(healthcheck.DefaultStartupEndpoint, healthcheck.NewHealthChecker(healthcheck.Config{
	//	Probe: func(c fiber.Ctx) bool {
	//		return true
	//	},
	//}))
	//// With a custom route and custom probe
	//app.Get("/live", healthcheck.NewHealthChecker(healthcheck.Config{
	//	Probe: func(c fiber.Ctx) bool {
	//		return true
	//	},
	//}))

	app.Get("/renewal", certRenewal)
	app.Get("/cleanup", cleanup)
	app.Get("/expiration", expiration)
	app.Get("/minor-upgrade", minorUpgrade)
	app.Get("/rollback", rollback)
	app.Get("/swagger", swagger.HandlerDefault)

	//api.Group("")
	//	.SetupRoutes(cert)
	//	.SetupRoutes(minorUpgrade)
}

// decide if you need a separate router folder or not. more like you are gonna need it

func SetupGroupRoutes(router fiber.Router) {

}

func main() {

	app := fiber.New()

	//logger, err := zap.NewProduction()

	//if err != nil {
	//}

	//app.Use(fiberzap.New(fiberzap.Config{
	//	Logger: logger,
	//}))

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, HEAD, PUT, PATCH, POST, DELETE",
	}))

	// TODO:
	//  Controller Definition need to be moved with the
	//  initial Setup and making sure there exists only one

	// setup routes
	setupRoutes(app)

	//app.Use("/livez", healthcheck.New())

	//app.Use(healthcheck.New(healthcheck.Config{
	//	LivenessProbe: func(c *fiber.Ctx) bool {
	//		return true
	//	},
	//	LivenessEndpoint: "/livez",
	//	ReadinessProbe: func(c *fiber.Ctx) bool {
	//		return true
	//	},
	//	ReadinessEndpoint: "/readyz",
	//}))

	Prerequisites(minorUpgradeNamespace)

	// TODO: prevent duplicate lighthouse instances
	// TODO: websocket
	err := app.Listen(":80")

	if err != nil {
		return
	}

}
