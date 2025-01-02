package cmd

import (
	"fmt"
	"strings"
	"time"
	"os"
	"slices"

	"github.com/rdark/standupnotes/internal/util"
	"github.com/rdark/standupnotes/internal/markdown"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	createStandupCmd string
	createJournalCmd string
)

func init() {
	rootCmd.AddCommand(generateStandupCmd)
	rootCmd.AddCommand(generateJournalCmd)
}

var generateStandupCmd = &cobra.Command{
	Use:   "generate-standup",
	Short: "Generate standup note and copy work done from the previous days journal into it",
	Long: `Generate the standup note for today using the configured external program
and copy the work done from the previous day's journal into it`,
	Run: generateStandupCmdFunc,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.PersistentPreRun(cmd, args)
		if createStandupCmd == "" {
			createStandupCmd = viper.GetString("standup.create.cmd")
		}
	},
}

func generateStandupCmdFunc(cmd *cobra.Command, args []string) {

	if len(createStandupCmd) == 0 {
		err := fmt.Errorf("No command configured to create standup notes")
		cobra.CheckErr(err)
	}
	createCmd := strings.Split(createStandupCmd, " ")

	standupNote, err := util.ExecReturnStdOut(createCmd)
	cobra.CheckErr(err)

	fmt.Println(standupNote)

}

var generateJournalCmd = &cobra.Command{
	Use:   "generate-journal",
	Short: "Generate journal note and copy work done from the previous days journal into it",
	Long: `Generate the journal note for today using the configured external program
and copy the work done from the previous day's journal into it`,
	Run: generateJournalCmdFunc,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.PersistentPreRun(cmd, args)
		if createJournalCmd == "" {
			createJournalCmd = viper.GetString("journal.create.cmd")
		}
	},
}

func generateJournalCmdFunc(cmd *cobra.Command, args []string) {
	if len(createJournalCmd) == 0 {
		err := fmt.Errorf("No command configured to create journal notes")
		cobra.CheckErr(err)
	}
	createCmd := strings.Split(createJournalCmd, " ")

	now := time.Now()

	previousDt := now.AddDate(0, 0, -1)

	previousJournalName, err := util.GetMostRecentMdFileName(journalDir, previousDt)
	cobra.CheckErr(err)

	journalPath, err := util.ExecReturnStdOut(createCmd)
	cobra.CheckErr(err)

	content, err := os.ReadFile(journalPath)
	cobra.CheckErr(err)

	parser := markdown.NewParser()
	md, err := parser.ParseNoteContent(string(content), journalSkipText, markdown.NoteTypeJournal)
	cobra.CheckErr(err)

	var fixableLinks []markdown.AdjacentLink
	for _, link := range md.AdjacentLinks {
		if link.TargetNoteType == markdown.NoteTypeJournal {
			if slices.Contains(journalLinkPreviousTitles, link.Title) && link.Target != previousJournalName {
				fixableLinks = append(fixableLinks, link)
			}
		}
	}

	if len(fixableLinks) > 0 {
		fmt.Println("Fixing links")
		for _, link := range fixableLinks {
			fmt.Printf("Fixing link: %s\n", link.Title)
			content = slices.Replace(content, link.Target, previousJournalName)
		}
		err = os.WriteFile(journalPath, content, 0644)
		cobra.CheckErr(err)
	}

}
