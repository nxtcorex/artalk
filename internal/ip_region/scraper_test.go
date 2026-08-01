package ip_region

import (
	"os"
	"testing"

	"github.com/artalkjs/artalk/v2/internal/config"
	geoip2 "github.com/oschwald/geoip2-golang/v2"
	"github.com/stretchr/testify/require"
)

func TestIpScraper(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  127.0.0.1  ", "127.0.0.1"},
		{"192.168.1.1, 10.0.0.1", "192.168.1.1"},
		{"  2001:0db8:85a3:0000:0000:8a2e:0370:7334  ", "2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
		{"", ""},
		{", 172.16.0.1", ""},
	}

	for _, test := range tests {
		result := ipScraper(test.input)
		if result != test.expected {
			t.Errorf("For input '%s', expected '%s' but got '%s'", test.input, test.expected, result)
		}
	}
}

func TestRegionScraper(t *testing.T) {
	tests := []struct {
		raw       string
		precision config.IPRegionPrecision
		expected  string
	}{
		{"0|0|0|0|0", config.IPRegionCountry, ""},
		{"中国|0|上海市|上海市|电信", config.IPRegionProvince, "上海"},
		{"中国|0|上海市|上海市|电信", config.IPRegionCity, "上海"},
		{"中国|0|重庆市|重庆市|电信", config.IPRegionCity, "重庆"},
		{"中国|0|浙江省|杭州市|电信", config.IPRegionProvince, "浙江"},
		{"中国|0|北京市|北京市|电信", config.IPRegionCity, "北京"},
		{"中国|0|四川省|泸州市|电信", config.IPRegionCity, "四川泸州"},
		{"中国|0|广东省|深圳市|电信", config.IPRegionCity, "广东深圳"},
		{"中国|0|湖南省|岳阳市|电信", config.IPRegionCity, "湖南岳阳"},
		{"中国|0|陕西省|西安市|电信", config.IPRegionCity, "陕西西安"},
		{"中国|0|河北省|石家庄市|电信", config.IPRegionCity, "河北石家庄"},
		{"中国|0|江苏省|南京市|电信", config.IPRegionCity, "江苏南京"},
		{"美国|0|0|0|微软云", config.IPRegionCity, "美国"},
	}

	for _, test := range tests {
		result := regionScraper(test.raw, test.precision)
		if result != test.expected {
			t.Errorf("For input '%s' and precision '%s', expected '%s' but got '%s'", test.raw, test.precision, test.expected, result)
		}
	}
}

func TestMaxMindRegionScraper(t *testing.T) {
	tests := []struct {
		raw       string
		precision config.IPRegionPrecision
		expected  string
	}{
		{"United States|California|Mountain View", config.IPRegionCountry, "United States"},
		{"United States|California|Mountain View", config.IPRegionProvince, "California"},
		{"United States|California|Mountain View", config.IPRegionCity, "California Mountain View"},
		{"United States||", config.IPRegionCity, "United States"},
		{"United States|California|", config.IPRegionCity, "California"},
		{"中国|浙江省|杭州市", config.IPRegionProvince, "浙江"},
		{"中国|浙江省|杭州市", config.IPRegionCity, "浙江杭州"},
		{"", config.IPRegionCountry, ""},
		{"United States|California|Mountain View", "unknown", ""},
	}

	for _, test := range tests {
		result := maxmindRegionScraper(test.raw, test.precision)
		if result != test.expected {
			t.Errorf("For input '%s' and precision '%s', expected '%s' but got '%s'", test.raw, test.precision, test.expected, result)
		}
	}
}

func TestNameByLocale(t *testing.T) {
	names := geoip2.Names{
		English:           "China",
		SimplifiedChinese: "中国",
	}

	require.Equal(t, "中国", nameByLocale(names, "zh-CN"))
	require.Equal(t, "China", nameByLocale(names, "en"))
	require.Equal(t, "China", nameByLocale(names, "fr"))
	require.Equal(t, "China", nameByLocale(geoip2.Names{English: "China"}, "zh-CN"))
}

func TestMaxMindProviderInvalidIP(t *testing.T) {
	provider := &maxMindProvider{dbPath: "not-opened.mmdb"}
	_, err := provider.Search("invalid")
	require.ErrorContains(t, err, "invalid ip address")
}

func TestMaxMindProviderWithCityDatabase(t *testing.T) {
	dbPath := os.Getenv("ARTALK_TEST_MAXMIND_CITY_DB")
	if dbPath == "" {
		t.Skip("set ARTALK_TEST_MAXMIND_CITY_DB to run MaxMind integration test")
	}
	defer func() {
		require.NoError(t, CloseMaxMindReaders())
	}()

	provider := &maxMindProvider{
		dbPath: dbPath,
		locale: "en",
	}
	raw, err := provider.Search("8.8.8.8")
	require.NoError(t, err)
	require.NotEmpty(t, raw)
	require.NotEmpty(t, maxmindRegionScraper(raw, config.IPRegionCountry))
}
