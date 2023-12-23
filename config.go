package combo

import (
	"errors"
	"strings"
)

// Config 包含了 Combo SDK 运行所必需的配置项。
type Config struct {
	// API Endpoint
	Endpoint Endpoint

	// 游戏的 Game ID
	GameId GameId

	// 游戏的 Secret Key
	SecretKey SecretKey
}

func (cfg *Config) validate() error {
	if cfg.Endpoint == "" {
		return errors.New("missing required Endpoint")
	}
	cfg.Endpoint = Endpoint(strings.TrimSuffix(string(cfg.Endpoint), "/"))
	if cfg.GameId == "" {
		return errors.New("missing required GameId")
	}
	if cfg.SecretKey == nil || len(cfg.SecretKey) == 0 {
		return errors.New("missing required SecretKey")
	}
	if !strings.HasPrefix(string(cfg.SecretKey), "sk_") {
		return errors.New("invalid SecretKey: must start with sk_")
	}
	return nil
}
