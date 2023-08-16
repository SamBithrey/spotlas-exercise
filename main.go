package main

import (
	"fmt"
	"log"

	"github.com/SamBithrey/spotlas-exercise/database"
	"github.com/SamBithrey/spotlas-exercise/handlers"
	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
)

func main() {
	app := fiber.New()
	fmt.Println("Server running OK!")

	if err := database.Connect(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connection made to DB")

	app.Get("/healthcheck", handlers.Healthcheck)

	app.Get(`/all`, handlers.ReturnAll)

	// http://localhost:4000/distance?lat=51&lon=0.0&radius=1000&shape=circle
	app.Get("/distance", handlers.ReturnSelection)

	log.Fatal(app.Listen(":4000"))
}
