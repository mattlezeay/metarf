/*
Copyright © 2022 Matthew Yeazel <mattlezeay@gmail.com>

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "metar-fetcher",
	Short: "A simple CLI to fetch METAR information and display it",
	Long: `metar-fetcher is a simple tool to grab METAR data on the CLI. The intention is to
use this tool for displaying METAR info in a spare terminal, i3status-type interface,
or other non-graphical method.`,
	Run: func(cmd *cobra.Command, args []string) {
		fetch()
	},
}
var cfgFile string
var stationsFlags []string

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Setup config file first
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.metar-fetcher.toml)")
	// Add override for stations if the user just wants to provide those
	rootCmd.PersistentFlags().StringArray("stations", stationsFlags, "Help message for toggle")
	err := viper.BindPFlag("metars.stations", rootCmd.PersistentFlags().Lookup("stations"))

	if err != nil {
		fmt.Errorf("unable to bind to stations: %s", err)
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// TODO: This is probably not the final list of search paths but for now is working
		viper.AddConfigPath(home + "/.config/")
		viper.AddConfigPath(".")
		viper.SetConfigType("toml")
		viper.SetConfigName("metar-fetcher")
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("No configuration file found, using defaults")
		} else {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}
}
