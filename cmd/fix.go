package cmd

import (
	"fmt"
	"time"
	"os"
	"slices"

	"github.com/rdark/standupnotes/internal/util"
	"github.com/rdark/standupnotes/internal/markdown"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(fixStandupCmd)
	rootCmd.AddCommand(fixJournalCmd)
}

var fixStandupCmd = &cobra.Command{
	Use:   "fix-standup",
	Short: "fix standup note and copy work done from the previous days journal into it",
	Long: `fix the standup note for today using the configured external program
and copy the work done from the previous day's journal into it`,
	Run: fixStandupCmdFunc,
}

func fixStandupCmdFunc(cmd *cobra.Command, args []string) {

// 	createCmd := strings.Split(fixStandupCmd, " ")
// 
// 	standupNote, err := util.ExecReturnStdOut(createCmd)
// 	cobra.CheckErr(err)
// 
// 	fmt.Println(standupNote)

}

var fixJournalCmd = &cobra.Command{
	Use:   "fix-journal",
	Short: "fix journal note and copy work done from the previous days journal into it",
	Long: `fix the journal note for today using the configured external program
and copy the work done from the previous day's journal into it`,
	Run: fixJournalCmdFunc,
}

func fixJournalCmdFunc(cmd *cobra.Command, args []string) {
	dt, err := time.Parse("2006-01-02", date)
	cobra.CheckErr(err)

	journalFile, err := util.GetExactMdFileName(journalDir, dt)
	cobra.CheckErr(err)

	content, err := os.ReadFile(journalFile)
	cobra.CheckErr(err)

	parser := markdown.NewParser()
	md, err := parser.ParseNoteContent(string(content), journalSkipText, markdown.NoteTypeJournal)
	cobra.CheckErr(err)

	var fixableLinks []markdown.AdjacentLink
	for _, link := range md.AdjacentLinks {
		if link.TargetNoteType == markdown.NoteTypeJournal {
			if slices.Contains(journalLinkPreviousTitles, link.Title) && link.Target != journalFile {
				fixableLinks = append(fixableLinks, link)
			}
		}
	}

	if len(fixableLinks) > 0 {
		fmt.Printf("Found %d links to fix\n", len(fixableLinks))
	}

	// now := time.Now()

	// previousDt := now.AddDate(0, 0, -1)

	// previousJournalName, err := util.GetMostRecentMdFileName(journalDir, previousDt)
	// cobra.CheckErr(err)

	// journalPath, err := util.ExecReturnStdOut(createCmd)
	// cobra.CheckErr(err)

	// content, err := os.ReadFile(journalPath)
	// cobra.CheckErr(err)

	// parser := markdown.NewParser()
	// md, err := parser.ParseNoteContent(string(content), journalSkipText, markdown.NoteTypeJournal)
	// cobra.CheckErr(err)

	// var fixableLinks []markdown.AdjacentLink
	// for _, link := range md.AdjacentLinks {
	// 	if link.TargetNoteType == markdown.NoteTypeJournal {
	// 		if slices.Contains(journalLinkPreviousTitles, link.Title) && link.Target != previousJournalName {
	// 			fixableLinks = append(fixableLinks, link)
	// 		}
	// 	}
	// }


}
