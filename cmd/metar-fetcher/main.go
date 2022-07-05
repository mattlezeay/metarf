package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/TwiN/go-color"
	"github.com/rs/zerolog/log"
	"github.com/urkk/addstogo"
	"io/ioutil"
	"metarFetcher/internal"
	"net/http"
	"os"
	"strings"
	"time"
)

const LOG_FILENAME = "/tmp/metarf.log"

type DisplayConfig struct {
	Category  bool `toml:"category"`
	Winds     bool
	Ceilings  bool
	LongLived bool
}

type Metars struct {
	Stations []string
}

// MetarConfig is the primary container for configuration
type MetarConfig struct {
	Metars  Metars        `toml:"metars"`
	Display DisplayConfig `toml:"display"`
}

// FormatStations produces a URL formatted list of stations for the API.
// TODO: make this part of a proper RESTful client
func (c *MetarConfig) FormatStations() (s string, err error) {
	if len(c.Metars.Stations) < 1 {
		return "", fmt.Errorf("metarfetcher: no stations defined")
	}
	return strings.Join(c.Metars.Stations, "%20"), nil
}

// ReadConfig takes the provided configuration file and fills out the struct for later use
// TODO: This should search for the file in known paths rather than be provided it.
func (c *MetarConfig) ReadConfig(fileLocation string) (err error) {
	rawConfig, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		return fmt.Errorf("metarfetcher: error reading config file located at %s: %s\n", fileLocation, err)
	}

	_, err = toml.Decode(string(rawConfig), c)
	if err != nil {
		return fmt.Errorf("metarfetcher: error decoding config: %s\n", err)
	}
	return nil
}

// FormatMetarUrl takes the config and produces a url with the configured stations to fetch the latest METAR
// TODO: Turn this into something more configurable and a proper client. Right now it just does a naive latest METAR
func FormatMetarUrl(config *MetarConfig) (url string, err error) {
	url = "https://aviationweather.gov/adds/dataserver_current/httpparam?dataSource=metars&requestType=retrieve&format=xml&mostRecentForEachStation=constraint&hoursBeforeNow=1.25&stationString="

	stationStr, err := config.FormatStations()
	if err != nil {
		return "", fmt.Errorf("Unable to format stations: %s", err)
	}
	return url + stationStr, nil
}

// drawOutput is the beginning of the formatter. Right now its handling color but will need to do more as config options
// come up
func drawOutput(config *MetarConfig, metars *addstogo.METARresponse) (err error) {
	for _, m := range metars.Data.METAR {
		output := fmt.Sprintf("%s", m.StationID)
		if config.Display.Winds {
			output += fmt.Sprintf("•%03d%d", m.WindDirDegrees, m.WindSpeedKt)
			if m.WindGustKt > 0 {
				output += fmt.Sprintf("G%d", m.WindGustKt)
			}
			output += "KT"
		}
		if config.Display.Category {
			output += fmt.Sprintf("•%s", m.FlightCategory)
		}
		if config.Display.Ceilings {
			if len(m.SkyCondition) > 1 {
				for _, sky := range m.SkyCondition {
					if sky.SkyCover == "BKN" || sky.SkyCover == "OVC" {
						output += fmt.Sprintf("•%d", sky.CloudBaseFtAgl)
						break
					}
				}
			} else {
				output += "•CLR"
			}

		}
		output += fmt.Sprintf("•%s", m.ObservationTime.Format("020304Z"))
		switch m.FlightCategory {
		case "VFR":
			fmt.Print(color.InGreen(output))
		case "MVFR":
			fmt.Print(color.InBlue(output))
		case "IFR":
			fmt.Print(color.InRed(output))
		case "LIFR":
			fmt.Print(color.InPurple(output))
		default:
			fmt.Print(output)
		}

		fmt.Print("\n", color.Reset)
	}
	return nil
}

// fetchMetarData is the primary function for pulling METAR data and returning it deserialized into a struct
// TODO: fix the passing of logger when this is refactored. It works, but feels strange
func fetchMetarData(config *MetarConfig, logger *internal.Logger) (metars *addstogo.METARresponse, err error) {
	// Get a usable Url for fetching the METARs
	url, err := FormatMetarUrl(config)
	if err != nil {
		logger.Fatal().Err(err).Msg("")
	}
	// Call the endpoint
	res, getErr := http.Get(url)
	if getErr != nil {
		logger.Fatal().Err(getErr).Msg("")
	}
	logger.Info().Msgf("Fetched data from %s", url)

	// Read the body which should be xml
	xmlResponse, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		logger.Fatal().Err(readErr).Msg("")
	}
	// Create the response to parse the xml
	metars, err = addstogo.UnmarshalMetars(xmlResponse)
	if err != nil {
		logger.Fatal().Err(err).Msg("")
	}
	return metars, nil
}

func shouldPull() bool {
	// Check if cache file is there, if it is, check if the data is stale
	if _, err := os.Stat("/tmp/metarf_cache.tmp"); err == nil {
		return false
	}
	return true
}

// main is the cli entry point. This will get refactored into a proper CLI but for now its doing all the heavy lifting
func main() {

	// Reads in the config primarily for the stations right now
	config := MetarConfig{}
	err := config.ReadConfig("metar-fetcher.toml")
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to read config file")
	}

	// Set up the logger
	logger := internal.NewLogger(false)
	defer logger.Close()

	var metars *addstogo.METARresponse

	// Debug code to ensure the METARs are coming through
	//fmt.Printf("First metar is: %s\n", metars.Data.METAR)
	if config.Display.LongLived {
		// Clear screen
		fmt.Print("\033[H\033[2J")
		for {
			// Fetch the data
			metars, err = fetchMetarData(&config, logger)
			if err != nil {
				logger.Error().Err(err).Msg("Unable to fetch METAR data")
			}
			// Get back to top of terminal for re-printing each time
			fmt.Printf("\033[0;0H")
			err = drawOutput(&config, metars)
			if err != nil {
				logger.Fatal().Err(err).Msg("")
			}
			time.Sleep(5 * time.Minute)

		}
	} else {
		// Fetch the data then print once
		metars, err = fetchMetarData(&config, logger)
		if err != nil {
			logger.Error().Err(err).Msg("Unable to fetch METAR data")
		}
		err = drawOutput(&config, metars)
		if err != nil {
			logger.Fatal().Err(err).Msg("")
		}
	}
}
