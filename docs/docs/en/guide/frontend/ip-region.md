# IP Region

Artalk has a built-in feature to display the geolocation of user IPs, with adjustable precision levels: city or province.

This feature is disabled by default. You can enable the IP region display feature in the Artalk Dashboard settings.

## IP Region Database

Artalk uses `ip2region` as the default IP region database provider.

Before enabling the IP region display feature with the default `ip2region` provider, you need to download a database file:

- [GitHub Download](https://github.com/lionsoul2014/ip2region/raw/master/data/ip2region.xdb) / [Mirror Download](https://gh-proxy.com/?q=https%3A%2F%2Fgithub.com%2Flionsoul2014%2Fip2region%2Fblob%2Fmaster%2Fdata%2Fip2region.xdb) (recommended within China)

After downloading, manually place it in the `./data/` directory, and name the file: `ip2region.xdb`.

You can also use the `maxmind` provider with MaxMind GeoLite2 databases. Artalk supports `GeoLite2-City.mmdb` and `GeoLite2-Country.mmdb`. These database files must be downloaded and maintained by the user; Artalk does not automatically download them and does not distribute them in release packages.

When using `GeoLite2-Country.mmdb`, province and city precision will fall back to the country/region value because the Country database does not include province or city data.

## Precision Settings

You can find this configuration item in the settings.

| Precision Level | Description       | Example       |
| --------------- | ----------------- | ------------- |
| `province`      | Province (default) | `Sichuan`     |
| `city`          | City              | `Sichuan Chengdu` |
| `country`       | Country/Region    | `China`       |

Configuration file:

```yaml
# IP Region
ip_region:
  # Enable IP region display
  enabled: false
  # Database provider ["ip2region", "maxmind"]
  provider: ip2region
  # Data file path (.xdb format)
  db_path: ./data/ip2region.xdb
  # Display precision ["province", "city", "country"]
  precision: province
```

MaxMind example:

```yaml
ip_region:
  enabled: true
  provider: maxmind
  db_path: ./data/GeoLite2-City.mmdb
  precision: city
```

## Obtaining the Correct IP Address

If you are using a CDN or a trusted reverse proxy server like Nginx, you need to specify the request header field containing the user's real IP in the "Settings" - "Server" option - "Proxy Header Name (`http.proxy_header`)", such as `X-Real-IP` (for security, this field is empty by default). After modification, please manually restart the Artalk service to take effect.

Otherwise, Artalk will not be able to obtain the user's real IP address (if using Docker, the IP obtained may always be 172.17.0.X, which is the IP of the Docker virtual network card).

## Privacy Policy

Artalk comments will record users' `IP` and `User-Agent` data. Since such data pertains to user privacy, please declare this in your website's privacy policy and inform users that privacy data will be collected when they comment.
