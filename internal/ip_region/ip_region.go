package ip_region

import (
	"strings"

	"github.com/artalkjs/artalk/v2/internal/config"
	"github.com/artalkjs/artalk/v2/internal/log"
)

type IPRegion struct {
	conf IPRegionConf
}

type IPRegionConf struct {
	config.IPRegionConf
	CacheEnabled bool
	Locale       string
}

func NewIPRegion(conf IPRegionConf) *IPRegion {
	return &IPRegion{
		conf: conf,
	}
}

func (ipRegion *IPRegion) IP2Region(ip string) string {
	if strings.TrimSpace(ip) == "" {
		return ""
	}

	ip = ipScraper(ip)
	provider, err := newProvider(ipRegion.conf)
	if err != nil {
		log.Warn("[IPRegion] ", err)
		return ""
	}

	region, err := provider.Search(ip)
	if err != nil {
		if !strings.HasPrefix(err.Error(), "invalid ip address") {
			log.Warn("[IPRegion] ", err)
		}
		return ""
	}

	precision := config.IPRegionPrecision(ipRegion.conf.Precision)
	if config.IPRegionProvider(ipRegion.conf.Provider) == config.IPRegionProviderMaxMind {
		return maxmindRegionScraper(region, precision)
	}
	return regionScraper(region, precision)
}
