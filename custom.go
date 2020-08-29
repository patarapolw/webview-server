package main

import (
	"io"
	"crypto/md5"
	"path"
	"log"
	"os"
	"io/ioutil"
	"net/http"
	"time"
	"fmt"
)

// WindowSize custom windom size
type WindowSize struct {
	Height	int
	Width	int
}

// Config configuration for the webview
type Config struct {
	Title	string	`json:",omitempty"`
	Port	int	`json:",omitempty"`
	Path	string	`json:",omitempty"`
	Debug	bool	`json:",omitempty"`
	Token	string	`json:",omitempty"`
	Size WindowSize	`json:",omitempty"`
}

// CreateServer create server with custom handlers
func CreateServer(config *Config) *http.Server {
	mux := http.NewServeMux()

	if config.Path == "" {
		config.Path = "./"
	}

	cookie := http.Cookie{
		Name:    "token",
		Value:   config.Token,
	}

	if config.Token != "" {
		if config.Token == "random" {
			h := md5.New()
			io.WriteString(h, time.Now().String())
			io.WriteString(h, "webview-server")
			config.Token = fmt.Sprintf("%x", h.Sum(nil))
			cookie.Value = config.Token
		}

		mux.Handle("/*", http.FileServer(http.Dir(config.Path)))

		mux.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
			http.SetCookie(w, &cookie)
			http.ServeFile(w, r, path.Join(config.Path, "index.html"))
		})
	} else {
		mux.Handle("/", http.FileServer(http.Dir(config.Path)))
	}

	mux.HandleFunc("/api/file", func(w http.ResponseWriter, r *http.Request) {
		isAuth := true

		if config.Token != "" {
			isAuth = false

			for _, c := range r.Cookies() {
				if (c.Name == "token" && c.Value == config.Token) {
					isAuth = true
				}
			}
		}

		if !isAuth {
			throwHTTP(&w, fmt.Errorf("unauthorized"), http.StatusUnauthorized)
			return
		}

		f := r.URL.Query()["filename"]
		if len(f) == 0 {
			throwHTTP(&w, fmt.Errorf("filename not supplied"), http.StatusNotFound)
			return
		}
		filename := f[0]

		if r.Method == "GET" {
			data, eReadFile := ioutil.ReadFile(filename)
			if eReadFile != nil {
				throwHTTP(&w, eReadFile, http.StatusInternalServerError)
				return
			}
			w.Write(data)
			return
		} else if r.Method == "PUT" {
			data, eReadAll := ioutil.ReadAll(r.Body)
			if eReadAll != nil {
				throwHTTP(&w, eReadAll, http.StatusInternalServerError)
				return
			}
			eWriteFile := ioutil.WriteFile(filename, data, 0666)
			if eWriteFile != nil {
				throwHTTP(&w, eWriteFile, http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)
			return
		} else if r.Method == "DELETE" {
			eRemove := os.Remove(filename)
			if eRemove != nil {
				throwHTTP(&w, eRemove, http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)
			return
		}

		throwHTTP(&w, fmt.Errorf("unsupported method"), http.StatusMethodNotAllowed)
	})

	server := &http.Server{
		Handler: mux,
	}

	return server
}

// OnExit execute on exit, including SIGTERM and SIGINT
func OnExit() {
	log.Println("Executing clean-up function")
	// time.Sleep(2 * time.Second)
	log.Println("Clean-up finished")
}

func throwHTTP(w *http.ResponseWriter, e error, code int) {
	http.Error(*w, e.Error(), code)
	log.Println(code, e)
}
