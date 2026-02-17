package main

import (
	"go-live-markdown/internal/host"
	"log"

	"github.com/neovim/go-client/nvim/plugin"
)

func main() {
	plugin.Main(func(p *plugin.Plugin) error {
		log.Println("[go-live-markdown] registering handlers")
		return host.Register(p)
	})
}
