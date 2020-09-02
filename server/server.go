package server

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	conf "github.com/patarapolw/webview-server/config"
)

// CreateServer create server with custom handlers
func CreateServer() Handlers {
	config := conf.Get()

	if !config.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	app := gin.Default()

	if config.Token != "" {
		app.Use(func(c *gin.Context) {
			log.Println(c.Request.URL.RawPath)
			if c.Request.URL.RawPath == "/" {
				http.SetCookie(c.Writer, &http.Cookie{
					Name:     "token",
					Value:    url.QueryEscape(config.Token),
					Path:     "/",
					Domain:   config.URL(),
					SameSite: http.SameSiteStrictMode,
					Secure:   false,
					HttpOnly: true,
				})
			}
		})
	}

	app.Use(static.Serve("/", static.LocalFile(config.Www, true)))

	apiRouter := app.Group("/api", func(c *gin.Context) {
		isAuth := true

		if config.Token != "" {
			isAuth = false

			if cookie, err := c.Cookie("token"); err != nil {
				if cookie == config.Token {
					isAuth = true
				}
			}
		}

		if !isAuth {
			var header struct {
				Authorization string
			}

			if err := c.ShouldBindHeader(&header); err != nil {
				p := strings.Split(header.Authorization, " ")
				if len(p) == 2 && p[0] == "Bearer" && p[1] == config.Token {
					isAuth = true
				}
			}
		}

		if !isAuth {
			c.JSON(http.StatusUnauthorized, gin.H{})
		}
	})

	return Handlers{
		Config: config,
		Root:   app,
		API:    apiRouter,
	}
}

// Handlers handlers for server
type Handlers struct {
	Config *conf.Config
	Root   *gin.Engine
	API    *gin.RouterGroup
}

// Server get server for handlers
func (h *Handlers) Server() *http.Server {
	return &http.Server{
		Handler: h.Root,
	}
}

// Serve convenient method for serving HTTP
func (h *Handlers) Serve() error {
	return h.Server().Serve(h.Config.Listener)
}
