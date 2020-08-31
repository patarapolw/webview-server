package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	/*
	   #cgo darwin LDFLAGS: -framework CoreGraphics
	   #cgo linux pkg-config: x11

	   #include <stdlib.h>

	   extern int OnExit();

	   void _cleanup() {
	   	OnExit();
	   }

	   void set_cleanup() {
	   	atexit(_cleanup);
	   }

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
	"C"

	conf "github.com/patarapolw/webview-server/config"
	"github.com/patarapolw/webview-server/server"
	"github.com/zserge/lorca"
	"golang.org/x/crypto/ssh/terminal"
)
import (
	"os/signal"
	"syscall"
)

func main() {
	config := conf.Get()

	if os.Getenv("WINDOW") == "" {
		if runInTerminal(config) {
			return
		}
	}

	if lorca.LocateChrome() == "" {
		openBrowser("https://github.com/patarapolw/webview-server/blob/master/deps.md")
		if runInTerminal(config) {
			return
		}
		log.Fatal(fmt.Errorf("cannot open outside Chrome and terminal"))
	} else {
		if (config.Size == conf.WindowSize{}) {
			config.Size = conf.WindowSize{
				Width:  int(C.display_width()),
				Height: int(C.display_height()),
			}

			// Current method of getting screen size in linux and windows makes it fall offscreen
			if runtime.GOOS == "linux" || runtime.GOOS == "windows" {
				config.Size.Width = config.Size.Width - 50
				config.Size.Height = config.Size.Height - 100
			}
		}

		width := config.Size.Width
		height := config.Size.Height

		if width == 0 || height == 0 {
			width = 1024
			height = 768
		}

		w, err := lorca.New("data:text/html,<title>Loading...</title>", "", width, height)
		if err != nil {
			log.Fatal(err)
		}

		defer OnExit()
		defer w.Close()

		server := server.CreateServer(config)

		go func() {
			log.Println("Listening at:", config.URL)
			if err := server.Serve(config.Listener); err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}()

		go func() {
			for {
				time.Sleep(1 * time.Second)
				_, err := http.Head(config.URL)
				if err == nil {
					break
				}
			}

			w.Load(config.URL)
		}()

		<-w.Done()
	}
}

func runInTerminal(config *conf.Config) bool {
	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		server := server.CreateServer(config)

		// Catch exit signals and always execute OnExit
		// including os.Interrupt, SIGINT and SIGTERM
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-signals
			server.Close()
			OnExit()
		}()

		log.Println("Listening at:", config.URL)
		server.Serve(config.Listener)

		return true
	}

	return false
}

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}
