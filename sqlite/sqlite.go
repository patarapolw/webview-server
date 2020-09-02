package sqlite

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gofiber/fiber"
	"github.com/jmoiron/sqlx"

	// Expose sqlite3 to database/sql
	_ "github.com/mattn/go-sqlite3"
)

// BindRoutes bind SQLite routes to REST client
func BindRoutes(router fiber.Router, connString string) {
	db, err := sqlx.Open("sqlite3", connString)
	if err != nil {
		log.Fatal(err)
	}

	store := StatementStore{
		DB: db,
	}

	defer store.Free()

	router.Post("/exec", func(c *fiber.Ctx) {
		var body BodyQuery
		if err := c.BodyParser(&body); err != nil {
			log.Fatal(err)
		}

		stmt := body.Prepare(&store)

		_, err := stmt.Exec(body.Params...)
		if err != nil {
			log.Fatal(err)
		}
		c.SendStatus(http.StatusCreated)
	})

	router.Post("/query", func(c *fiber.Ctx) {
		var body BodyQuery
		if err := c.BodyParser(&body); err != nil {
			log.Fatal(err)
		}

		stmt := body.Prepare(&store)

		rows, err := stmt.Queryx(body.Params...)
		if err != nil {
			log.Fatal(err)
		}

		result := []map[string]interface{}{}

		for rows.Next() {
			var row map[string]interface{}
			err := rows.StructScan(&row)
			if err != nil {
				log.Fatalln(err)
			}

			result = append(result, row)
		}
		c.Send(map[string]interface{}{
			"result": result,
		})
	})

	// Return ndjson stream
	router.Post("/stream", func(c *fiber.Ctx) {
		var body BodyQuery
		if err := c.BodyParser(&body); err != nil {
			log.Fatal(err)
		}

		stmt := body.Prepare(&store)

		rows, err := stmt.Queryx(body.Params...)
		if err != nil {
			log.Fatal(err)
		}

		for rows.Next() {
			var row map[string]interface{}
			err := rows.StructScan(&row)
			if err != nil {
				log.Fatalln(err)
			}

			json, err := json.Marshal(row)
			if err != nil {
				log.Fatalln(err)
			}

			c.Write(json)
			c.Write([]byte("\n"))
		}
	})
}
