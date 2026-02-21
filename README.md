# go-live-markdown.nvim

Highly opinionated live Markdown preview for Neovim, powered by a small Go host with websockets.

## What it does

- Starts a browser preview with `:GoLiveMarkdownStart`
- Auto-updates on markdown edits and cursor movement (`*.md`)
- Syncs browser line clicks back to Neovim (Double click in browser)

This plugin is intentionally minimal and opinionated

## Install and configure

### lazy.nvim

```lua
{
  "karavellas/go-live-markdown.nvim",
  ft = "markdown",
  config = function()
    -- Optional: override host binary path.
    -- vim.g.go_live_markdown_host_prog = "/absolute/path/to/go-live-markdown-nvim"
  end,
}
```

Only one option is supported:


- `vim.g.go_live_markdown_host_prog`: absolute path to a custom `go-live-markdown-nvim` binary.

By default, the plugin resolves the host binary in this order:
1. `vim.g.go_live_markdown_host_prog`
2. local `./bin/go-live-markdown-nvim` inside the plugin
3. `go-live-markdown-nvim` from your `$PATH`

## Binary and building

This repository ships with a **Linux binary bundled** at `bin/go-live-markdown-nvim`.

If you want to build it yourself:
```bash
./build
```

That script compiles `./cmd/go-live-markdown-nvim` into `./bin/go-live-markdown-nvim`.

> [!NOTE]
> - Almost all of the client side logic was written by AI
> - A known bug is that only one preview tab is supported reliably at a time right now.
