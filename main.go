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
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/muhammadmuzzammil1998/jsonc"
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
		Port:  port,
		Path:  os.Getenv("WEBPATH"),
		Debug: debug,
	}

	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	if data, err := ioutil.ReadFile(path.Join(dir, "config.json")); err == nil {
		// Discard errors
		jsonc.Unmarshal(data, &config)
	}

	config.Path = path.Join(dir, config.Path)

	listener, err := net.Listen("tcp", "localhost:"+strconv.Itoa(config.Port))
	if err != nil {
		log.Fatal(err)
	}

	if (config.Size == WindowSize{} && runtime.GOOS == "darwin") {
		config.Size = WindowSize{
			Width:  int(C.display_width()),
			Height: int(C.display_height()),
		}
	}

	width := config.Size.Width
	height := config.Size.Height

	if width == 0 || height == 0 {
		width = 1024
		height = 768
	}

	w := webview.New(webview.Settings{
		Title:     config.Title,
		Debug:     debug,
		Width:     width,
		Height:    height,
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

		time.Sleep(200 * time.Millisecond)

		w.Dispatch(func() {
			w.Bind("webview", &Webview{
				webview: w,
			})

			w.Eval(`
			var titleEl = document.querySelector("title");
			if (titleEl) {
				webview.setTitle(titleEl.innerText);
			}
			`)
		})
	}()

	w.Run()
	OnExit()
}

// Webview struct to be exported to JS
type Webview struct {
	webview webview.WebView
}

// SetTitle set title function
func (w *Webview) SetTitle(title string) error {
	(*w).webview.SetTitle(title)

	return nil
}
