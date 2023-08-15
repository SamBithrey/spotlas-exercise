package main

import (
	"database/sql"
	"fmt"
	"log"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
)

func main() {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	app := fiber.New()

	fmt.Println("Server running OK!")

	app.Get("/healthcheck", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	connectionStr := "user=postgres dbname=spotlas-work port=5432 sslmode=disable"

	type DistanceParams struct {
		Lat    float64 `query:"lat"`
		Lon    float64 `query:"lon"`
		Radius int     `query:"radius"`
		Shape  string  `query:"shape"`
	}

	type Spot struct {
		Id          string  `json:"id"`
		Name        string  `json:"name"`
		Website     any     `json:"website"` // typed as any to allow for null values
		Coordinates string  `json:"coordinates"`
		Description any     `json:"description"` // typed as any to allow for null values
		Rating      float64 `json:"rating"`
	}

	app.Get("/distance", func(c *fiber.Ctx) error {
		p := new(DistanceParams)
		// shape param defaults to circle if no value entered.
		p.Shape = "circle"

		// Get the query params and type check using DistanceParams struct typing.
		if err := c.QueryParser(p); err != nil {
			return c.Status(400).JSON(`Error:Incorrect query params`)
		}

		// if no values for latitude and longitude are entered return error.
		if p.Lat == 0 || p.Lon == 0 {
			return c.Status(400).JSON(`Error:Please input coordinates`)
		}
		// if no value for Radius is entered return error.
		if p.Radius == 0 {
			return c.Status(400).JSON(`Error:Please input radius`)
		}

		// Connect to DB
		conn, err := sql.Open("postgres", connectionStr)
		if err != nil {
			panic(err)
		} else {
			fmt.Println("Connection to db made.")
		}

		// Create spots slice.
		spots := []Spot{}
		// Create spot object.
		var spot Spot

		if p.Shape == "circle" {
			// Create Filter string
			filter := fmt.Sprintf("ST_DWithin(coordinates, 'SRID=4326;POINT(%v %v)'::geography, %v)", p.Lon, p.Lat, p.Radius)
			// Create Order string
			order := fmt.Sprintf("ST_Distance(coordinates, 'SRID=4326;POINT(%v %v)'::geography)", p.Lon, p.Lat)
			// Create SQL string
			sql, _, _ := psql.Select("*").From("\"MY_TABLE\"").Where(filter).OrderBy(order).ToSql()
			// Query DB
			rows, err := conn.Query(sql)
			if err != nil {
				panic(err)
			}
			// Iterate through rows and enter data into the spot object
			for rows.Next() {
				if err := rows.Scan(&spot.Id, &spot.Name, &spot.Website, &spot.Coordinates, &spot.Description, &spot.Rating); err != nil {
					log.Fatal(err)
				}
				// Append spot object to spots slice
				spots = append(spots, spot)
			}

		} else if p.Shape == "square" {
			// Create Filter string
			filter := fmt.Sprintf("coordinates && ST_Buffer('SRID=4326;POINT(%v %v)'::geography, %v, 'quad_segs=1')", p.Lon, p.Lat, p.Radius)
			// Create Order string
			order := fmt.Sprintf("ST_Distance(coordinates, 'SRID=4326;POINT(%v %v)'::geography)", p.Lon, p.Lat)
			// Create SQL string
			sql, _, _ := psql.Select("*").From("\"MY_TABLE\"").Where(filter).OrderBy(order).ToSql()
			// Query DB
			rows, err := conn.Query(sql)
			if err != nil {
				panic(err)
			}
			// Iterate through rows and enter data into the spot object
			for rows.Next() {
				if err := rows.Scan(&spot.Id, &spot.Name, &spot.Website, &spot.Coordinates, &spot.Description, &spot.Rating); err != nil {
					log.Fatal(err)
				}
				// Append spot object to spots slice
				spots = append(spots, spot)
			}

		} else {
			// If shape is any other value than "square" or "circle" return error.
			return c.Status(400).JSON(`Error:Please input correct shape`)
		}
		defer conn.Close()
		defer fmt.Println("Connection to db closed")
		return c.JSON(spots)
	})

	log.Fatal(app.Listen(":4000"))
}
