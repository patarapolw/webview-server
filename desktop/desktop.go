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
	"C"

	"github.com/patarapolw/webview-server/server"
	"github.com/zserge/lorca"
)
import "github.com/patarapolw/webview-server/config"

// Init initialize desktop app based on server
func Init(hs *server.Handlers, cleanup ...func()) {
	if lorca.LocateChrome() == "" {
		OpenBrowser("https://github.com/patarapolw/webview-server/blob/master/deps.md")
		panic(fmt.Errorf("cannot open outside Chrome desktop application"))
	} else {
		width := hs.Config.Size.Width
		height := hs.Config.Size.Height

		if (hs.Config.Size == config.WindowSize{}) {
			width = int(C.display_width())
			height = int(C.display_height())
		}

		if width == 0 || height == 0 {
			width = 1024
			height = 768
		}

		w, err := lorca.New("data:text/html,<title>Loading...</title>", "", width, height)
		if err != nil {
			panic(err)
		}

		if (hs.Config.Size == config.WindowSize{}) {
			w.SetBounds(lorca.Bounds{
				WindowState: lorca.WindowStateMaximized,
			})
		}

		defer func() {
			for _, c := range cleanup {
				c()
			}
		}()
		defer w.Close()

		go func() {
			log.Println("Listening at:", hs.Config.URL)
			if err := hs.Serve(); err != http.ErrServerClosed {
				panic(err)
			}
		}()

		go func() {
			for {
				time.Sleep(1 * time.Second)
				_, err := http.Head(hs.Config.URL)
				if err == nil {
					break
				}
			}

			w.Load(hs.Config.URL)
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
		panic(err)
	}
}
