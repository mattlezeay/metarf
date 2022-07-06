# metarf
METAR Fetcher for the CLI. This is intended to be a tool that can run in the background (like i3 bar) for constant updates to latest METARs.

## Data sources
This uses the data sources from aviationweather.gov and their dataserver to fetch METAR data. See [their docs](https://aviationweather.gov/dataserver) for more info.

## Configuration
The configuration file is metar-fetcher.toml and this is how it knows which stations to fetch. The default search path
is in ~/.config/ and . but can be specified with --config 

```toml
# Configuration file for METAR fetcher

title = "METAR Fetcher"

[metars]
stations = ["KPAE", "KBFI", "KFLY", "KRNT"] # List of weather stations
cached = true                               # Whether to cache results

[display]
category = false                            # Print actual fligh category
winds = true                                # Print winds
ceilings = true                             # Print the ceiling
longlived = false                           # Continue running and refreshing
i3json = false                              # Print in i3 JSON format for i3status
```


