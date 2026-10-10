<p align="center">
<img src="/images/logo.png" alt="wallchemy logo" width="140px"/>
<h2>Wallchemy</h2>
</p>

Wallchemy is a CLI tool written in Go for fetching random wallpapers from [wallhaven.cc](https://wallhaven.cc/). Wallchemy will randomly choose a wallpaper from various lists including your own collections using your API key. Wallchemy will set the wallpaper on your desktop on Windows, Mac and many Linux desktop environments.

[![Release](https://img.shields.io/github/release/zaxx-q/wallchemy.svg)](https://github.com/zaxx-q/wallchemy/releases/latest)
[![Platforms](https://img.shields.io/badge/platforms-linux%20|%20macos%20|%20windows-blue)](<>)

### Supported desktop environments

- Windows
- MacOS
- Linux
  - Gnome
  - Ubuntu
  - KDE Plasma
  - XFCE
  - Cinnamon
  - MATE
  - Budgie

Any unsupported can be manually set by passing a custom script. Feel free to raise an issue or a pull request for any missing environments.

All queries and selected wallpapers are cached to cut down on API calls.

---

### Installation

See the releases page for suitable executable for your system. Alternatively you can run one of the install scripts.

#### Linux/Mac

```sh
curl -sSL https://raw.githubusercontent.com/zaxx-q/wallchemy/main/install.sh | bash
```

#### Windows

```
irm https://raw.githubusercontent.com/zaxx-q/wallchemy/main/install.ps1 | iex
```

---

### Configuration

Configuration is stored in `$HOME/.config/wallchemy/config.yml` on linux/mac or `%appdata%/Roaming/wallchemy/config.yml` on Windows. Wallchemy's defaults will be overwritten by these options, which in turn will be overwritten by any environment variables set proceeded with `WALLCHEMY_`. These can then be overridden with any command line flags.

#### Example `config.yml`

```yaml
apikey: "your-api-key-here"
username: "yourusername"
categories: ["general", "anime"]
purity: ["sfw", "sketchy"]
expiry: 1200
pages: 10
ratios: ["16x9"]
minresolution: "1920x1080"
script: "noctalia msg wallpaper-set"

# Rotation & script integration options
search_terms:
  - "landscape"
  - "nature mountains"
  - "cyberpunk city"
  - "anime scenery"
  - "space nebula"
rotate_collections: [] # Collections to include in rotation
rotate_modes: ["random", "top", "hot"]
top_range: "1M" # Default range for top wallpapers
cache_keep: 150 # Number of cached wallpapers to retain
history_keep: 100 # Number of history entries to remember
notifications: true # Send desktop notifications
# dmenu_cmd: "rofi -dmenu -p" # Custom launcher command (auto-detects noctalia, rofi, wofi, dmenu)
# file_manager: "thunar"     # Custom file manager for favorites gallery
```

## Quick Start

### Basic Usage

```bash
# Rotate through modes (default action when called without arguments)
wallchemy
wallchemy rotate

# Random search (term optional; picks from configured search terms if omitted)
wallchemy random
wallchemy random "cyberpunk city"

# Hot/trending wallpaper
wallchemy hot

# Top-rated wallpaper (range optional, default 1M)
wallchemy top 1w

# Wallpaper from named collection (name optional; picks from rotate_collections if omitted)
wallchemy collection "my favourites"

# Save current wallpaper to favorites
wallchemy favorite

# Browse and manage favorites via interactive launcher (noctalia, rofi, wofi, dmenu)
wallchemy favorites

# Open favorites gallery in file manager
wallchemy favorites gallery

# Restore previous wallpaper from history
wallchemy previous

# Open current wallpaper on Wallhaven in browser
wallchemy open

# Blacklist current wallpaper and immediately rotate to next
wallchemy blacklist

# Prune cache keeping newest wallpapers
wallchemy prune-cache

# Run silently without desktop notifications (useful for systemd timers/cron)
wallchemy rotate --silent
```

### With API Key (Required for Collections & NSFW)

```bash
# Set your API key and username
wallchemy --apikey "your-api-key" --username "yourusername" --collection "my favourites"
```

## Command Reference

### Commands & Rotation Modes

| Command             | Description                                                 | Example                       |
| ------------------- | ----------------------------------------------------------- | ----------------------------- |
| `next`, `rotate`    | Rotate through modes (`random`, `top`, `hot`, `collection`) | `wallchemy rotate`            |
| `random [term]`     | Random search (term optional)                               | `wallchemy random landscape`  |
| `top [range]`       | Top wallpapers (default range: 1M)                          | `wallchemy top 1M`            |
| `hot`               | Hot / trending wallpapers                                   | `wallchemy hot`               |
| `collection [name]` | Wallpaper from named collection                             | `wallchemy collection`        |
| `favorite`          | Save current wallpaper locally to favorites                 | `wallchemy favorite`          |
| `favorites`         | Browse favorites (set, open, remove) via launcher           | `wallchemy favorites`         |
| `favorites gallery` | Open favorites directory in file manager                    | `wallchemy favorites gallery` |
| `previous`, `prev`  | Restore previous wallpaper from history                     | `wallchemy previous`          |
| `prune-cache`       | Prune cache keeping newest wallpapers                       | `wallchemy prune-cache`       |
| `open`              | Open current wallpaper's page on Wallhaven                  | `wallchemy open`              |
| `blacklist`         | Blacklist current wallpaper and fetch next                  | `wallchemy blacklist`         |

### Wallpaper Selection Flags

| Flag                      | Description                         | Example                        |
| ------------------------- | ----------------------------------- | ------------------------------ |
| `-R, --rotate`            | Rotate through wallpaper modes      | `wallchemy -R`                 |
| `-i, --id <id>`           | Download specific wallpaper         | `wallchemy -i ox83dp3`         |
| `-r, --random [term]`     | Search for random wallpapers        | `wallchemy -r "cyberpunk"`     |
| `-H, --hot`               | Get hot/trending wallpapers         | `wallchemy -H`                 |
| `-t, --top`               | Get top-rated wallpapers            | `wallchemy -t --range 1w`      |
| `-C, --collection [name]` | Random from your collection         | `wallchemy -C "my favourites"` |
| `--collections`           | List available collections          | `wallchemy --collections`      |
| `--previous`              | Restore previous wallpaper          | `wallchemy --previous`         |
| `--favorite`              | Save current wallpaper to favorites | `wallchemy --favorite`         |
| `--favorites`             | Browse favorite wallpapers          | `wallchemy --favorites`        |
| `--prune-cache`           | Prune cache directory               | `wallchemy --prune-cache`      |
| `-O, --open`              | Open current wallpaper in browser   | `wallchemy -O`                 |

### Filtering Options

| Flag              | Values                                   | Description               |
| ----------------- | ---------------------------------------- | ------------------------- |
| `--categories`    | `general`, `anime`, `people`             | Wallpaper categories      |
| `--purity`        | `sfw`, `sketchy`, `nsfw`                 | Content rating            |
| `--range`         | `1d`, `3d`, `1w`, `1M`, `3M`, `6M`, `1y` | Time range for `--top`    |
| `--minresolution` | `1920x1080`, `2560x1440`, etc.           | Minimum resolution        |
| `--resolutions`   | `1920x1080,2560x1440`                    | Exact resolutions         |
| `--ratios`        | `16x9,21x9,4x3`                          | Aspect ratios             |
| `--colors`        | `ff0000`, `00ff00`                       | Hex color (without #)     |
| `--seed`          | Any string                               | Consistent random results |

### Execution Options

| Flag                     | Description                         |
| ------------------------ | ----------------------------------- |
| `-s, --script <path>`    | Custom wallpaper-setting script     |
| `--output <dir>`         | Custom output directory             |
| `-q, --quiet`            | Suppress output except errors       |
| `--stealth`              | Download but don't set wallpaper    |
| `-e, --expiry <seconds>` | Cache expiry time (default: 600)    |
| `-p, --pages <num>`      | Max API pages to fetch (default: 5) |

### Blacklist Management

| Flag                            | Description                     |
| ------------------------------- | ------------------------------- |
| `-b, --blacklist <id\|current>` | Add wallpaper to blacklist      |
| `--listblacklist`               | List all blacklisted wallpapers |

### Utilities

| Flag                                 | Description                 |
| ------------------------------------ | --------------------------- |
| `-I, --info`                         | Show current wallpaper info |
| `-x,--clean`                         | Clean cache directory       |
| `-d, --daemon <start\|stop\|status>` | Manage background daemon    |
| `-p, --port <num>`                   | Daemon port (default: 2388) |

---

### Browser extensions and daemon mode

Wallchemy can set the desktop wallpaper directly from the browser using an extension and the Wallchemy background daemon. Start the daemon by running `wallchemy --daemon start`. The extension can be installed from the [Chrome web store](https://chromewebstore.google.com/detail/wallchemy-helper/iegfdnonehdmlnmldgkekhmoiplkokbd) or the [Firefox Add-Ons](https://addons.mozilla.org/en-GB/firefox/addon/wallchemy-helper/)

For more info check the [wiki](https://github.com/zaxx-q/wallchemy/wiki/Browser-extensions)

---

### Setting wallpaper with external script

The wallchemy app sets the wallpaper depending on OS and desktop environment used. If yours is currently unsupported or you need a custom script you can set it using the `script` option in `config.yml` or via the `--script` flag (`wallchemy --hot --script "swww img"`)

You can also just pipe the output directly into a script.

`wallchemy --top | swww img` for example. Or if it does not allow piping `feh --bg-fill $(wallchemy --top)`
