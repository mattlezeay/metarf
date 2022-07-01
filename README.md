# metarf
METAR Fetcher for the CLI. This is intended to be a tool that can run in the background (like i3 bar) for constant updates to latest METARs.

## Data sources
This uses the data sources from aviationweather.gov and their dataserver to fetch METAR data. See [their docs](https://aviationweather.gov/dataserver) for more info.

## Configuration
The configuration file is metar-fetcher.toml and this is how it knows which stations to fetch.

```toml
# Configuration file for METAR fetcher

title = "METAR Fetcher"

[metars]
stations = ["KPAE", "KBFI", "KFLY", "KRNT"]

[display]
category = true
winds = true
ceilings = false
```

