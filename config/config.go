package config

import (
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/muhammadmuzzammil1998/jsonc"
)

// Get get config from the root config.json
func Get() *Config {
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
	config.Listener = listener

	return &config
}
