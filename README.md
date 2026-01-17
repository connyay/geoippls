# geoippls

A lightweight GeoIP service built with Rust and Cloudflare Workers. Returns geographic and network information about the requesting client based on Cloudflare's edge data.

## API

### `GET /v1.json`

Returns GeoIP information as JSON.

Add `?pretty` for formatted output.

**Response:**

```json
{
  "colo": "DFW",
  "asn": 12345,
  "as_organization": "Example ISP",
  "country": "US",
  "city": "Dallas",
  "continent": "NA",
  "coordinates": {
    "latitude": 32.7767,
    "longitude": -96.797
  },
  "postal_code": "75201",
  "metro_code": "623",
  "region": "Texas",
  "region_code": "TX",
  "timezone": "America/Chicago"
}
```

## Development

### Prerequisites

- [Rust](https://rustup.rs/)
- [Wrangler CLI](https://developers.cloudflare.com/workers/wrangler/install-and-update/)

### Run locally

```sh
wrangler dev
```

### Deploy

```sh
wrangler deploy
```
