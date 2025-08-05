package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
)

// Shades of green similar to GitHub's contribution graph
var greens = []color.Attribute{
	color.BgBlack,   // No contributions
	color.BgGreen,   // Low
	color.BgHiGreen, // Medium
	color.BgHiWhite, // High (simulate with bright color)
	color.BgWhite,   // Very High (simulate with white)
}

var style = lipgloss.NewStyle().
	Bold(true).
	Background(lipgloss.Color("#30383aff")).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#ffffff")).
	BorderStyle(lipgloss.NormalBorder()).
	PaddingTop(1).
	PaddingLeft(1).
	PaddingRight(2).
	PaddingBottom(0)

type Commit struct {
	Hash      string
	Author    string
	Timestamp time.Time
}

type CommitHistory struct {
	Author  string
	Commits []Commit
}

type Config struct {
	Author string `yaml:"author"`
}

// Enums to strings shorthand
func MonthString(m time.Month) string {
	switch m {
	case time.January:
		return "Jan"
	case time.February:
		return "Feb"
	case time.March:
		return "Mar"
	case time.April:
		return "Apr"
	case time.May:
		return "May"
	case time.June:
		return "Jun"
	case time.July:
		return "Jul"
	case time.August:
		return "Aug"
	case time.September:
		return "Sep"
	case time.October:
		return "Oct"
	case time.November:
		return "Nov"
	case time.December:
		return "Dec"
	default:
		return ""
	}
}

// Simulate a GitHub-like calendar with 7 rows and 52 columns
const (
	rows    = 7  // Days of the week
	columns = 52 // Weeks of the year
)

func printCommitHistory(history CommitHistory, uptoDate time.Time) {
	// Go back a 7 * 52 = 364 days from the current date
	startDate := uptoDate.AddDate(0, 0, -rows*columns+1)

	// Get all commits in the last year
	commits := make([]Commit, 0)
	for _, commit := range history.Commits {
		if commit.Timestamp.After(startDate) && commit.Timestamp.Before(uptoDate) {
			commits = append(commits, commit)
		}
	}

	// Print the header with the rough month names
	fmt.Print(" ")
	for week := range columns {
		// Calculate the month for this week
		month := startDate.AddDate(0, 0, week*7).Month()
		// Print the month name
		if week%4 == 0 { // Print month name every 4 weeks
			fmt.Printf("%s ", MonthString(month))
		} else {
			fmt.Print("   ") // Print spaces for other weeks
		}
	}
	fmt.Println()
	output := ""
	for row := range rows {
		for col := 0; col < columns; col++ {
			// Calculate the date for this cell
			date := startDate.AddDate(0, 0, row*columns+col)
			level := 0 // Default level for no contributions
			// Count contributions for this date
			for _, commit := range commits {

				if commit.Timestamp.Year() == date.Year() && commit.Timestamp.YearDay() == date.YearDay() {
					level++ // Increment level for each contribution on this date
				}
			}
			// Print the level in the calendar data
			if level >= len(greens) {
				level = len(greens) - 1 // Cap the level to the maximum defined
			}
			c := color.New(greens[level%len(greens)])
			output += " " + c.Sprint("  ") // Two spaces for each cell
		}
		output += "\n" // New line after each row
	}
	styledOutput := style.Render(output)
	fmt.Print(styledOutput)
}

// run git log and parse into a format similar to GitHub's contribution graph
func runGitLog(author string) (CommitHistory, error) {
	// use os/exec to run git log and parse the output
	cmd := exec.Command("git", "log", "--author="+author, "--pretty=format:%h %ad", "--date=short")
	cmd.Dir = "." // Set the working directory to the current directory
	outputbytes, err := cmd.Output()
	if err != nil {
		fmt.Printf("Failed to get output: %v", err)
		return CommitHistory{}, err
	}
	// Split the output into lines
	output := string(outputbytes)
	if len(output) == 0 {
		return CommitHistory{}, fmt.Errorf("no contributions found")
	}

	lines := strings.Split(output, "\n")
	commits := make([]Commit, 0, len(lines))
	for _, line := range lines {

		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue // Skip lines that don't have enough parts
		}
		hash := parts[0]
		dateStr := parts[1]
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			fmt.Printf("Failed to parse date %s: %v\n", dateStr, err)
			continue // Skip lines with invalid dates
		}
		commits = append(commits, Commit{
			Hash:      hash,
			Author:    author,
			Timestamp: date,
		})

	}

	return CommitHistory{Author: author, Commits: commits}, nil
}

func main() {
	fmt.Println("Git Contribution Calendar:")
	yamlFile, err := os.ReadFile("gitcal.conf")
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		return
	}
	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		fmt.Printf("Error parsing config file: %v\n", err)
		return
	}
	authorName := config.Author
	if authorName == "" {
		fmt.Println("No author specified in config file")
		return
	}

	// Create a graphql client to connect to GitHub's GraphQL API
	commitHistory, err := runGitLog(authorName)
	if err != nil {
		fmt.Printf("Error running git log: %v\n", err)
		return
	}
	now := time.Now()
	printCommitHistory(commitHistory, now)
}
