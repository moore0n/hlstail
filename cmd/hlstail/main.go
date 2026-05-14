package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/moore0n/hlstail/pkg/hls"
	"github.com/moore0n/hlstail/pkg/term"
	"github.com/moore0n/hlstail/pkg/tools"
	"github.com/urfave/cli/v2"
)

var errQuit = errors.New("quit")

type inputCommand int

const (
	commandPause inputCommand = iota
	commandResume
	commandChangeVariant
	commandQuit
)

func main() {
	app := cli.NewApp()
	app.Name = "hlstail"
	app.Version = "1.0.13"

	app.Usage = "Query an HLS playlist and then tail the new segments of a selected variant"

	app.UsageText = "hlstail [options...] <playlist>"

	app.Action = func(c *cli.Context) error {

		playlist := c.Args().Get(0)
		var variant *int

		// Validate that we have a playlist value.
		if playlist == "" {
			cli.ShowAppHelpAndExit(c, 0)
		}

		if c.IsSet("variant") {
			v := c.Int("variant")
			variant = &v
		}

		return tail(playlist, c.Int("count"), c.Int("interval"), variant)
	}

	app.Flags = []cli.Flag{
		&cli.IntFlag{
			Name:  "count",
			Usage: "The number of segments to display",
			Value: 10,
		},
		&cli.IntFlag{
			Name:  "interval",
			Usage: "The number of seconds to wait between updates",
			Value: 3,
		},
		&cli.IntFlag{
			Name:        "variant",
			Usage:       "The zero-based variant index you'd like to use; omit to choose interactively",
			DefaultText: "interactive",
			Value:       0,
		},
	}

	err := app.Run(os.Args)

	if err != nil {
		log.Fatal(err)
	}
}

func tail(playlist string, count int, interval int, variant *int) error {
	termSess := term.NewSession()

	if err := termSess.MakeRaw(); err != nil {
		return err
	}

	// Start the new terminal session
	termSess.Start()
	defer termSess.End()
	stopSignals := restoreTerminalOnSignal(termSess)
	defer stopSignals()

	// Print the loading screen here before we make the request.
	width, err := termSess.GetCliWidth()
	if err != nil {
		return err
	}
	tools.PrintLoading(width)

	// Create a new HLS Session to manage the requests.
	hls, err := hls.NewSession(playlist)

	if err != nil {
		return err
	}

	for {
		selectedVariant := 0

		if variant == nil {
			selectedVariant, err = PollForVariant(termSess, hls)

			if err != nil {
				if errors.Is(err, errQuit) {
					return nil
				}

				return err
			}
		} else {
			selectedVariant = *variant
		}

		// Set the variant that was selected in the previous loop.
		if err := hls.SetVariant(selectedVariant); err != nil {
			return err
		}

		commands := make(chan inputCommand, 1)

		// Run the updates in a go routine but respect input commands.
		go updateLoop(termSess, interval, count, hls, commands)

		// Run the loop to poll input for commands.
		if err := PollForInput(commands); err != nil {
			if errors.Is(err, errQuit) {
				return nil
			}

			return err
		}

		// Reset the variant so that we can prompt for variant selection if the user selects that option
		variant = nil
	}
}

func restoreTerminalOnSignal(termSess *term.Session) func() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	go func() {
		if _, ok := <-signals; !ok {
			return
		}

		termSess.End()
		os.Exit(1)
	}()

	return func() {
		signal.Stop(signals)
		close(signals)
	}
}

// PollForInput will query the stdin to determine if someone has entered a command
func PollForInput(commands chan<- inputCommand) error {
	// Read the std input
	reader := bufio.NewReader(os.Stdin)

	// Loop and read the input waiting for keyboard input
	for {
		r, _, err := reader.ReadRune()

		if err != nil {
			commands <- commandQuit
			return err
		}

		switch r {
		case rune(112):
			// (p)ause
			commands <- commandPause
		case rune(114):
			// (r)esume
			commands <- commandResume
		case rune(99):
			// (c)hange variant
			commands <- commandChangeVariant
			return nil
		case rune(113):
			// (q)uit
			commands <- commandQuit
			return errQuit
		}
	}
}

// PollForVariant will prompt the user to select a variant
func PollForVariant(termSess *term.Session, hls *hls.Session) (int, error) {
	selectedIndex := 0

	width, err := termSess.GetCliWidth()
	if err != nil {
		return 0, err
	}

	// Get the Master and return the variant list.
	content, err := hls.GetMasterPlaylistOptions(width, selectedIndex, true)
	if err != nil {
		return 0, err
	}

	// Show the variant list to the user
	tools.PrintBuffer(content)

	// Read the std input
	reader := bufio.NewReader(os.Stdin)

	// Loop until we have a valid option for a variant to tail.
	for {
		r, _, err := reader.ReadRune()

		if err != nil {
			return 0, err
		}

		switch r {
		case rune(113):
			// (q)uit
			return 0, errQuit
		case rune(114):
			// (r)efresh
			width, err = termSess.GetCliWidth()
			if err != nil {
				return 0, err
			}
			selectedIndex = 0
			// Get the Master and return the variant list.
			content, err = hls.GetMasterPlaylistOptions(width, selectedIndex, false)
			if err != nil {
				return 0, err
			}
			// Reprint the variant list.
			tools.PrintBuffer(content)
			// Continue to monitor user input
			continue
		case rune(13):
			// enter key
			return selectedIndex, nil
		case rune(27):
			// Ignore control characters
			continue
		case rune(91):
			// Ignore control characters
			continue
		case rune(65):
			// Up arrow
			width, err = termSess.GetCliWidth()
			if err != nil {
				return 0, err
			}

			if selectedIndex > 0 {
				selectedIndex--
			}

			// Get the Master and return the variant list.
			content, err = hls.GetMasterPlaylistOptions(width, selectedIndex, false)
			if err != nil {
				return 0, err
			}
			// Reprint the variant list.
			tools.PrintBuffer(content)
			// Continue to monitor user input
			continue
		case rune(66):
			// Down arrow
			width, err = termSess.GetCliWidth()
			if err != nil {
				return 0, err
			}

			if selectedIndex < len(hls.Master.Variants)-1 {
				selectedIndex++
			}

			// Get the Master and return the variant list.
			content, err = hls.GetMasterPlaylistOptions(width, selectedIndex, false)
			if err != nil {
				return 0, err
			}
			// Reprint the variant list.
			tools.PrintBuffer(content)
			// Continue to monitor user input
			continue
		}
	}
}

// updateLoop will query for updates at the supplied interval
func updateLoop(termSess *term.Session, interval int, count int, hls *hls.Session, commands <-chan inputCommand) {
	var variantInfo string
	var nextRun int64 = time.Now().Unix()
	var paused bool
	var lastPauseState bool

	// Loop forever and request updates every n number of seconds.
	for {
		// Prevent maxing out the CPU.
		time.Sleep(time.Millisecond * 50)

		select {
		case command := <-commands:
			switch command {
			case commandPause:
				paused = true
			case commandResume:
				paused = false
			case commandChangeVariant:
				// clear the previous segments
				hls.Variant.Segments = make([][]string, 0)
				return
			case commandQuit:
				return
			}
		default:
		}

		// Check timer and statechange. If we are still paused then don't update the screen.
		if nextRun > time.Now().Unix() && lastPauseState == paused {
			continue
		}

		if !paused {
			width, err := termSess.GetCliWidth()
			if err != nil {
				return
			}
			variantInfo = hls.GetVariantPrintData(width, count)
			tools.PrintBuffer(variantInfo)
		} else {

			// This will print only when the state changes to pause, reduce the wonkiness of redrawing the screen
			if lastPauseState != paused {
				width, err := termSess.GetCliWidth()
				if err != nil {
					return
				}
				parts := strings.Split(variantInfo, "\r\n")
				if len(parts) < 4 {
					continue
				}

				footerIndex := len(parts) - 4
				end := parts[footerIndex]
				end = strings.ReplaceAll(end, "=", "")

				end = strings.Trim(end, " ")

				end = fmt.Sprintf("PAUSED @%s", end)

				parts[footerIndex] = tools.PadString(end, width, "=")

				// Trim the pause instructions.
				tools.PrintBuffer(strings.Join(parts, "\r\n"))
			}
		}

		lastPauseState = paused
		nextRun = time.Now().Unix() + int64(interval)
	}
}
