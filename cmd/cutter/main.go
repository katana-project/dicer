package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"runtime/debug"
	"strconv"
	"syscall"
	"time"

	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/run-slicer/attach"
	"github.com/run-slicer/cutter/api"
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
			Padding(0, 1, 0, 1).
			Background(lipgloss.Color("86")).
			Foreground(lipgloss.Color("0")),
		log.WarnLevel: lipgloss.NewStyle().
			SetString("WARN").
			Padding(0, 1, 0, 1).
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
			Foreground(lipgloss.Color("86")),
		log.WarnLevel: lipgloss.NewStyle().
			SetString("WARN").
			Bold(true).
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
	port    int
)

func findPort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}

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
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			displayBanner(time.Since(t))

			provider, err := attach.Default()
			if err != nil {
				return fmt.Errorf("failed to get default provider, are you on a supported system? (%w)", err)
			}
			log.Debug("using attach provider", "name", reflect.TypeOf(provider).Elem().Name())

			if port == 0 {
				var err error
				port, err = findPort()
				if err != nil {
					return fmt.Errorf("failed to find free port: %w", err)
				}

				log.Debug("found free port", "port", port)
			}

			server := api.NewServer(provider)
			h := api.HandlerWithOptions(api.NewStrictHandler(server, nil), api.StdHTTPServerOptions{
				BaseURL: "/api",
				Middlewares: []api.MiddlewareFunc{
					api.Logging,
				},
			})
			s := &http.Server{
				Handler: h,
				Addr:    fmt.Sprintf("localhost:%d", port),
			}

			go func() {
				sigChan := make(chan os.Signal, 1)
				signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
				<-sigChan

				timeoutCtx, release := context.WithTimeout(context.Background(), 10*time.Second)
				defer release()

				log.Info("shutting down gracefully, thanks for using cutter!")
				if err := s.Shutdown(timeoutCtx); err != nil {
					log.Fatal("failed to shutdown HTTP server", "error", err)
				}
			}()

			log.Info("listening for requests", "port", port)
			if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				return fmt.Errorf("HTTP server failed: %w", err)
			}
			return nil
		},
	}
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	cmd.Flags().IntVarP(&port, "port", "p", 0, "port to listen on (default: find a free port)")

	if err := fang.Execute(context.Background(), cmd); err != nil {
		os.Exit(1)
	}
}
