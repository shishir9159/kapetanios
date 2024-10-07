package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/swagger"
	"net/http"
)

var (
	certRenewalNamespace = "default"
	//	TODO: or should it be klovercloud with additional service accounts?
)

func certRenewal(c *fiber.Ctx) error {

	go Cert(certRenewalNamespace)

	return c.JSON(fiber.Map{"status": http.StatusOK})
}

func setupRoutes(app *fiber.App) {

	//api := app.Group("/cert", logger.New())
	//minorUpgrade := app.Group("minor-upgrade")

	app.Get("/swagger", swagger.HandlerDefault)
	app.Get("/renewal", certRenewal)
	//api.Group("")
	//	.SetupRoutes(cert)
	//	.SetupRoutes(minorUpgrade)

}

// decide if you need a seperate router folder or not. more like you are gonna need it
func SetupGroupRoutes(router fiber.Router) {

}

func main() {

	app := fiber.New()

	//app.Use(logger.New())
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

	err := app.Listen(":8080")
	if err != nil {
		return
	}

	// minor upgrade

	//	what happens when lighthouse fails in the middle of the cert renewal process?

}
