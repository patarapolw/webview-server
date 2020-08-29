package main

import (
	"log"
	"os"
	"io/ioutil"
	"net/http"
	"time"
	"fmt"
)

// Config configuration for the webview
type Config struct {
	Title string
	Port string
	Path string
	Debug string
}

// CreateServer create server with custom handlers
func CreateServer() *http.Server {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("./dist")))
	mux.HandleFunc("/api/file", func(w http.ResponseWriter, r *http.Request) {
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

		throwHTTP(&w, fmt.Errorf("unsupported method"), http.StatusNotFound)
	})

	server := &http.Server{
		Handler: mux,
	}

	return server
}

// OnExit execute on exit, including SIGTERM and SIGINT
func OnExit() {
	log.Println("Executing clean-up function")
	time.Sleep(2 * time.Second)
	log.Println("Clean-up finished")
}

func throwHTTP(w *http.ResponseWriter, e error, code int) {
	http.Error(*w, e.Error(), code)
	log.Println(code, e)
}
