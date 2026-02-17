if exists('g:loaded_go_live_markdown_host')
  finish
endif
let g:loaded_go_live_markdown_host = 1

call remote#host#Register('go_live_markdown', '*', function('provider#go_live_markdown#Require'))
