use serde::Serialize;
use worker::*;

#[derive(Serialize)]
struct CfProperties {
    colo: String,
    asn: Option<u32>,
    as_organization: Option<String>,
    country: Option<String>,
    city: Option<String>,
    continent: Option<String>,
    coordinates: Option<Coordinates>,
    postal_code: Option<String>,
    metro_code: Option<String>,
    region: Option<String>,
    region_code: Option<String>,
    timezone: String,
}

#[derive(Serialize)]
struct Coordinates {
    latitude: f32,
    longitude: f32,
}

impl From<&Cf> for CfProperties {
    fn from(cf: &Cf) -> Self {
        Self {
            colo: cf.colo(),
            asn: cf.asn(),
            as_organization: cf.as_organization(),
            country: cf.country(),
            city: cf.city(),
            continent: cf.continent(),
            coordinates: cf.coordinates().map(|(lat, lon)| Coordinates {
                latitude: lat,
                longitude: lon,
            }),
            postal_code: cf.postal_code(),
            metro_code: cf.metro_code(),
            region: cf.region(),
            region_code: cf.region_code(),
            timezone: cf.timezone_name(),
        }
    }
}

#[event(fetch)]
async fn fetch(req: Request, _env: Env, _ctx: Context) -> Result<Response> {
    let url = req.url()?;
    let path = url.path();

    match path {
        "/v1.json" => {
            let cf = req
                .cf()
                .ok_or_else(|| Error::RustError("No CF properties".into()))?;
            let props = CfProperties::from(cf);

            let pretty = url.query_pairs().any(|(k, _)| k == "pretty");
            let json = if pretty {
                serde_json::to_string_pretty(&props)
            } else {
                serde_json::to_string(&props)
            }
            .map_err(|e| Error::RustError(e.to_string()))?;

            Response::ok(json).map(|r| r.with_headers(json_headers()))
        }
        _ => Response::ok("geoip.pls - try /v1.json"),
    }
}

fn json_headers() -> Headers {
    let headers = Headers::new();
    let _ = headers.set("Content-Type", "application/json");
    headers
}
