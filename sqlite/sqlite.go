package sqlite

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	// Expose sqlite3 to database/sql
	_ "github.com/mattn/go-sqlite3"
)

// BindRoutes bind SQLite routes to REST client
func BindRoutes(router *gin.RouterGroup, connString string) {
	db, err := sqlx.Open("sqlite3", connString)
	if err != nil {
		log.Fatal(err)
	}

	store := StatementStore{
		DB: db,
	}

	defer store.Free()

	router.POST("/exec", func(c *gin.Context) {
		var body BodyQuery
		c.BindJSON(&body)

		stmt := body.Prepare(&store)

		_, err := stmt.Exec(body.Params...)
		if err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusCreated, gin.H{})
	})

	router.POST("/query", func(c *gin.Context) {
		var body BodyQuery
		c.BindJSON(&body)

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
		c.JSON(http.StatusOK, gin.H{
			"result": result,
		})
	})

	// Return ndjson stream
	router.POST("/stream", func(c *gin.Context) {
		var body BodyQuery
		c.BindJSON(&body)

		stmt := body.Prepare(&store)

		rows, err := stmt.Queryx(body.Params...)
		if err != nil {
			log.Fatal(err)
		}

		c.Stream(func(w io.Writer) bool {
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

				w.Write(json)
				w.Write([]byte("\n"))
			}

			return false
		})
	})
}
