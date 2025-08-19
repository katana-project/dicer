package main

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var (
	defaultStyles = map[log.Level]lipgloss.Style{
		log.DebugLevel: lipgloss.NewStyle().
			SetString("DEBUG").
			Padding(0, 1, 0, 1).
			Background(lipgloss.Color("63")).
			Foreground(lipgloss.Color("0")),
		log.InfoLevel: lipgloss.NewStyle().
			SetString("INFO").
			Padding(0, 2, 0, 1).
			Background(lipgloss.Color("86")).
			Foreground(lipgloss.Color("0")),
		log.WarnLevel: lipgloss.NewStyle().
			SetString("WARN").
			Padding(0, 2, 0, 1).
			Background(lipgloss.Color("192")).
			Foreground(lipgloss.Color("0")),
		log.ErrorLevel: lipgloss.NewStyle().
			SetString("ERROR").
			Padding(0, 1, 0, 1).
			Background(lipgloss.Color("204")).
			Foreground(lipgloss.Color("0")),
		log.FatalLevel: lipgloss.NewStyle().
			SetString("FATAL").
			Padding(0, 1, 0, 1).
			Background(lipgloss.Color("134")).
			Foreground(lipgloss.Color("0")),
	}
	verboseStyles = map[log.Level]lipgloss.Style{
		log.DebugLevel: lipgloss.NewStyle().
			SetString("DEBUG").
			Bold(true).
			Foreground(lipgloss.Color("63")),
		log.InfoLevel: lipgloss.NewStyle().
			SetString("INFO").
			Bold(true).
			PaddingRight(1).
			Foreground(lipgloss.Color("86")),
		log.WarnLevel: lipgloss.NewStyle().
			SetString("WARN").
			Bold(true).
			PaddingRight(1).
			Foreground(lipgloss.Color("192")),
		log.ErrorLevel: lipgloss.NewStyle().
			SetString("ERROR").
			Bold(true).
			Foreground(lipgloss.Color("204")),
		log.FatalLevel: lipgloss.NewStyle().
			SetString("FATAL").
			Bold(true).
			Foreground(lipgloss.Color("134")),
	}
)

func configureLog(logger *log.Logger, verbose bool) {
	styles := log.DefaultStyles()
	if verbose {
		styles.Levels = verboseStyles
		logger.SetLevel(log.DebugLevel)
	} else {
		styles.Levels = defaultStyles
	}
	logger.SetStyles(styles)
	logger.SetReportCaller(verbose)
	logger.SetReportTimestamp(verbose)
	logger.SetTimeFormat(time.TimeOnly)
}

func displayBanner(t time.Duration) {
	versionStr := "development build"
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		if buildInfo.Main.Version != "(devel)" {
			versionStr = fmt.Sprintf("v%s", buildInfo.Main.Version)
		}
	}

	banner := fmt.Sprintf(
		"\n%s %s %s %s µs\n",
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("5")).
			Bold(true).
			Render("cutter"),
		versionStr,
		lipgloss.NewStyle().
			Faint(true).
			Render("ready in"),
		lipgloss.NewStyle().
			Bold(true).
			Render(strconv.FormatInt(t.Microseconds(), 10)),
	)
	fmt.Println(lipgloss.NewStyle().Padding(0, 2, 0, 2).Render(banner))
}

var (
	verbose bool
)

func main() {
	t := time.Now()

	// apply non-verbose configuration by default
	configureLog(log.Default(), false)

	cmd := &cobra.Command{
		Use:   "cutter",
		Short: "a tool for remote JVM instrumentation",
		Long:  "cutter is a CLI tool for exposing local JVM instances to remote instrumentation tools, such as https://slicer.run.",
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			configureLog(log.Default(), verbose)

			displayBanner(time.Since(t))
		},
	}
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")

	if err := fang.Execute(context.Background(), cmd); err != nil {
		os.Exit(1)
	}
}
