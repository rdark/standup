package util

import (
	"os"
	"regexp"
	"time"
)

// GetMostRecentMdFileName returns the most recent markdown (.md) extension file 30 days of endTime
func GetMostRecentMdFileName(dirPath string, endTime time.Time) (string, error) {
	thirtyDaysAgo := endTime.AddDate(0, 0, -30)

	// Compile regex for YYYY-MM-DD.md format
	pattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}\.md$`)

	var mostRecentFile string
	var mostRecentDate time.Time

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		filename := entry.Name()
		if !pattern.MatchString(filename) {
			continue
		}

		// Parse date from filename (removing .md extension)
		fileDate, err := time.Parse("2006-01-02", filename[:10])
		if err != nil {
			continue
		}

		if fileDate.After(thirtyDaysAgo) && fileDate.Before(endTime) || fileDate.Equal(endTime) {
			if mostRecentFile == "" || fileDate.After(mostRecentDate) {
				mostRecentFile = filename
				mostRecentDate = fileDate
			}
		}
	}

	return mostRecentFile, nil
}
