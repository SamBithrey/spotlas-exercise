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

	app.Get("/allspots", func(c *fiber.Ctx) error {
		conn, err := sql.Open("postgres", connectionStr)
		if err != nil {
			panic(err)
		}

		sql, _, _ := psql.Select("name").From("\"MY_TABLE\"").ToSql()
		allspots, err := conn.Query(sql)
		if err != nil {
			panic(err)
		}

		names := make([]string, 0)
		for allspots.Next() {
			var name string
			if err := allspots.Scan(&name); err != nil {
				log.Fatal(err)
			}
			names = append(names, name)
		}

		conn.Close()

		return c.JSON(names)
	})

	type DistanceParams struct {
		Lat    float64 `query:"lat"`
		Lon    float64 `query:"lon"`
		Radius int     `query:"radius"`
		Shape  string  `query:"shape"`
	}

	type Spot struct {
		Id          string  `json:"id"`
		Name        string  `json:"name"`
		Website     any     `json:"website"`
		Coordinates string  `json:"coordinates"`
		Description any     `json:"description"`
		Rating      float64 `json:"rating"`
	}

	app.Get("/distance", func(c *fiber.Ctx) error {
		p := new(DistanceParams)
		p.Shape = "circle"

		if err := c.QueryParser(p); err != nil {
			return err
		}

		if p.Lat == 0 || p.Lon == 0 {
			return c.Status(404).JSON(`Error:Please input coordinates`)
		}
		if p.Radius == 0 {
			return c.Status(404).JSON(`Error:Please input radius`)
		}

		conn, err := sql.Open("postgres", connectionStr)
		if err != nil {
			panic(err)
		}

		lat := fmt.Sprintf("%v", p.Lat)
		lon := fmt.Sprintf("%v", p.Lon)
		spots := []Spot{}

		if p.Shape == "circle" {
			filter := fmt.Sprintf("ST_DWithin(coordinates, 'SRID=4326;POINT(%v %v)'::geography, %v)", lon, lat, p.Radius)
			order := fmt.Sprintf("ST_Distance(coordinates, 'SRID=4326;POINT(%v %v)'::geography)", lon, lat)

			sql, _, _ := psql.Select("*").From("\"MY_TABLE\"").Where(filter).OrderBy(order).ToSql()
			var spot Spot
			rows, err := conn.Query(sql)
			if err != nil {
				panic(err)
			}

			for rows.Next() {
				spot.Id = ""
				spot.Name = ""
				spot.Website = ""
				spot.Coordinates = ""
				spot.Description = ""
				spot.Rating = 0
				if err := rows.Scan(&spot.Id, &spot.Name, &spot.Website, &spot.Coordinates, &spot.Description, &spot.Rating); err != nil {
					log.Fatal(err)
				}
				spots = append(spots, spot)
			}

		} else if p.Shape == "square" {
			filter := fmt.Sprintf("coordinates && ST_Buffer('SRID=4326;POINT(%v %v)'::geography, %v, 'quad_segs=1')", lon, lat, p.Radius)
			order := fmt.Sprintf("ST_Distance(coordinates, 'SRID=4326;POINT(%v %v)'::geography)", lon, lat)

			sql, _, _ := psql.Select("*").From("\"MY_TABLE\"").Where(filter).OrderBy(order).ToSql()
			var spot Spot
			rows, err := conn.Query(sql)
			if err != nil {
				panic(err)
			}

			for rows.Next() {
				spot.Id = ""
				spot.Name = ""
				spot.Website = ""
				spot.Coordinates = ""
				spot.Description = ""
				spot.Rating = 0
				if err := rows.Scan(&spot.Id, &spot.Name, &spot.Website, &spot.Coordinates, &spot.Description, &spot.Rating); err != nil {
					log.Fatal(err)
				}
				spots = append(spots, spot)
			}

		} else {
			return c.Status(404).JSON(`Error:Please input correct shape`)
		}

		return c.JSON(spots)
	})

	log.Fatal(app.Listen(":4000"))
}
