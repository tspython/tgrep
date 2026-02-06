# tgrep

Terminal UI grep tool built with Bubble Tea + Lip Gloss.

## Install

### Install binary from source

```bash
go install github.com/tspython/tgrep@latest
```

Make sure your Go bin directory is in `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Run directly

```bash
tgrep
```

## Neovim Integration

This repo includes a lightweight Neovim plugin wrapper.

`lazy.nvim` example:

```lua
{
  "tspython/tgrep",
  config = function()
    require("tgrep").setup({
      bin = "tgrep",
      border = "rounded",
      width = 0.92,
      height = 0.88,
    })

    vim.keymap.set("n", "<leader>fg", "<cmd>Tgrep<cr>", { desc = "Open tgrep" })
  end,
}
```

Commands:

- `:Tgrep` opens in a floating terminal using the active Neovim cwd.
- `:tgrep` also works (command-line abbreviation to `:Tgrep`).
- `:Tgrep /path/to/project` opens with an explicit cwd.

## Release Process

Releases are automated with GoReleaser via GitHub Actions.

1. Commit and push to `main`.
2. Create and push a version tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

3. GitHub Action `.github/workflows/release.yml` publishes binaries to GitHub Releases.

Release config is in `.goreleaser.yaml`.
