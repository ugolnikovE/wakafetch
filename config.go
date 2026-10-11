package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	flag "github.com/spf13/pflag"
	"gopkg.in/ini.v1"
)

const defaultAPIURL = "https://api.wakatime.com/api/v1"

// Range is the time period to fetch statistics for.
// It is passed as the first positional argument.
type Range string

// validRanges lists every value accepted by Range.Set.
var validRanges = []string{"today", "yesterday", "week", "month", "half", "year", "all"}

// TopValue holds the number of entries to show in a top list.
// When IsAll is true, every entry is shown and Value is ignored.
type TopValue struct {
	IsAll bool
	Value int
}

// Config holds the runtime settings of wakafetch.
type Config struct {
	// WakaTime API settings
	APIURL string
	APIKey string
	// IsWakaTime is true if APIURL points to wakatime.com. Otherwise
	// a WakaTime-compatible server, such as Wakapi, is assumed.
	IsWakaTime bool

	// Output settings
	Range   Range
	IsShort bool
	LangTop TopValue
}

// String implements pflag.Value.
func (v *TopValue) String() string {
	if v.IsAll {
		return "all"
	}
	return strconv.Itoa(v.Value)
}

// Set implements pflag.Value. It accepts an integer or "all".
func (v *TopValue) Set(s string) error {
	if s == "all" {
		v.IsAll = true
		v.Value = 0
		return nil
	}

	num, err := strconv.Atoi(s)
	if err != nil || num <= 0 {
		return errors.New("must be a positive integer or 'all'")
	}

	v.IsAll = false
	v.Value = num
	return nil
}

// Type implements pflag.Value.
func (v *TopValue) Type() string {
	return "int|all"
}

// String implements pflag.Value.
func (r *Range) String() string {
	return string(*r)
}

// Set implements pflag.Value. It accepts only values from validRanges.
func (r *Range) Set(s string) error {
	for _, v := range validRanges {
		if s == v {
			*r = Range(s)
			return nil
		}
	}
	return errors.New("invalid range, must be one of: " + strings.Join(validRanges, ", "))
}

// Type implements pflag.Value.
func (r *Range) Type() string {
	return "string"
}

// ParseFlags parses command-line flags and the optional range argument.
// Flags override the current API settings, so it must be called after
// ParseConfigFile and ParseEnv. The range defaults to "today".
// It returns an error if the range argument is invalid.
func (c *Config) ParseFlags() error {
	flag.StringVarP(&c.APIKey, "api-key", "k", c.APIKey, "WakaTime API key (for debugging only)")
	flag.StringVarP(&c.APIURL, "api-url", "u", c.APIURL, "base URL of the WakaTime-compatible API")
	flag.BoolVarP(&c.IsShort, "one-line", "1", false, "print statistics on a single line")
	flag.VarP(&c.LangTop, "lang-top", "t", "number of top languages to show, or 'all'")

	flag.Usage = func() {
		w := os.Stdout
		fmt.Fprint(w, "Usage: wakafetch [flags] [range]\n")
		fmt.Fprint(w, "\n")
		fmt.Fprint(w, "Display your coding statistics from WakaTime or any WakaTime-compatible server.\n")
		fmt.Fprint(w, "\n")
		fmt.Fprint(w, "Range:\n")
		fmt.Fprint(w, "  today (default), yesterday, week, month, half, year, all\n")
		fmt.Fprint(w, "\n")
		fmt.Fprint(w, "Flags:\n")
		fmt.Fprint(w, "  -k, --api-key string    WakaTime API key. Intended for debugging only:\n")
		fmt.Fprint(w, "                          the key stays in your shell history. Use the\n")
		fmt.Fprint(w, "                          config file or WAKATIME_API_KEY instead\n")
		fmt.Fprint(w, "  -u, --api-url string    base URL of the WakaTime-compatible API\n")
		fmt.Fprint(w, "  -1, --one-line          print statistics on a single line\n")
		fmt.Fprint(w, "  -t, --lang-top int|all  number of top languages to show, or 'all'\n")
		fmt.Fprint(w, "  -h, --help              show this help\n")
		fmt.Fprint(w, "\n")
		fmt.Fprint(w, "Configuration sources, each overriding the previous one:\n")
		fmt.Fprint(w, "  1. $WAKATIME_HOME/.wakatime.cfg or ~/.wakatime.cfg\n")
		fmt.Fprint(w, "     (api_key and api_url in the [settings] section)\n")
		fmt.Fprint(w, "  2. WAKATIME_API_KEY and WAKATIME_API_URL environment variables\n")
		fmt.Fprint(w, "  3. command-line flags\n")
	}

	flag.Parse()

	if flag.NArg() > 1 {
		return fmt.Errorf("too many arguments: expected at most 1 range, got %d: %s",
			flag.NArg(), strings.Join(flag.Args(), " "))
	}

	if flag.NArg() == 0 {
		c.Range.Set("today")
		return nil
	}

	return c.Range.Set(flag.Arg(0))
}

// ParseEnv overrides the API settings with non-empty
// WAKATIME_API_KEY and WAKATIME_API_URL environment variables.
func (c *Config) ParseEnv() {
	envAPIKey := os.Getenv("WAKATIME_API_KEY")
	envAPIURL := os.Getenv("WAKATIME_API_URL")

	if envAPIKey != "" {
		c.APIKey = envAPIKey
	}
	if envAPIURL != "" {
		c.APIURL = envAPIURL
	}
}

// ParseConfigFile reads api_key and api_url from the [settings] section
// of .wakatime.cfg, located in $WAKATIME_HOME or, if it is unset, in the
// user's home directory. A missing file is not an error. It returns an
// error if the home directory cannot be determined or the file is malformed.
func (c *Config) ParseConfigFile() error {
	options := ini.LoadOptions{
		AllowPythonMultilineValues: true,
		Loose:                      true,
	}

	var cfgFilePath string
	envCfgPath, exists := os.LookupEnv("WAKATIME_HOME")

	if !exists {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot determine home directory: %w", err)
		}
		cfgFilePath = filepath.Join(homeDir, ".wakatime.cfg")
	} else {
		cfgFilePath = filepath.Join(envCfgPath, ".wakatime.cfg")
	}

	cfg, err := ini.LoadSources(options, cfgFilePath)
	if err != nil {
		return fmt.Errorf("cannot read config file: %w", err)
	}

	c.APIKey = cfg.Section("settings").Key("api_key").String()
	c.APIURL = cfg.Section("settings").Key("api_url").String()

	return nil
}

// NewConfig builds a Config from the config file, environment variables
// and command-line flags, in that order of increasing priority.
// It returns an error if the API key is not set or the API URL is malformed.
func NewConfig() (*Config, error) {
	var c Config

	if err := c.ParseConfigFile(); err != nil {
		return nil, err
	}

	c.ParseEnv()

	if err := c.ParseFlags(); err != nil {
		return nil, err
	}

	if c.APIKey == "" {
		return nil, errors.New("api key not set: add api_key to .wakatime.cfg or set WAKATIME_API_KEY")
	}

	if c.APIURL == "" {
		c.APIURL = defaultAPIURL
	}

	parsedURL, err := url.Parse(c.APIURL)
	if err != nil {
		return nil, err
	}

	if parsedURL.Host == "api.wakatime.com" || parsedURL.Host == "wakatime.com" {
		c.IsWakaTime = true
	} else {
		c.IsWakaTime = false
	}

	return &c, nil
}
