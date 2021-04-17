package config

import (
	"net/url"
	"path/filepath"
	"regexp"
)

var re *regexp.Regexp = regexp.MustCompile(`[^-_a-zA-Z0-9]+`)

type Config struct {
	Host             string
	Scheme           string
	Port             int
	PemPath, KeyPath string
	StorePath        string
	FileName         string
	Verbose          bool
	ForceSave        bool
	DynamoFileName   bool
	URL              *url.URL

	// This option provides deleting records which weren't requested during tests.
	SessionMode bool

	//If NoUseDomain is true proxy don't use domain name and port for storing data.
	//So you may use it for test with different servers.
	NoUseDomain bool

	// If NoUseUserData is true proxy don't use user's name for storing data.
	// So you may use it for test with different user.
	NoUseUserData bool
}

func (cfg *Config) Init() (err error) {
	cfg.URL, err = url.Parse(cfg.Host)
	return
}

func (cfg *Config) File(urlPath string) string {

	if !cfg.DynamoFileName && cfg.FileName != "" {
		return filepath.Join(cfg.StorePath, cfg.FileName)
	}

	urlPath = re.ReplaceAllString(urlPath, "") + ".db"
	return filepath.Join(cfg.StorePath, urlPath)
}
