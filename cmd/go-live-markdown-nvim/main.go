// Package main boots the Go host process used by the Neovim plugin bridge.
package main

import (
	"go-live-markdown/internal/host"
	"log"

	"github.com/neovim/go-client/nvim/plugin"
)

// main registers plugin handlers and starts the Neovim host loop.
func main() {
	plugin.Main(func(p *plugin.Plugin) error {
		log.Println("[go-live-markdown] registering handlers")
		return host.Register(p)
	})
}
