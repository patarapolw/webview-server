package main

import (
	"log"
	"time"

	"github.com/patarapolw/webview-server/desktop"
	"github.com/patarapolw/webview-server/file"
	"github.com/patarapolw/webview-server/server"
)

func main() {
	handlers := server.CreateServer()

	file.BindRoutes(handlers.API.Group("/file"), handlers.Config.Root)

	// Enable SQLite routes and compile to make it work for you
	// sqlite.BindRoutes(handlers.API.Group("/sqlite"), handlers.Config.Sqlite)

	// It is also possible to run without desktop, via
	/*
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-signals
			OnExit()
		}()

		handlers.Serve()
	*/
	// Or just create your own Gin server

	desktop.Init(handlers, func() {
		log.Println("Executing clean-up function")
		time.Sleep(1 * time.Second)
		log.Println("Clean-up finished")
	})
}
