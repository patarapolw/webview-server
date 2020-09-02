package file

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gin-gonic/gin"
)

// BindRoutes bind file routes to REST client
func BindRoutes(router *gin.RouterGroup, root string) {
	router.GET("/", func(c *gin.Context) {
		var qs querystring
		c.BindQuery(&qs)

		data, err := ioutil.ReadFile(qs.Filepath(root))
		if err != nil {
			log.Fatal(err)
			return
		}

		c.Stream(func(w io.Writer) bool {
			w.Write(data)
			return false
		})
	})

	router.PUT("/", func(c *gin.Context) {
		var qs querystring
		c.BindQuery(&qs)

		data, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Fatal(err)
			return
		}
		err = ioutil.WriteFile(qs.Filepath(root), data, 0666)
		if err != nil {
			log.Fatal(err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{})
	})

	router.DELETE("/", func(c *gin.Context) {
		var qs querystring
		c.BindQuery(&qs)

		err := os.Remove(qs.Filepath(root))
		if err != nil {
			log.Fatal(err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{})
	})
}

type querystring struct {
	filename string `query:"filename" binding:"required"`
}

// Filepath get filepath from querystring
func (qs *querystring) Filepath(root string) string {
	return path.Join(root, qs.filename)
}
