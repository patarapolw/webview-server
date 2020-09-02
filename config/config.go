package config

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

// Get get config from the root config.json
func Get() *Config {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	debug := false
	if os.Getenv("DEBUG") != "" {
		debug = true
	}

	config := Config{
		Port:  port,
		Www:   os.Getenv("WWW"),
		Debug: debug,
	}

	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	if data, err := ioutil.ReadFile(path.Join(dir, "config.yaml")); err == nil {
		// Discard errors
		yaml.Unmarshal(data, &config)
	}

	config.Root = dir
	config.Www = path.Join(dir, config.Www)

	// Normalize filepaths on all platforms, and make it absolute (to executable)
	for i, d := range config.Cmd {
		if strings.HasPrefix(d, "./") || strings.HasPrefix(d, ".\\") {
			config.Cmd[i] = path.Join(config.Root, filepath.FromSlash(d[2:]))
		}
	}

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

	os.Setenv("PORT", strconv.Itoa(config.Port))
	os.Setenv("WWW", config.Www)
	os.Setenv("TOKEN", config.Token)
	os.Setenv("ROOT", config.Root)

	listener, err := net.Listen("tcp", "localhost:"+strconv.Itoa(config.Port))
	if err != nil {
		log.Fatal(err)
	}

	config.Listener = listener

	return &config
}
