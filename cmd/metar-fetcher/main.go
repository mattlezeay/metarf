package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/TwiN/go-color"
	"github.com/urkk/addstogo"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type DisplayConfig struct {
	Category bool `toml:"category"`
	Winds    bool
	Ceilings bool
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

// main is the cli entry point. This will get refactored into a proper CLI but for now its doing all the heavy lifting
func main() {

	// Reads in the config primarily for the stations right now
	config := MetarConfig{}
	err := config.ReadConfig("metar-fetcher.toml")
	if err != nil {
		log.Fatalf("Unable to read config file: %s", err)
	}

	// Get a usable Url for fetching the METARs
	url, err := FormatMetarUrl(&config)
	if err != nil {
		log.Fatal(err)
	}
	// Call the endpoint
	res, getErr := http.Get(url)
	if getErr != nil {
		log.Fatal(getErr)
	}

	// Read the body which should be xml
	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	// Create the response to parse the xml
	var metars *addstogo.METARresponse
	metars, err = addstogo.UnmarshalMetars(body)
	if err != nil {
		log.Fatal(err)
	}
	// Debug code to ensure the METARs are coming through
	//fmt.Printf("First metar is: %s\n", metars.Data.METAR)

	// This is the beginning of the formatter. Right now its handling color but will need to do more as config options
	// come up
	for _, m := range metars.Data.METAR {
		output := fmt.Sprintf("%s • %s • %s", m.StationID, m.FlightCategory, m.ObservationTime.Format("020304Z"))
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
