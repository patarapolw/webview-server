package server

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"

	conf "github.com/patarapolw/webview-server/config"
)

// CreateServer create server with custom handlers
func CreateServer(config *conf.Config) *http.Server {
	mux := http.NewServeMux()

	if config.Token == "" {
		n, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 256))
		if err == nil {
			config.Token = fmt.Sprintf("%x", n)
		} else {
			config.Token = "disabled"
		}
	}

	if config.Token == "disabled" {
		config.Token = ""
	}

	cookie := http.Cookie{
		Name:  "token",
		Value: config.Token,
	}

	if config.Token != "" {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.SetCookie(w, &cookie)
			http.FileServer(http.Dir(config.Path)).ServeHTTP(w, r)
		})
	} else {
		mux.Handle("/", http.FileServer(http.Dir(config.Path)))
	}

	mux.HandleFunc("/api/file", func(w http.ResponseWriter, r *http.Request) {
		isAuth := true

		if config.Token != "" {
			isAuth = false

			for _, c := range r.Cookies() {
				if c.Name == "token" && c.Value == config.Token {
					isAuth = true
					break
				}
			}
		}

		if !isAuth {
			for _, c := range r.Header["Authorization"] {
				p := strings.Split(c, " ")
				if len(p) == 2 && p[0] == "Bearer" && p[1] == config.Token {
					isAuth = true
					break
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

func throwHTTP(w *http.ResponseWriter, e error, code int) {
	http.Error(*w, e.Error(), code)
	log.Println(code, e)
}
