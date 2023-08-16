package handlers

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/SamBithrey/spotlas-exercise/database"
	"github.com/gofiber/fiber/v2"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type Spot struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	Website     any     `json:"website"` // typed as any to allow for null values
	Coordinates string  `json:"coordinates"`
	Description any     `json:"description"` // typed as any to allow for null values
	Rating      float64 `json:"rating"`
}

type Spots struct {
	Spots []Spot `json:"spots"`
}

func Healthcheck(c *fiber.Ctx) error {
	return c.SendString("OK")
}

func ReturnAll(c *fiber.Ctx) error {
	var db = database.Get()

	sql, _, _ := psql.Select("*").From("\"MY_TABLE\"").ToSql()
	rows, err := db.Query(sql)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer rows.Close()
	result := Spots{}

	for rows.Next() {
		spot := Spot{}
		if err := rows.Scan(&spot.Id, &spot.Name, &spot.Website, &spot.Coordinates, &spot.Description, &spot.Rating); err != nil {
			return err
		}
		result.Spots = append(result.Spots, spot)
	}

	return c.JSON(result.Spots)
}

func ReturnSelection(c *fiber.Ctx) error {
	var db = database.Get()

	queries := c.Queries()
	if queries["lon"] == "" || queries["lat"] == "" || queries["radius"] == "" {
		return c.Status(400).SendString("Please input your queries")
	}

	type DistanceParams struct {
		Lat    float64 `query:"lat"`
		Lon    float64 `query:"lon"`
		Radius int     `query:"radius"`
		Shape  string  `query:"shape"`
	}

	p := new(DistanceParams)
	p.Shape = "circle"

	if err := c.QueryParser(p); err != nil {
		return c.Status(400).SendString(err.Error())
	}

	result := Spots{}

	if p.Shape == "circle" {

		filter := fmt.Sprintf("ST_DWithin(coordinates, 'SRID=4326;POINT(%v %v)'::geography, %v)", p.Lon, p.Lat, p.Radius)
		order := fmt.Sprintf("ST_Distance(coordinates, 'SRID=4326;POINT(%v %v)'::geography)", p.Lon, p.Lat)

		sql, _, _ := psql.Select("*").From("\"MY_TABLE\"").Where(filter).OrderBy(order).ToSql()

		rows, err := db.Query(sql)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		defer rows.Close()

		for rows.Next() {
			spot := Spot{}
			if err := rows.Scan(&spot.Id, &spot.Name, &spot.Website, &spot.Coordinates, &spot.Description, &spot.Rating); err != nil {
				return err
			}

			result.Spots = append(result.Spots, spot)
		}

	} else if p.Shape == "square" {

		filter := fmt.Sprintf("coordinates && ST_Buffer('SRID=4326;POINT(%v %v)'::geography, %v, 'quad_segs=1')", p.Lon, p.Lat, p.Radius)
		order := fmt.Sprintf("ST_Distance(coordinates, 'SRID=4326;POINT(%v %v)'::geography)", p.Lon, p.Lat)

		sql, _, _ := psql.Select("*").From("\"MY_TABLE\"").Where(filter).OrderBy(order).ToSql()

		rows, err := db.Query(sql)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		defer rows.Close()

		for rows.Next() {
			spot := Spot{}
			if err := rows.Scan(&spot.Id, &spot.Name, &spot.Website, &spot.Coordinates, &spot.Description, &spot.Rating); err != nil {
				return err
			}

			result.Spots = append(result.Spots, spot)
		}

	} else {
		return c.Status(500).JSON(`Error:Please input correct shape`)
	}

	return c.JSON(result.Spots)
}
