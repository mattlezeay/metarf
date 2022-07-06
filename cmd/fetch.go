package cmd

import (
	"fmt"
	"github.com/TwiN/go-color"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/urkk/addstogo"
	"io/ioutil"
	"metarFetcher/internal"
	"net/http"
	"os"
	"strings"
	"time"
)

// DisplayConfig holds the display options like winds, flight category, and if it is long lived
type DisplayConfig struct {
	Category  bool `toml:"category"`
	Winds     bool `toml:"winds"`
	Ceilings  bool `toml:"ceilings"`
	LongLived bool `toml:"longlived"`
	I3Json    bool `toml:"i3json"`
}

// Metars is a container for stations and related data
type Metars struct {
	Stations []string
	Cached   bool
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
		var output string
		if config.Display.I3Json {
			output = fmt.Sprintf("{\"full_text\": \"%s", m.StationID)
		} else {
			output = fmt.Sprintf("%s", m.StationID)
		}

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
		if config.Display.I3Json {
			switch m.FlightCategory {
			case "VFR":
				output += "\", \"color\": \"#008000\" }"
			case "MVFR":
				output += "\", \"color\": \"#0000FF\" }"
			case "IFR":
				output += "\", \"color\": \"#FF0000\" }"
			case "LIFR":
				output += "\", \"color\": \"#FF00FF\" }"
			default:
				output += "\", \"color\": \"#FFFFFF\" }"
			}
			fmt.Print(output)
		} else {
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

// shouldPull is the beginning of caching, its not ready yet
// TODO: complete this function and caching logic
func shouldPull() bool {
	// Check if cache file is there, if it is, check if the data is stale
	if _, err := os.Stat("/tmp/metarf_cache.tmp"); err == nil {
		return false
	}
	return true
}

// init is part of the cobra command prep
func init() {
	rootCmd.AddCommand(fetchCmd)
}

// fetchCmd is for wiring up cobra command line things. This is the default command as well
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch METARs from avaitionweather.gov",
	Long: `This is the primary use of metar-fetcher. This will by default find and read in the config file
and fetch all the relevant METARs for the configured stations`,
	Run: func(cmd *cobra.Command, args []string) {
		fetch()
	},
}

// fetch is the main cli entry point. This will get refactored into a proper CLI but for now its doing all the heavy lifting
func fetch() {

	// Reads in the config primarily for the stations right now
	config := MetarConfig{}
	// Use Viper to read config (pairs nicely with cobra)
	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to read config file")
	}

	// Set up the logger
	logger := internal.NewLogger(false)
	defer logger.Close()
	logger.Info().Msgf("Using config file: %s", viper.ConfigFileUsed())

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
