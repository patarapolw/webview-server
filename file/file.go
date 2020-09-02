package file

import (
	"io"
	"io/ioutil"
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
			if os.IsNotExist(err) {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}

			panic(err)
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
			panic(err)
		}
		err = ioutil.WriteFile(qs.Filepath(root), data, 0666)
		if err != nil {
			panic(err)
		}

		c.AbortWithStatus(http.StatusCreated)
	})

	router.DELETE("/", func(c *gin.Context) {
		var qs querystring
		c.BindQuery(&qs)

		err := os.Remove(qs.Filepath(root))
		if err != nil {
			if os.IsNotExist(err) {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}

			panic(err)
		}

		c.AbortWithStatus(http.StatusCreated)
	})
}

type querystring struct {
	Filename string `form:"filename" binding:"required"`
}

// Filepath get filepath from querystring
func (qs *querystring) Filepath(root string) string {
	return path.Join(root, qs.Filename)
}
