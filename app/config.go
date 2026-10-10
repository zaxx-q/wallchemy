package app

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	RandomSearchMarker        = "__RANDOM_SEARCH_TERM__"
	RandomCollectionMarker    = "__RANDOM_COLLECTION__"
	BlacklistCurrentAndRotate = "__BLACKLIST_CURRENT_AND_ROTATE__"
)

var DefaultSearchTerms = []string{
	"landscape",
	"nature mountains",
	"cyberpunk city",
	"anime scenery",
	"dark abstract",
	"space nebula",
	"forest",
	"ocean sunset",
	"minimalist",
	"architecture",
	"anime",
	"video game",
	"videogame",
}

var configFile string

type Config struct {
	Version       bool     `mapstructure:"version"`
	Help          bool     `mapstructure:"help"`
	ConfigFile    string   `mapstructure:"config"`
	APIKey        string   `mapstructure:"apikey"`
	Username      string   `mapstructure:"username"`
	ID            string   `mapstructure:"id"`
	Collection    string   `mapstructure:"collection"`
	Collections   bool     `mapstructure:"collections"`
	Random        string   `mapstructure:"random"`
	Hot           bool     `mapstructure:"hot"`
	TopList       bool     `mapstructure:"top"`
	TopRange      string   `mapstructure:"range"`
	Categories    []string `mapstructure:"categories"`
	Purity        []string `mapstructure:"purity"`
	MinResolution string   `mapstructure:"minresolution"`
	Resolutions   []string `mapstructure:"resolutions"`
	Ratios        []string `mapstructure:"ratios"`
	Colors        string   `mapstructure:"colors"`
	Seed          string   `mapstructure:"seed"`
	Expiry        int      `mapstructure:"expiry"`
	MaxPages      int      `mapstructure:"pages"`
	Output        string   `mapstructure:"output"`
	Script        string   `mapstructure:"script"`
	Quiet         bool     `mapstructure:"quiet"`
	Stealth       bool     `mapstructure:"stealth"`
	Clean         bool     `mapstructure:"clean"`
	Info          bool     `mapstructure:"info"`
	Daemon        string   `mapstructure:"daemon"`
	LaunchDaemon  bool     `mapstructure:"launchdaemon"`
	Port          int      `mapstructure:"port"`
	Blacklist     string   `mapstructure:"blacklist"`
	ListBlacklist bool     `mapstructure:"listblacklist"`

	// Rotate & script features
	Silent            bool     `mapstructure:"silent"`
	Notifications     *bool    `mapstructure:"notifications"`
	SearchTerms       []string `mapstructure:"search_terms"`
	RotateCollections []string `mapstructure:"rotate_collections"`
	RotateModes       []string `mapstructure:"rotate_modes"`
	CacheKeep         int      `mapstructure:"cache_keep"`
	HistoryKeep       int      `mapstructure:"history_keep"`
	StateDir          string   `mapstructure:"state_dir"`
	FavoritesDir      string   `mapstructure:"favorites_dir"`
	DmenuCmd          string   `mapstructure:"dmenu_cmd"`
	FileManager       string   `mapstructure:"file_manager"`

	// Action flags
	Rotate           bool `mapstructure:"rotate"`
	Next             bool `mapstructure:"next"`
	Favorite         bool `mapstructure:"favorite"`
	Favorites        bool `mapstructure:"favorites"`
	FavoritesGallery bool `mapstructure:"favorites_gallery"`
	Previous         bool `mapstructure:"previous"`
	PruneCache       bool `mapstructure:"prune_cache"`
	OpenCurrent      bool `mapstructure:"open"`
}

func (c *Config) GetRandomSearchTerm() string {
	terms := c.SearchTerms
	if len(terms) == 0 {
		terms = DefaultSearchTerms
	}
	return terms[rand.Intn(len(terms))]
}

func (c *Config) GetRandomCollection() (string, error) {
	colls := c.RotateCollections
	if len(colls) == 0 {
		return "", fmt.Errorf("no collections configured in wallchemy")
	}
	return colls[rand.Intn(len(colls))], nil
}

func GetDefaultStateDir() string {
	if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
		return filepath.Join(xdg, "wallchemy")
	}
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".local", "state", "wallchemy")
	}
	if appData := os.Getenv("LOCALAPPDATA"); appData != "" {
		return filepath.Join(appData, "wallchemy", "state")
	}
	return "./state"
}

func GetDefaultFavoritesDir() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "wallchemy", "favorites")
	}
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".local", "share", "wallchemy", "favorites")
	}
	if appData := os.Getenv("APPDATA"); appData != "" {
		return filepath.Join(appData, "wallchemy", "favorites")
	}
	return "./favorites"
}

func hasExplicitAction(c *Config) bool {
	return c.ID != "" ||
		c.Collection != "" ||
		c.Collections ||
		c.Random != "" ||
		c.Hot ||
		c.TopList ||
		c.Info ||
		c.Clean ||
		c.Daemon != "" ||
		c.LaunchDaemon ||
		c.Blacklist != "" ||
		c.ListBlacklist ||
		c.Rotate ||
		c.Next ||
		c.Previous ||
		c.Favorite ||
		c.Favorites ||
		c.PruneCache ||
		c.OpenCurrent
}

func setupFlags() {
	pflag.BoolP("version", "v", false, "Show version information and exit")
	pflag.BoolP("help", "h", false, "Show this help message and exit")

	pflag.StringVarP(&configFile, "config", "c", "", "Path to custom configuration file (YAML)")
	pflag.StringP("apikey", "k", "", "Wallhaven.cc API key for accessing private collections and NSFW content")
	pflag.StringP("username", "u", "", "Wallhaven.cc username (required for accessing your collections)")

	pflag.StringP("id", "i", "", "Download and set a specific wallpaper by its Wallhaven ID (e.g., 'ox1234')")
	pflag.StringP("collection", "C", "", "Set a random wallpaper from one of your named collections")
	pflag.Bool("collections", false, "List all available collections for your account")
	pflag.StringP("random", "r", "", "Search term for random wallpapers (e.g., 'nature', 'cars', 'abstract')")
	pflag.BoolP("hot", "H", false, "Select from currently hot/trending wallpapers")
	pflag.BoolP("top", "t", false, "Select from top-rated wallpapers (use --range to specify time period)")

	pflag.String("range", "", "Time range for top wallpapers: 1d, 3d, 1w, 1M, 3M, 6M, 1y")
	pflag.StringSlice("categories", []string{}, "Wallpaper categories to include: general, anime, people")
	pflag.StringSlice("purity", []string{}, "Content purity levels: sfw, sketchy, nsfw (requires API key for nsfw)")

	pflag.String("minresolution", "", "Minimum resolution filter (e.g., '1920x1080', '2560x1440')")
	pflag.StringSlice("resolutions", []string{}, "Exact resolution matches (e.g., '1920x1080,2560x1440,3840x2160')")
	pflag.StringSlice("ratios", []string{}, "Aspect ratio filters (e.g., '16x9,16x10,21x9,4x3,5x4')")

	pflag.String("colors", "", "Dominant color filter as hex code without # (e.g., 'ff0000' for red, '00ff00' for green)")
	pflag.String("seed", "", "Random seed for consistent results across searches (any string)")

	pflag.IntP("expiry", "e", 600, "Cache expiry time in seconds (how long to keep search results cached)")
	pflag.Int("pages", 5, "Maximum number of pages to fetch from API (24 wallpapers per page)")

	pflag.StringP("script", "s", "", "Custom script to set wallpaper instead of built-in method (receives file path as argument)")
	pflag.StringP("output", "o", "", "Output folder for wallpaper downloads")
	pflag.BoolP("quiet", "q", false, "Suppress all console output except errors")
	pflag.Bool("stealth", false, "Download wallpaper but don't actually set it (useful for testing)")

	pflag.BoolP("clean", "x", false, "Clean the cache directory (removes all cached wallpapers and search results)")
	pflag.BoolP("info", "I", false, "Show information about the currently set wallpaper")

	pflag.StringP("daemon", "d", "", "Start the background daemon for browser extension support")
	pflag.Bool("launchdaemon", false, "")
	pflag.IntP("port", "p", 2388, "Port for the background daemon to listen on.")

	pflag.StringP("blacklist", "b", "", "Add wallpaper ID to blacklist (use 'current' to blacklist currently set wallpaper)")
	pflag.Bool("listblacklist", false, "List all blacklisted wallpaper IDs")

	// Rotate & script integration flags
	pflag.BoolP("silent", "S", false, "Suppress desktop notifications")
	pflag.BoolP("rotate", "R", false, "Rotate through wallpaper modes (random, top, hot, collection)")
	pflag.BoolP("next", "N", false, "Next wallpaper in history (or rotate mode if at the end)")
	pflag.Bool("previous", false, "Restore previous wallpaper from history")
	pflag.Bool("favorite", false, "Save current wallpaper to favorites")
	pflag.Bool("favorites", false, "Browse favorite wallpapers")
	pflag.Bool("prune-cache", false, "Prune cached wallpapers to keep newest (default: 150)")
	pflag.BoolP("open", "O", false, "Open the current wallpaper on Wallhaven in browser")

	// Allow flags to have default markers when invoked without arguments:
	pflag.Lookup("random").NoOptDefVal = RandomSearchMarker
	pflag.Lookup("collection").NoOptDefVal = RandomCollectionMarker
	pflag.Lookup("blacklist").NoOptDefVal = "current"

	pflag.CommandLine.MarkHidden("launchdaemon")

	pflag.Parse()
}

func LoadConfig() (*Config, error) {
	setupFlags()

	v := viper.New()

	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
		v.AddConfigPath("$HOME/.config/wallchemy")
		v.AddConfigPath(filepath.Join(os.Getenv("AppData"), "wallchemy"))
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			Logger.Debug("Config file not found. using flags/env")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	} else {
		Logger.WithField("file", v.ConfigFileUsed()).Info("Config file found")
	}

	v.SetEnvPrefix("WALLCHEMY")
	v.AutomaticEnv()

	v.BindPFlags(pflag.CommandLine)

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	// Parse positional arguments / subcommands
	args := pflag.Args()
	if len(args) > 0 {
		subcmd := strings.ToLower(args[0])
		var subArg string
		if len(args) > 1 {
			subArg = args[1]
		}

		switch subcmd {
		case "next", "rotate":
			config.Rotate = true
		case "random":
			if subArg != "" {
				config.Random = subArg
			} else {
				config.Random = RandomSearchMarker
			}
		case "top":
			config.TopList = true
			if subArg != "" {
				config.TopRange = subArg
			}
		case "hot":
			config.Hot = true
		case "collection":
			if subArg != "" {
				config.Collection = subArg
			} else {
				config.Collection = RandomCollectionMarker
			}
		case "favorite", "favourite":
			config.Favorite = true
		case "favorites", "favourites":
			config.Favorites = true
			if subArg == "gallery" {
				config.FavoritesGallery = true
			}
		case "prune-cache", "prunecache":
			config.PruneCache = true
		case "previous", "prev":
			config.Previous = true
		case "open":
			config.OpenCurrent = true
		case "blacklist":
			if subArg != "" {
				config.Blacklist = subArg
			} else {
				config.Blacklist = BlacklistCurrentAndRotate
			}
		case "clean":
			config.Clean = true
		case "info":
			config.Info = true
		case "help":
			config.Help = true
		case "version":
			config.Version = true
		default:
			return nil, fmt.Errorf("unknown command: %s", subcmd)
		}
	}

	if config.Next {
		config.Rotate = true
	}

	// Apply defaults
	if config.CacheKeep <= 0 {
		config.CacheKeep = 150
	}
	if config.HistoryKeep <= 0 {
		config.HistoryKeep = 100
	}
	if len(config.SearchTerms) == 0 {
		config.SearchTerms = DefaultSearchTerms
	}
	if config.StateDir == "" {
		config.StateDir = GetDefaultStateDir()
	} else {
		config.StateDir = expandPath(config.StateDir)
	}
	if config.FavoritesDir == "" {
		config.FavoritesDir = GetDefaultFavoritesDir()
	} else {
		config.FavoritesDir = expandPath(config.FavoritesDir)
	}

	if config.Notifications == nil {
		defaultNotifications := true
		config.Notifications = &defaultNotifications
	}
	if config.Silent {
		*config.Notifications = false
	}

	// Check if invoked as wallchemy-rotate or no action was provided
	binName := filepath.Base(os.Args[0])
	if strings.Contains(binName, "rotate") && !hasExplicitAction(&config) {
		config.Rotate = true
	} else if !hasExplicitAction(&config) && !config.Help && !config.Version {
		// Default action is rotate
		config.Rotate = true
	}

	return &config, nil
}

func ShowHelp() {
	fmt.Println(`Wallchemy - Wallhaven.cc Wallpaper Manager

USAGE:
  wallchemy [COMMAND] [OPTIONS]

COMMANDS:
  next, rotate                      Next wallpaper in history (rotates through modes if at end) (default)
  random [term]                     Random search (term optional, uses configured terms if omitted)
  top [range]                       Top wallpapers (default range: 1M)
  hot                               Hot / trending wallpapers
  collection [name]                 Random wallpaper from collection (name optional)
  favorite                          Save the current wallpaper locally to favorites
  favorites                         Browse, apply, open, or remove favorites via launcher
  favorites gallery                 Open favorites folder in file manager
  previous, prev                    Restore previous wallpaper from history
  prune-cache                       Keep only newest cached wallpapers (default: 150)
  open                              Open current wallpaper on Wallhaven in browser
  blacklist [id|current]            Blacklist wallpaper (current and rotate next by default)
  clean                             Clean the cache directory
  info                              Show current wallpaper information

WALLPAPER SELECTION OPTIONS:
  -R, --rotate                      Rotate through wallpaper modes
  -N, --next                        Next wallpaper in history (rotates if at end)
  -i, --id <wallpaper-id>           Download specific wallpaper by ID (e.g., 'ox1234')
  -r, --random [search-term]        Search for random wallpapers
  -H, --hot                         Get hot/trending wallpapers
  -t, --top                         Get top-rated wallpapers
  -C, --collection [name]           Random wallpaper from collection
      --collections                 List your available collections
      --previous                    Restore previous wallpaper
      --favorite                    Save current wallpaper to favorites
      --favorites                   Browse favorite wallpapers
  -O, --open                        Open current wallpaper in browser
      --prune-cache                 Prune cache

FILTERING OPTIONS:
      --range <period>              Time range for --top: 1d, 3d, 1w, 1M, 3M, 6M, 1y
      --categories <list>           Categories: general, anime, people
      --purity <list>               Content: sfw, sketchy, nsfw (nsfw needs API key)
      --minresolution <res>         Minimum resolution (e.g., 1920x1080)
      --resolutions <list>          Exact resolutions (e.g., 1920x1080,2560x1440)
      --ratios <list>               Aspect ratios (e.g., 16x9,16x10,21x9)
      --colors <hex>                Color filter (hex without #, e.g., ff0000)
      --seed <string>               Random seed for consistent results

CONFIGURATION:
  -c, --config <file>               Custom config file path
  -k, --apikey <key>                Wallhaven.cc API key
  -u, --username <name>             Wallhaven.cc username

EXECUTION & NOTIFICATIONS:
  -S, --silent                      Suppress desktop notifications (useful for timer)
  -s, --script <path>               Custom wallpaper-setting script
  -o, --output <path>               Output folder for wallpaper downloads
  -q, --quiet                       Suppress console output
      --stealth                     Download but don't set wallpaper
  -e, --expiry <seconds>            Cache expiry (default: 600)
      --pages <num>                 Max pages to fetch (default: 5)

UTILITIES:
  -I, --info                        Show current wallpaper info
  -x, --clean                       Clean cache directory
  -b, --blacklist [id|current]      Add wallpaper ID to blacklist (default: current)
      --listblacklist               List all blacklisted wallpapers
  -v, --version                     Show version
  -h, --help                        Show this help

BROWSER EXTENSION DAEMON:
  -d, --daemon start                Start background daemon
  -d, --daemon status               Check background daemon status
  -d, --daemon stop                 Stop background daemon
  -p, --port <num>                  Daemon port (default: 2388)

EXAMPLES:
  wallchemy                         # Rotate to next wallpaper
  wallchemy rotate                  # Rotate to next wallpaper
  wallchemy random anime            # Search anime wallpapers
  wallchemy top 1w                  # Top wallpapers from past week
  wallchemy hot                     # Trending wallpapers
  wallchemy favorite                # Save current wallpaper to favorites
  wallchemy favorites               # Browse favorites menu
  wallchemy favorites gallery       # Open favorites folder
  wallchemy previous                # Restore previous wallpaper
  wallchemy open                    # Open wallpaper page on Wallhaven
  wallchemy blacklist               # Blacklist current and rotate next
  wallchemy rotate --silent         # Rotate silently (for systemd timer)`)
}
