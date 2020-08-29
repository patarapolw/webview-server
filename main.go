package main

/*
#if defined(__APPLE__)
#cgo LDFLAGS: -framework CoreGraphics
#include <CoreGraphics/CGDisplayConfiguration.h>
int display_width() {
	return CGDisplayPixelsWide(CGMainDisplayID());
}
int display_height() {
	return CGDisplayPixelsHigh(CGMainDisplayID());
}
#endif
*/
import "C"
import (
	"encoding/json"
	"io/ioutil"
	"fmt"
	"time"
	"net/http"
	"syscall"
	"os/signal"
	"log"
	"os"
	"net"
	"runtime"

	"github.com/webview/webview"
)

func main() {
	config := Config{
		Title: os.Getenv("TITLE"),
		Port: os.Getenv("PORT"),
		Path: os.Getenv("PATH"),
		Debug: os.Getenv("DEBUG"),
	}

	if data, err := ioutil.ReadFile("./config.json"); err != nil {
		// Discard errors
		json.Unmarshal(data, &config)
	}

	if config.Port == "" {
		config.Port = "0"
	}

	listener, err := net.Listen("tcp", "localhost:"+config.Port)
	if err != nil {
		log.Fatal(err)
	}

	width := 1024
	height := 768

	if runtime.GOOS == "darwin" {
		width = int(C.display_width())
		height = int(C.display_height())
	}

	debug := false
	if config.Debug != "" {
		debug = true
	}

	w := webview.New(webview.Settings{
		Title: config.Title,
		Debug: debug,
		Width: width,
		Height: height,
		Resizable: true,
	})

	if runtime.GOOS != "darwin" {
		w.SetFullscreen(true)
	}

	// Catch exit signals and always execute OnExit
	// including os.Interrupt, SIGINT and SIGTERM
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signals
		OnExit()
		w.Exit()
	}()

	url := "http://" + listener.Addr().String()

	server := CreateServer()

	go func() {
		log.Println("Listening at:", url)
		if err := server.Serve(listener); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	w.Dispatch(func() {
		for {
			time.Sleep(1 * time.Second)
			_, err := http.Head(url)
			if err == nil {
				break
			}
		}
		w.Eval(fmt.Sprintf("location.href = '%s'", url))
	})

	w.Run()
	OnExit()
}
