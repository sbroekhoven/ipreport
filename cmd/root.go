package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	cfgFile string

	rootCmd = &cobra.Command{
		Use:   "ipreport",
		Short: "Get information from an IP address",
		Long:  `Get information from an IP address.`,
	}
)

var log = logrus.New()

// Execute executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Logging
	log.SetOutput(os.Stdout)
	// log.Out = os.Stdout
	log.SetLevel(logrus.InfoLevel)

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ipreport.yaml)")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "nameserver", "8.8.8.8", "what nameserver to use)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.WithFields(logrus.Fields{
				"Step": "homedir",
			}).Info(err)
		}

		// Search config in home directory with name ".ipreport" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".ipreport")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
