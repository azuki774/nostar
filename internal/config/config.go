package config

import (
	"github.com/BurntSushi/toml"
)

const softwareSrcURL = "https://github.com/azuki774/nostar"

// TODO: 一部コンフィグは、not yet implemented
type Config struct {
	// Database  DatabaseConfig  `toml:"database"`
	// Server    ServerConfig    `toml:"server"`
	RelayInfo RelayInfoConfig `toml:"relay_info"`
}

type RelayInfoConfig struct {
	Name          string `toml:"name"`
	Description   string `toml:"description"`
	Pubkey        string `toml:"pubkey"`
	Contact       string `toml:"contact"`
	Software      string `toml:"software"`
	Version       string `toml:"version"`
	SupportedNIPs []int  `toml:"supported_nips"`
	// Limitations    LimitationsConfig    `toml:"limitation"`
	RelayCountries []string `toml:"relay_countries"`
	LanguageTags   []string `toml:"language_tags"`
	// Tags           TagsConfig           `toml:"tags"`
	PostingPolicy string `toml:"posting_policy"`
}

// type LimitationsConfig struct {
// 	MaxMessageLength int  `toml:"max_message_length"`
// 	MaxSubscriptions int  `toml:"max_subscriptions"`
// 	MaxFilters       int  `toml:"max_filters"`
// 	MaxLimit         int  `toml:"max_limit"`
// 	MaxSubIDLength   int  `toml:"max_subid_length"`
// 	MinPowDifficulty int  `toml:"min_pow_difficulty"`
// 	AuthRequired     bool `toml:"auth_required"`
// 	PaymentRequired  bool `toml:"payment_required"`
// }

func LoadConfig(path string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, err
	}

	config.RelayInfo.Software = softwareSrcURL
	// TODO: version を自動で設定
	return &config, nil
}
