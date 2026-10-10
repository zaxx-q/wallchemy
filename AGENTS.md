# WALLCHEMY

**Generated:** 2026-10-10T10:18:00Z
**Commit:** 8fea264

CLI wallpaper manager and daemon for Wallhaven.cc written in Go. Supports mode rotation, browser-style history traversal, favorites, and compositor IPC script execution.

## STRUCTURE

```
.
├── main.go            # CLI entry point, logger bootstrap, context wiring
├── Justfile           # Recipes for testing, versioned ldflags building, and user installation
├── app/               # Domain engines: config, cache, history stack, rotation state, downloader, notifications
└── handlers/          # Command execution: rotate, previous, single, random, favorites, daemon
```

## WHERE TO LOOK

| Task                                    | Location                                                                      |
| --------------------------------------- | ----------------------------------------------------------------------------- |
| CLI flags & subcommand parsing          | `app/config.go` (`setupFlags`, `LoadConfig`)                                  |
| History navigation & index tracking     | `app/history.go`, `handlers/exec.go` (`restoreHistoryEntry`, `handleRotate`)  |
| Rotation modes (`random`, `top`, `hot`) | `app/rotate.go`, `handlers/exec.go` (`handleRotate`)                          |
| Cache pruning & query invalidation      | `app/cache.go`, `app/cachemanager.go`, `app/prune.go`                         |
| Wallpaper setter invocation             | `app/apply.go` (`ApplyWallpaper`)                                             |
| Desktop notifications & run logs        | `app/notifier.go` (`NotifyStart`, `NotifySuccess`, `NotifyError`)             |
| Browser extension daemon & IPC          | `handlers/daemon.go`, `handlers/daemon_unix.go`, `handlers/daemon_windows.go` |
| Current wallpaper cache info            | `app/current.go`, `handlers/current.go`                                       |
| Favorites collection & launcher         | `app/favorites.go`, `app/menu.go`                                             |

## CONVENTIONS

- Subcommands (`rotate`, `next`, `previous`, `hot`, `top`, `favorite`) parsed from positional `pflag.Args()` in `app/config.go` and normalized to config flags.
- History traversal is non-destructive: `previous` decrements `Current`; `rotate`/`next` increments `Current` until at the edge. Only new fetches branch and truncate forward history.
- Always use `appCtx.Config.StateDir` and `appCtx.Config.FavoritesDir`; default to `$XDG_STATE_HOME/wallchemy` and `$XDG_DATA_HOME/wallchemy/favorites`.
- Compositor scripts configured via `script` in `config.yml` receive downloaded file path as `$1` via `sh -c`.
- API results cached in `<sort>_history` alongside `<sort>_query` URL comparisons to eliminate unnecessary Wallhaven API queries within `expiry` seconds.

## ANTI-PATTERNS

- Never mutate `history.json` directly; always go through `appCtx.HistoryManager` methods to ensure mutex synchronization and atomic temp file renames.
- Do not trigger a fresh network fetch in `handleRotate` when `data.Current < len(data.Entries)-1`.
- Never hardcode user paths (`~/.cache`, `~/.local/state`); always expand via `expandPath()` and respect XDG environment variables.
- Do not suppress errors silently in `ApplyWallpaper`; failures from external IPC (`noctalia`, `swww`, etc.) must bubble up to `Notifier.NotifyError`.

## COMMANDS

```bash
just test              # Run unit tests: go test ./...
just build             # Build local binary with git version ldflags
just install           # Test and install to ~/.local/bin/wallchemy
just install-system    # Test and install to /usr/local/bin/wallchemy (requires sudo)
```

## KEY CONFIGS

| Tool             | Entry                                   | Notes                                                       |
| ---------------- | --------------------------------------- | ----------------------------------------------------------- |
| Wallchemy config | `~/.config/wallchemy/config.yml`        | API keys, categories, purity, ratios, custom `script`       |
| Rotation state   | `~/.local/state/wallchemy/mode-index`   | Tracks round-robin index across modes                       |
| History state    | `~/.local/state/wallchemy/history.json` | Stores `current` index pointer and `entries` list           |
| Run log          | `~/.local/state/wallchemy/last-run.log` | Appended by `Notifier.Log` on starts, successes, and errors |
| Cached files     | `~/.cache/wallchemy/wallpapers/`        | Pruned automatically to `cache_keep` (default 150)          |

## UNIQUE STYLES

- `noctalia msg wallpaper-set <file>` is the standard integration pattern on Sway/Niri with Noctalia backdrop shell.
- When binding `wallchemy` in Wayland compositors (Niri, Hyprland, Sway), set `repeat=false` to avoid high-frequency key-repeat races during multi-second network downloads.
