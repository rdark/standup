package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile                 string
	journalDir              string
	standupDir              string
	journalWorkDoneSections []string
	standupWorkDoneSection  string
	standupSkipText         []string
	journalSkipText         []string
	journalLinkPreviousTitles []string
	journalLinkNextTitles []string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "standupnotes",
	Short: "Generate standup notes from journal entries",
	Long:  `Generate standup notes from journal entries`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var err error
		if journalDir == "" {
			journalDir = viper.GetString("journal.dir")
		}
		journalDir, err = filepath.Abs(journalDir)
		cobra.CheckErr(err)

		if standupDir == "" {
			standupDir = viper.GetString("standup.dir")
		}
		standupDir, err = filepath.Abs(standupDir)
		cobra.CheckErr(err)

		if len(journalWorkDoneSections) == 0 {
			journalWorkDoneSections = viper.GetStringSlice("journal.work_done_sections")
		}
		if standupWorkDoneSection == "" {
			standupWorkDoneSection = viper.GetString("standup.work_done_section")
		}

		if len(journalSkipText) == 0 {
			journalSkipText = viper.GetStringSlice("journal.skip_text")
		}
		if len(standupSkipText) == 0 {
			standupSkipText = viper.GetStringSlice("standup.skip_text")
		}

		if len(journalLinkPreviousTitles) == 0 {
			journalLinkPreviousTitles = viper.GetStringSlice("journal.link_previous_titles")
			if len(journalLinkPreviousTitles) == 0 {
				journalLinkPreviousTitles = []string{"Yesterday", "Previous"}
			}
		}
		if len(journalLinkNextTitles) == 0 {
			journalLinkNextTitles = viper.GetStringSlice("journal.link_next_titles")
			if len(journalLinkNextTitles) == 0 {
				journalLinkNextTitles = []string{"Tomorrow", "Next"}
			}
		}

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .standupnotes.yaml)")

	rootCmd.PersistentFlags().StringVar(&journalDir, "journal-dir", "", "journal notes directory")
	rootCmd.PersistentFlags().StringVar(&standupDir, "standup-dir", "", "standup notes directory")

	rootCmd.PersistentFlags().StringSliceVar(&journalWorkDoneSections, "journal-work-done-sections", []string{}, "journal work done sections")
	rootCmd.PersistentFlags().StringVar(&standupWorkDoneSection, "standup-work-done-section", "Worked on yesterday", "standup work done section")

	rootCmd.PersistentFlags().StringSliceVar(&standupSkipText, "standup-skip-text", []string{}, "Text lines to skip in standup notes")
	rootCmd.PersistentFlags().StringSliceVar(&journalSkipText, "journal-skip-text", []string{}, "Text lines to skip in journal notes")

	rootCmd.PersistentFlags().StringSliceVar(&standupSkipText, "journal-link-prevous-titles", []string{}, "A list of link titles to match within journal notes that should be references to the previous journal")
	rootCmd.PersistentFlags().StringSliceVar(&journalSkipText, "journal-link-next-titles", []string{}, "A list of link titles to match within journal notes that should be references to the next journal")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find cwd directory.
		cwd, err := os.Getwd()
		cobra.CheckErr(err)

		// Search config in cwd with name ".standupnotes" (without extension).
		viper.AddConfigPath(cwd)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".standupnotes")
	}

	viper.SetEnvPrefix("STANDUP")

	// read in environment variables that match
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	cobra.CheckErr(err)
}
