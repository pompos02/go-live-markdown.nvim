package main

import (
	"go-live-markdown/internal/host"
	"log"

	"github.com/neovim/go-client/nvim/plugin"
)

// Set up the connection to Neovim
// Take the plugin object we register commands
// Keep the connection alive and listen for request
func main() {
	plugin.Main(func(p *plugin.Plugin) error {
		log.Println("[go-live-markdown] registering handlers")
		return host.Register(p)
	})
}
