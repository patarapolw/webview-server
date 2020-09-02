package server

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	conf "github.com/patarapolw/webview-server/config"
)

// CreateServer create server with custom handlers
func CreateServer() *Handlers {
	config := conf.Get()
	if !config.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	app := gin.Default()

	if config.Token != "" {
		app.Use(func(c *gin.Context) {
			if c.Request.URL.Path == "/" {
				http.SetCookie(c.Writer, &http.Cookie{
					Name:     "token",
					Value:    url.QueryEscape(config.Token),
					Path:     "/",
					SameSite: http.SameSiteStrictMode,
					Secure:   false,
					HttpOnly: true,
				})
			}
			c.Next()
		})
	}

	app.Use(static.Serve("/", static.LocalFile(config.Www, true)))

	apiRouter := app.Group("/api", func(c *gin.Context) {
		isAuth := true

		if config.Token != "" {
			isAuth = false

			if cookie, err := c.Cookie("token"); err == nil {
				if cookie == config.Token {
					isAuth = true
				}
			}
		}

		if !isAuth {
			var header struct {
				Authorization string
			}

			if err := c.ShouldBindHeader(&header); err == nil {
				p := strings.Split(header.Authorization, " ")
				if len(p) == 2 && p[0] == "Bearer" && p[1] == config.Token {
					isAuth = true
				}
			}
		}

		if !isAuth {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	})

	return &Handlers{
		Config: config,
		Server: app,
		API:    apiRouter,
	}
}

// Handlers handlers for server
type Handlers struct {
	Config *conf.Config
	Server *gin.Engine
	API    *gin.RouterGroup
}

// Serve convenient method for serving HTTP
func (h *Handlers) Serve() error {
	return h.Server.Run("localhost:" + strconv.Itoa(h.Config.Port))
}
