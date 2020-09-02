package desktop

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"time"

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
	"C"

	"github.com/patarapolw/webview-server/server"
	"github.com/zserge/lorca"
)
import "github.com/patarapolw/webview-server/config"

// Init initialize desktop app based on server
func Init(hs *server.Handlers, cleanup ...func()) {
	if lorca.LocateChrome() == "" {
		OpenBrowser("https://github.com/patarapolw/webview-server/blob/master/deps.md")
		log.Fatal(fmt.Errorf("cannot open outside Chrome desktop application"))
	} else {
		if (hs.Config.Size == config.WindowSize{}) {
			hs.Config.Size = config.WindowSize{
				Width:  int(C.display_width()),
				Height: int(C.display_height()),
			}

			// Current method of getting screen size in linux and windows makes it fall offscreen
			if runtime.GOOS == "linux" || runtime.GOOS == "windows" {
				hs.Config.Size.Width = hs.Config.Size.Width - 50
				hs.Config.Size.Height = hs.Config.Size.Height - 100
			}
		}

		width := hs.Config.Size.Width
		height := hs.Config.Size.Height

		if width == 0 || height == 0 {
			width = 1024
			height = 768
		}

		w, err := lorca.New("data:text/html,<title>Loading...</title>", "", width, height)
		if err != nil {
			log.Fatal(err)
		}

		defer func() {
			for _, c := range cleanup {
				c()
			}
		}()
		defer w.Close()

		go func() {
			log.Println("Listening at:", hs.Config.URL())
			if err := hs.Serve(); err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}()

		go func() {
			for {
				time.Sleep(1 * time.Second)
				_, err := http.Head(hs.Config.URL())
				if err == nil {
					break
				}
			}

			w.Load(hs.Config.URL())
		}()

		<-w.Done()
	}
}

// OpenBrowser open URL in default browser
func OpenBrowser(url string) {
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
