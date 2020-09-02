package server

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/middleware"

	conf "github.com/patarapolw/webview-server/config"
	"github.com/patarapolw/webview-server/file"
	"github.com/patarapolw/webview-server/sqlite"
)

// CreateServer create server with custom handlers
func CreateServer(config *conf.Config) *fiber.App {
	app := fiber.New()
	app.Use(middleware.Recover())
	app.Use(middleware.Logger())
	app.Use(middleware.FileSystem(middleware.FileSystemConfig{
		Next: func(ctx *fiber.Ctx) bool {
			path := ctx.Path()
			if path == "/" || path == "/index.html" {
				ctx.Cookie(&fiber.Cookie{
					Name:     "token",
					Value:    config.Token,
					Path:     "/",
					Domain:   config.URL(),
					SameSite: "strict",
					Secure:   false,
					HTTPOnly: true,
				})
			}

			if strings.HasPrefix(path, "/api/") {
				return true
			}

			return false
		},
		Root:   http.Dir(config.Www),
		Index:  "/index.html",
		Browse: false,
	}))

	apiRouter := app.Group("/api", func(ctx *fiber.Ctx) {
		isAuth := true

		if config.Token != "" {
			isAuth = false

			if ctx.Cookies("token") == config.Token {
				isAuth = true
			}
		}

		if !isAuth {
			p := strings.Split(ctx.Get("Authorization"), " ")
			if len(p) == 2 && p[0] == "Bearer" && p[1] == config.Token {
				isAuth = true
			}
		}

		if !isAuth {
			ctx.Next(fiber.ErrUnauthorized)
			return
		}

		ctx.Next()
	})

	file.BindRoutes(apiRouter.Group("/file"), config.Root)

	if config.Sqlite != "" {
		sqlite.BindRoutes(apiRouter.Group("/sqlite"), config.Sqlite)
	}

	app.Use(func(c *fiber.Ctx) {
		c.Status(fiber.StatusNotFound).SendString("Sorry can't find that!")
	})

	return app
}
