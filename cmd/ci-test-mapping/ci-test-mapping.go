package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var logLevel string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ci-test-mapping",
	Short: "ci-test-mapping maps a test to component owners and capabilities",
	Long:  "ci-test-mapping maps a test to component owners and capabilities",
}

func Execute() {
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		level, err := log.ParseLevel(logLevel)
		if err != nil {
			log.WithError(err).Fatal("cannot parse log-level")
			return err
		}
		log.SetLevel(level)
		log.SetOutput(os.Stderr)
		return nil
	}

	err := rootCmd.Execute()
	if err != nil {
		log.WithError(err).Fatal("could not execute root command")
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info",
		"Log level (trace,debug,info,warn,error) (default info)")
}
