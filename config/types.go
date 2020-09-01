package config

import "net"

// WindowSize custom windom size
type WindowSize struct {
	Height int
	Width  int
}

// Config configuration for the webview
type Config struct {
	Www   string
	Port  int
	Debug bool
	Token string
	Size  WindowSize
	Cmd   []string

	// Internal
	Root     string
	URL      string
	Listener net.Listener
}
