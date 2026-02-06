if vim.g.loaded_tgrep_nvim == 1 then
  return
end

vim.g.loaded_tgrep_nvim = 1

require("tgrep").setup()
