local M = {}

local defaults = {
  bin = "tgrep",
  border = "rounded",
  width = 0.92,
  height = 0.88,
  cwd = function()
    return vim.fn.getcwd(0, 0)
  end,
}

local opts = vim.deepcopy(defaults)
local command_created = false

local function resolve_cwd(cwd_override)
  if cwd_override and cwd_override ~= "" then
    return cwd_override
  end

  if type(opts.cwd) == "function" then
    return opts.cwd()
  end

  if type(opts.cwd) == "string" and opts.cwd ~= "" then
    return opts.cwd
  end

  return vim.fn.getcwd(0, 0)
end

local function notify(msg, level)
  vim.notify(msg, level or vim.log.levels.INFO, { title = "tgrep.nvim" })
end

function M.open(cwd_override)
  if vim.fn.executable(opts.bin) ~= 1 then
    notify(("Binary '%s' not found in PATH"):format(opts.bin), vim.log.levels.ERROR)
    return
  end

  local cwd = resolve_cwd(cwd_override)
  if vim.fn.isdirectory(cwd) ~= 1 then
    notify(("Invalid cwd: %s"):format(cwd), vim.log.levels.ERROR)
    return
  end

  local width = math.max(60, math.floor(vim.o.columns * opts.width))
  local height = math.max(16, math.floor(vim.o.lines * opts.height))
  local row = math.floor((vim.o.lines - height) / 2)
  local col = math.floor((vim.o.columns - width) / 2)

  local buf = vim.api.nvim_create_buf(false, true)
  vim.bo[buf].bufhidden = "wipe"

  local win = vim.api.nvim_open_win(buf, true, {
    relative = "editor",
    style = "minimal",
    border = opts.border,
    width = width,
    height = height,
    row = row,
    col = col,
  })

  vim.fn.termopen({ opts.bin }, {
    cwd = cwd,
    on_exit = function()
      vim.schedule(function()
        if vim.api.nvim_win_is_valid(win) then
          vim.api.nvim_win_close(win, true)
        end
      end)
    end,
  })

  vim.cmd("startinsert")
end

function M.setup(user_opts)
  opts = vim.tbl_deep_extend("force", opts, user_opts or {})

  if command_created then
    return
  end

  vim.api.nvim_create_user_command("Tgrep", function(cmd)
    M.open(cmd.args)
  end, {
    nargs = "?",
    complete = "dir",
    desc = "Open tgrep in a floating terminal",
  })

  vim.cmd([[cnoreabbrev <expr> tgrep getcmdtype() == ':' && getcmdline() ==# 'tgrep' ? 'Tgrep' : 'tgrep']])

  command_created = true
end

return M
