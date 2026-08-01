package ip_region

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"sync"

	"github.com/artalkjs/artalk/v2/internal/config"
	geoip2 "github.com/oschwald/geoip2-golang/v2"
)

type maxMindProvider struct {
	dbPath string
	locale string
}

var maxMindReaderCache = struct {
	sync.Mutex
	readers map[string]*geoip2.Reader
}{
	readers: map[string]*geoip2.Reader{},
}

func (p *maxMindProvider) Search(ip string) (string, error) {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return "", fmt.Errorf("invalid ip address: %w", err)
	}

	reader, err := getMaxMindReader(p.dbPath)
	if err != nil {
		return "", err
	}

	city, cityErr := reader.City(addr)
	if cityErr == nil {
		if city == nil || !city.HasData() {
			return "", nil
		}

		country := nameByLocale(city.Country.Names, p.locale)
		if country == "" {
			country = nameByLocale(city.RegisteredCountry.Names, p.locale)
		}

		subdivision := ""
		if len(city.Subdivisions) > 0 {
			subdivision = nameByLocale(city.Subdivisions[0].Names, p.locale)
		}

		cityName := nameByLocale(city.City.Names, p.locale)
		return strings.Join([]string{country, subdivision, cityName}, "|"), nil
	}

	country, countryErr := reader.Country(addr)
	if countryErr != nil {
		return "", fmt.Errorf("maxmind lookup failed: %w", errors.Join(cityErr, countryErr))
	}
	if country == nil || !country.HasData() {
		return "", nil
	}

	countryName := nameByLocale(country.Country.Names, p.locale)
	if countryName == "" {
		countryName = nameByLocale(country.RegisteredCountry.Names, p.locale)
	}
	return strings.Join([]string{countryName, "", ""}, "|"), nil
}

func getMaxMindReader(dbPath string) (*geoip2.Reader, error) {
	maxMindReaderCache.Lock()
	defer maxMindReaderCache.Unlock()

	if reader, ok := maxMindReaderCache.readers[dbPath]; ok {
		return reader, nil
	}

	reader, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open maxmind database %q: %w", dbPath, err)
	}
	maxMindReaderCache.readers[dbPath] = reader
	return reader, nil
}

func CloseMaxMindReaders() error {
	maxMindReaderCache.Lock()
	defer maxMindReaderCache.Unlock()

	var closeErr error
	for dbPath, reader := range maxMindReaderCache.readers {
		if err := reader.Close(); err != nil {
			closeErr = errors.Join(closeErr, fmt.Errorf("failed to close maxmind database %q: %w", dbPath, err))
		}
	}
	maxMindReaderCache.readers = map[string]*geoip2.Reader{}
	return closeErr
}

func nameByLocale(names geoip2.Names, locale string) string {
	if strings.HasPrefix(strings.ToLower(locale), "zh") && names.SimplifiedChinese != "" {
		return names.SimplifiedChinese
	}
	return names.English
}

func maxmindRegionScraper(raw string, precision config.IPRegionPrecision) string {
	if precision == "" {
		precision = config.IPRegionProvince
	}

	sep := strings.Split(raw, "|")
	if len(sep) < 3 {
		return ""
	}

	country := normalizeRegionPart(sep[0])
	province := normalizeRegionPart(sep[1])
	city := normalizeRegionPart(sep[2])

	if country == "" {
		return ""
	}

	switch precision {
	case config.IPRegionCountry:
		return country
	case config.IPRegionProvince:
		if province == "" {
			return country
		}
		return province
	case config.IPRegionCity:
		if province == "" && city == "" {
			return country
		}
		return joinRegionParts(province, city)
	default:
		return ""
	}
}

func normalizeRegionPart(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(strings.TrimSuffix(s, "省"), "市")
	return s
}

func joinRegionParts(province string, city string) string {
	if province == "" {
		return city
	}
	if city == "" || city == province {
		return province
	}
	if containsCJK(province) || containsCJK(city) {
		return province + city
	}
	return province + " " + city
}

func containsCJK(s string) bool {
	for _, r := range s {
		if (r >= '\u4e00' && r <= '\u9fff') || (r >= '\u3400' && r <= '\u4dbf') {
			return true
		}
	}
	return false
}
