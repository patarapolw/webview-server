package config

import "net"

// WindowSize custom windom size
type WindowSize struct {
	Height int
	Width  int
}

// Config configuration for the webview
type Config struct {
	Title    string     `json:",omitempty"`
	Port     int        `json:",omitempty"`
	Path     string     `json:",omitempty"`
	Debug    bool       `json:",omitempty"`
	Token    string     `json:",omitempty"`
	Size     WindowSize `json:",omitempty"`
	Listener net.Listener
}
