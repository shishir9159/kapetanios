package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"net/http"
)

type response struct {
	statusCode int32
}

func certRenewal(c *fiber.Ctx) error {

	matchLabels := map[string]string{"": ""}

	// certificate
	labelSelector := metav1.LabelSelector{MatchLabels: matchLabels}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	// refactor

	client, err := orchestration.NewClient()
	if err != nil {
		fmt.Printf("Error creating Kubernetes client: %v\n", err)
	}

	nodes, err := client.Clientset().CoreV1().Nodes().List(context.Background(), listOptions)
	if err != nil {

	}

	if len(nodes.Items) == 0 {
		//
	}

	for _, node := range nodes.Items {
		// call cert for each nodes individually
		Cert("default", node.Name)
	}

	r := response{
		statusCode: http.StatusOK, // http collision with rpc version of http
	}
	j, err := json.Marshal(r)
	if err != nil {

	}

	// return c.JSON(fiber.Map{"status":"success"})
	return c.JSON(j)
}

func setupRoutes(app *fiber.App) {

	api := app.Group("/cert", logger.New())
	//minorUpgrade := app.Group("minor-upgrade")

	app.Get("/swagger", swagger.HandlerDefault)
	app.Get("/renewal", certRenewal)
	api.Group("")
	//	.SetupRoutes(cert)
	//	.SetupRoutes(minorUpgrade)

}

// decide if you need a seperate router folder or not. more like you are gonna need it
func SetupGroupRoutes(router fiber.Router) {

}

func main() {

	app := fiber.New()
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, HEAD, PUT, PATCH, POST, DELETE",
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
