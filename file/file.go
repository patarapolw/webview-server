package file

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gofiber/fiber"
)

// BindRoutes bind file routes to REST client
func BindRoutes(router fiber.Router, root string) {
	router.Get("/", func(c *fiber.Ctx) {
		var qs querystring
		if err := c.QueryParser(qs); err != nil {
			log.Fatal(err)
		}

		data, err := ioutil.ReadFile(qs.Filepath(root))
		if err != nil {
			log.Fatal(err)
		}

		c.Write(data)
	})

	router.Put("/", func(c *fiber.Ctx) {
		var qs querystring
		if err := c.QueryParser(qs); err != nil {
			log.Fatal(err)
		}

		err := ioutil.WriteFile(qs.Filepath(root), []byte(c.Body()), 0666)
		if err != nil {
			log.Fatal(err)
		}

		c.SendStatus(http.StatusCreated)
	})

	router.Delete("/", func(c *fiber.Ctx) {
		var qs querystring
		if err := c.QueryParser(qs); err != nil {
			log.Fatal(err)
		}

		err := os.Remove(qs.Filepath(root))
		if err != nil {
			log.Fatal(err)
		}

		c.SendStatus(http.StatusCreated)
	})
}

type querystring struct {
	filename string `query:"filename" binding:"required"`
}

// Filepath get filepath from querystring
func (qs *querystring) Filepath(root string) string {
	if qs.filename == "" {
		log.Fatal("query.filename is required")
	}

	return path.Join(root, qs.filename)
}
