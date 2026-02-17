function! provider#go_live_markdown#Prog() abort
  if exists('g:go_live_markdown_host_prog')
    return expand(g:go_live_markdown_host_prog, v:true)
  endif

  let l:repo_root = fnamemodify(expand('<sfile>:p'), ':h:h:h')
  let l:local_bin = l:repo_root . '/bin/go-live-markdown-nvim'
  if executable(l:local_bin)
    return l:local_bin
  endif

  return exepath('go-live-markdown-nvim')
endfunction

function! provider#go_live_markdown#Require(host) abort
  let prog = provider#go_live_markdown#Prog()
  if empty(prog)
    echoerr 'go-live-markdown host binary not found; run ./build or set g:go_live_markdown_host_prog'
    return 0
  endif
  return provider#Poll([prog], a:host.orig_name, '$NVIM_GO_LIVE_MARKDOWN_LOG_FILE')
endfunction
