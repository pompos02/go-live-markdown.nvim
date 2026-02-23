if vim.g.loaded_go_live_markdown then
    return
end
vim.g.loaded_go_live_markdown = 1

-- Resolve the Go host executable. Priority:
-- 1) g:go_live_markdown_host_prog
-- 2) local repo build in ./bin
-- 3) executable found in PATH
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

-- Called by Neovim's remote-host registry to validate/start the plugin host.
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

-- Register the Lua bridge function used by remote#host#Register.
vim.cmd([[function! GoLiveMarkdownRequireHost(host) abort
  return v:lua.go_live_markdown_require_host(a:host)
endfunction]])

-- Register the remote host and the command/function entry points implemented in Go.
vim.cmd([[call remote#host#Register('go_live_markdown', '*', function('GoLiveMarkdownRequireHost'))]])

vim.cmd([[call remote#host#RegisterPlugin('go_live_markdown', '0', [
\ {'type': 'command', 'name': 'GoLiveMarkdownStart', 'sync': 1, 'opts': {}},
\ {'type': 'function', 'name': 'GoLiveMarkdownInternalUpdate', 'sync': 1, 'opts': {}},
\ {'type': 'function', 'name': 'GoLiveMarkdownInternalCursor', 'sync': 1, 'opts': {}},
\ ])]])

local group = vim.api.nvim_create_augroup("go_live_markdown_updates", { clear = true })

--------------------------------------------------------------------------------
--------------------------- AUTOCOMMANDS ---------------------------------------
--------------------------------------------------------------------------------

vim.api.nvim_create_autocmd({ "TextChanged", "TextChangedI" }, {
    group = group,
    pattern = "*.md",
    callback = function()
        -- Ignore RPC errors to avoid disrupting normal editing flow.
        pcall(vim.api.nvim_call_function, "GoLiveMarkdownInternalUpdate", {})
    end,
})

vim.api.nvim_create_autocmd({ "CursorMoved", "CursorMovedI" }, {
    group = group,
    pattern = "*.md",
    callback = function()
        -- Ignore RPC errors to avoid disrupting normal editing flow.
        pcall(vim.api.nvim_call_function, "GoLiveMarkdownInternalCursor", {})
    end,
})

vim.api.nvim_create_autocmd({ "BufEnter" }, {
    group = group,
    pattern = "*.md",
    callback = function()
        -- Keep preview in sync when markdown buffers gain/lose focus.
        pcall(vim.api.nvim_call_function, "GoLiveMarkdownInternalUpdate", {})
    end,
})
