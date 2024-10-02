package main

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"net/http"
	"time"
)

func run(ch chan struct{}) {
	fmt.Println("run")
	time.Sleep(1 * time.Second)
	ch <- struct{}{}
}

func RunForever() {
	wait := make(chan struct{})
	for {
		go run(wait)
		<-wait
	}
}

type response struct {
	statusCode int32
}

func certRenewal(c *fiber.Ctx) error {

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

	setupRoutes(app)

	err := app.Listen(":8080")
	if err != nil {
		return
	}
	// setup routes

	router.SetupRoutes(app)

	matchLabels := map[string]string{"": ""}

	// certificate
	labelSelector := metav1.LabelSelector{MatchLabels: matchLabels}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	nodes, err := client.Clientset().CoreV1().Nodes().List(context.Background(), listOptions)
	if err != nil {
	}

	for _, node := range nodes.Items {
		// call cert for each nodes individually
	}

	Cert("default")
	RunForever()

	// minor upgrade

	//	what happens when lighthouse fails in the middle of the cert renewal process?

}
