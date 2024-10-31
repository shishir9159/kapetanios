package main

import (
	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/swagger"
	"go.uber.org/zap"
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

	app.Get("/minor-upgrade", minorUpgrade)
	app.Get("/renewal", certRenewal)
	app.Get("/cleanup", cleanup)
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

	logger, err := zap.NewProduction()

	if err != nil {
	}

	app.Use(fiberzap.New(fiberzap.Config{
		Logger: logger,
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, HEAD, PUT, PATCH, POST, DELETE",
	}))

	app.Use(healthcheck.New(healthcheck.Config{
		LivenessProbe: func(c *fiber.Ctx) bool {
			return true
		},
		LivenessEndpoint: "/healthz",
		ReadinessProbe: func(c *fiber.Ctx) bool {
			return true
		},
		ReadinessEndpoint: "/healthz",
	}))

	// setup routes
	setupRoutes(app)

	Prerequisites()

	err = app.Listen(":80")
	if err != nil {
		return
	}

	//	what happens to the current state
	//	when lighthouse fails in the middle
	//	of the cert renewal process
	//

}
