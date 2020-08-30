package main

/*
#cgo darwin LDFLAGS: -framework CoreGraphics
#cgo linux pkg-config: x11

#if defined(__APPLE__)
#include <CoreGraphics/CGDisplayConfiguration.h>
int display_width() {
	return CGDisplayPixelsWide(CGMainDisplayID());
}
int display_height() {
	return CGDisplayPixelsHigh(CGMainDisplayID());
}
#elif defined(_WIN32)
#include <wtypes.h>
int display_width() {
	RECT desktop;
	const HWND hDesktop = GetDesktopWindow();
	GetWindowRect(hDesktop, &desktop);
	return desktop.right;
}
int display_height() {
	RECT desktop;
	const HWND hDesktop = GetDesktopWindow();
	GetWindowRect(hDesktop, &desktop);
	return desktop.bottom;
}
#else
#include <X11/Xlib.h>
int display_width() {
	Display* d = XOpenDisplay(NULL);
	Screen*  s = DefaultScreenOfDisplay(d);
	return s->width;
}
int display_height() {
	Display* d = XOpenDisplay(NULL);
	Screen*  s = DefaultScreenOfDisplay(d);
	return s->height;
}
#endif
*/
import "C"
import (
	"io/ioutil"
	"log"
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

	if (config.Size == WindowSize{}) {
		config.Size = WindowSize{
			Width:  int(C.display_width()),
			Height: int(C.display_height()),
		}

		// Current method of getting position in linux makes it fall offscreen
		if runtime.GOOS == "linux" {
			config.Size.Height = config.Size.Height - 100
		}
	}

	width := config.Size.Width
	height := config.Size.Height

	if width == 0 || height == 0 {
		width = 1024
		height = 768
	}

	w := webview.New(debug)
	defer OnExit()
	defer w.Destroy()

	w.SetTitle(config.Title)
	w.SetSize(width, height, webview.HintNone)

	w.Bind("setTitle", func(title string) string {
		w.SetTitle(title)
		return title
	})

	w.Init(`
	window.onload = () => {
		var titleEl = document.querySelector("title");
		if (titleEl) {
			setTitle(titleEl.innerText);
		}
	}
	`)

	// Catch exit signals and always execute OnExit
	// including os.Interrupt, SIGINT and SIGTERM
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signals
		OnExit()
		w.Terminate()
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
			w.Navigate(url)
		})
	}()

	w.Run()
}
