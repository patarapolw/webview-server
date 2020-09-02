package config

import "net"

// WindowSize custom windom size
type WindowSize struct {
	Height int
	Width  int
}

// Config configuration for the webview
type Config struct {
	Www    string
	Port   int
	Debug  bool
	Token  string
	Size   WindowSize
	Cmd    []string
	Sqlite string // Connection string of sqlite connection, see https://github.com/mattn/go-sqlite3#connection-string

	// Internal
	Root     string
	Listener net.Listener
}

// URL make the URL
func (c *Config) URL() string {
	return "http://" + c.Listener.Addr().String()
}
