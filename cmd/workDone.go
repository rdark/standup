package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/rdark/standupnotes/internal/markdown"
	"github.com/rdark/standupnotes/internal/util"

	"github.com/spf13/cobra"
)

var (
	date string
)

var journalWorkDoneCmd = &cobra.Command{
	Use:   "journal-work-done",
	Short: "Export work done from the journal for a given day",
	Long: `Export work done from the journal for a given day
If a journal note does not exist for the given day, the journal directory will
be searched backwards for the newest journal within 30 days of the given date
	`,
	Run: journalWorkDoneCmdFunc,
}

func init() {
	journalWorkDoneCmd.PersistentFlags().StringVarP(&date, "date", "d", time.Now().AddDate(0, 0, -1).Format("2006-01-02"), "Date to print work done for")
	rootCmd.AddCommand(journalWorkDoneCmd)
}

func journalWorkDoneCmdFunc(cmd *cobra.Command, args []string) {
	dt, err := time.Parse("2006-01-02", date)
	cobra.CheckErr(err)

	mostRecentJournal, err := util.GetMostRecentMdFileName(journalDir, dt)
	cobra.CheckErr(err)

	content, err := os.ReadFile(path.Join(journalDir, mostRecentJournal))
	cobra.CheckErr(err)

	parser := markdown.NewParser()
	md, err := parser.ParseNoteContent(string(content), journalSkipText, markdown.NoteTypeJournal)
	cobra.CheckErr(err)

	fmt.Println(md.FrontMatter.Start)
	fmt.Println(md.FrontMatter.End)
	// fmt.Println(md.FrontMatter.Meta)


	for _, section := range md.Sections {
		for _, cfgSection := range journalWorkDoneSections {
			if strings.EqualFold(section.Title, cfgSection) {
				fmt.Printf("### %s\n", section.Title)
				fmt.Println(section.Content)
			}
		}
	}
	for _, link := range md.AdjacentLinks {
		fmt.Printf("Source Note Type: %d\n", link.SourceNoteType)
		fmt.Printf("Target Note Type: %d\n", link.TargetNoteType)
		fmt.Printf("Link Title: %s\n", link.Title)
		fmt.Printf("Link Target: %s\n", link.Target)
		fmt.Printf("Link Start: %d\n", link.LinkStart)
		fmt.Printf("Link End: %d\n", link.LinkEnd)
		fmt.Println(md.Body[link.LinkStart:link.LinkEnd])
	}
}

var standupWorkDoneCmd = &cobra.Command{
	Use:   "standup-work-done",
	Short: "Export work done from the standup for a given day",
	Long: `Export work done from the standup for a given day
If a standup note does not exist for the given day, the standup directory will
be searched backwards for the newest standup within 30 days of the given date
	`,
	Run: standupWorkDoneCmdFunc,
}

func init() {
	standupWorkDoneCmd.PersistentFlags().StringVarP(&date, "date", "d", time.Now().Format("2006-01-02"), "Date to print work done for")
	rootCmd.AddCommand(standupWorkDoneCmd)
}

func standupWorkDoneCmdFunc(cmd *cobra.Command, args []string) {
	dt, err := time.Parse("2006-01-02", date)
	cobra.CheckErr(err)

	mostRecentStandup, err := util.GetMostRecentMdFileName(standupDir, dt)
	cobra.CheckErr(err)

	content, err := os.ReadFile(path.Join(standupDir, mostRecentStandup))
	cobra.CheckErr(err)

	parser := markdown.NewParser()
	md, err := parser.ParseNoteContent(string(content), standupSkipText, markdown.NoteTypeStandup)
	cobra.CheckErr(err)

	fmt.Println(md.FrontMatter.Meta.Tags)

	for _, section := range md.Sections {
		if strings.EqualFold(section.Title, standupWorkDoneSection) {
			fmt.Printf("### %s\n", section.Title)
			fmt.Println(section.Content)
			// fmt.Println("here")
			// fmt.Println(string(md.Body[section.ContentStart:section.ContentEnd]))
		}
	}
}
