if vim.g.loaded_go_live_markdown then
    return
end
vim.g.loaded_go_live_markdown = 1

local function host_prog()
    if vim.g.go_live_markdown_host_prog ~= nil then
        return vim.fn.expand(vim.g.go_live_markdown_host_prog, true)
    end

    local source = debug.getinfo(1, "S").source
    local script_path = source:sub(1, 1) == "@" and source:sub(2) or source
    local repo_root = vim.fn.fnamemodify(script_path, ":h:h")
    local local_bin = repo_root .. "/bin/go-live-markdown-nvim"
    if vim.fn.executable(local_bin) == 1 then
        return local_bin
    end

    return vim.fn.exepath("go-live-markdown-nvim")
end

function _G.go_live_markdown_require_host(host)
  local prog = host_prog()
  if prog == "" then
    vim.notify(
      "go-live-markdown host binary not found; run ./build or set g:go_live_markdown_host_prog",
      vim.log.levels.ERROR
    )
    return 0
  end

    return vim.fn["provider#Poll"]({ prog }, host.orig_name, "$NVIM_GO_LIVE_MARKDOWN_LOG_FILE")
end

vim.cmd([[function! GoLiveMarkdownRequireHost(host) abort
  return v:lua.go_live_markdown_require_host(a:host)
endfunction]])
vim.cmd([[call remote#host#Register('go_live_markdown', '*', function('GoLiveMarkdownRequireHost'))]])

vim.cmd([[call remote#host#RegisterPlugin('go_live_markdown', '0', [
\ {'type': 'command', 'name': 'GoLiveMarkdownStart', 'sync': 1, 'opts': {}},
\ {'type': 'command', 'name': 'GoLiveMarkdownUpdate', 'sync': 1, 'opts': {}},
\ ])]])

local group = vim.api.nvim_create_augroup("go_live_markdown_updates", { clear = true })

--------------------------------------------------------------------------------
--------------------------- AUTOCOMMANDS ---------------------------------------
--------------------------------------------------------------------------------

vim.api.nvim_create_autocmd({ "TextChanged", "TextChangedI" }, {
  group = group,
  pattern = "*.md",
  callback = function()
    pcall(vim.api.nvim_cmd, {
      cmd = "GoLiveMarkdownUpdate",
      mods = { silent = true, emsg_silent = true },
    }, {})
  end,
})
