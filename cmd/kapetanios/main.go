package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
	"github.com/quic-go/quic-go"
	"net"
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

func sanityChecking(c *fiber.Ctx) error {

	c.Accepts(`shuttle="launched"`)
	c.Status(http.StatusOK)

	return nil
}

func shuttleLaunched(c *fiber.Ctx) error {

	c.Accepts(`sanity="checked"`)
	c.Status(http.StatusOK)

	return nil
}

func setupRoutes(app *fiber.App) {

	app.Get("/readyz", sanityChecking)
	app.Get("/livez", shuttleLaunched)
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

	udpConn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: 1234})
	// ... error handling
	tr := quic.Transport{
		Conn: udpConn,
	}
	ln, err := tr.Listen(tlsConf, quicConf)
	// ... error handling
	for {
		conn, err := ln.Accept()
		// ... error handling
		// handle the connection, usually in a new Go routine
	}

	app := fiber.New()

	//logger, err := zap.NewProduction()

	//if err != nil {
	//}

	//app.Use(recover.New())
	//app.Use(compress.New())

	//app.Use(fiberzap.New(fiberzap.Config{
	//	Logger: logger,
	//}))

	// TODO:
	//  sync.pool for caching mechanism

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
