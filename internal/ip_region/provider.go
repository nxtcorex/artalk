package ip_region

import (
	"fmt"

	"github.com/artalkjs/artalk/v2/internal/config"
)

type Provider interface {
	Search(ip string) (rawRegion string, err error)
}

type ip2regionProvider struct {
	dbPath       string
	cacheEnabled bool
}

func newProvider(conf IPRegionConf) (Provider, error) {
	provider := config.IPRegionProvider(conf.Provider)
	if provider == "" {
		provider = config.IPRegionProviderIP2Region
	}

	switch provider {
	case config.IPRegionProviderIP2Region:
		return &ip2regionProvider{
			dbPath:       conf.DBPath,
			cacheEnabled: conf.CacheEnabled,
		}, nil
	case config.IPRegionProviderMaxMind:
		return &maxMindProvider{
			dbPath: conf.DBPath,
			locale: conf.Locale,
		}, nil
	default:
		return nil, fmt.Errorf("unknown IP region provider: %q", conf.Provider)
	}
}

func (p *ip2regionProvider) Search(ip string) (string, error) {
	return search(ip, p.dbPath, p.cacheEnabled)
}
