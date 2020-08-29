package main

/*
#cgo darwin LDFLAGS: -framework CoreGraphics
#if defined(__APPLE__)
#include <CoreGraphics/CGDisplayConfiguration.h>
int display_width() {
	return CGDisplayPixelsWide(CGMainDisplayID());
}
int display_height() {
	return CGDisplayPixelsHigh(CGMainDisplayID());
}
#else
int display_width() {
	return 0;
}
int display_height() {
	return 0;
}
#endif
*/
import "C"
import (
	"math/rand"
	"strconv"
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
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	debug := false
	if os.Getenv("DEBUG") != "" {
		debug = true
	}

	config := Config{
		Title: os.Getenv("TITLE"),
		Port: port,
		Path: os.Getenv("WEBPATH"),
		Debug: debug,
	}

	if data, err := ioutil.ReadFile("config.json"); err == nil {
		// Discard errors
		json.Unmarshal(data, &config)
	}

	listener, err := net.Listen("tcp", "localhost:"+strconv.Itoa(config.Port))
	if err != nil {
		log.Fatal(err)
	}

	if (config.Size == WindowSize{} && runtime.GOOS == "darwin") {
		config.Size = WindowSize{
			Width: int(C.display_width()),
			Height: int(C.display_height()),
		}
	}

	w := webview.New(webview.Settings{
		Title: config.Title,
		Debug: debug,
		Width: config.Size.Width,
		Height: config.Size.Height,
		Resizable: true,
	})

	if (config.Size == WindowSize{} && runtime.GOOS != "darwin") {
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

	server := CreateServer(&config)

	go func() {
		log.Println("Listening at:", url)
		if err := server.Serve(listener); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	go func() {
		for {
			time.Sleep(1 * time.Second)
			_, err := http.Head(url)
			if err == nil {
				break
			}
		}

		w.Dispatch(func() {
			w.Eval(fmt.Sprintf("location.href = '%s?v=%d';", url, rand.Int()))
		})
	}()

	w.Run()
	OnExit()
}
