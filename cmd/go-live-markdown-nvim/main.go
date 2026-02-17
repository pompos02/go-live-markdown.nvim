package main

import (
	"go-live-markdown/internal/nvimhost"
	"log"

	"github.com/neovim/go-client/nvim/plugin"
)

func main() {
	plugin.Main(func(p *plugin.Plugin) error {
		log.Println("[go-live-markdown] registering handlers")
		return nvimhost.Register(p)
	})
}
