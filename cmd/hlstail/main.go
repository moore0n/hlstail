package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/moore0n/hlstail/pkg/hls"
	"github.com/moore0n/hlstail/pkg/term"
	"github.com/moore0n/hlstail/pkg/tools"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "hlstail"
	app.Version = "1.0.7"

	app.Usage = "Query an HLS playlist and then tail the new segments of a selected variant"

	app.UsageText = "hlstail [options...] <playlist>"

	app.Action = func(c *cli.Context) error {

		playlist := c.Args().Get(0)

		// Validate that we have a playlist value.
		if playlist == "" {
			cli.ShowAppHelpAndExit(c, 0)
		}

		return tail(playlist, c.Int("count"), c.Int("interval"), c.Int("variant"))
	}

	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "count",
			Usage: "The number of segments to display",
			Value: 5,
		},
		cli.IntFlag{
			Name:  "interval",
			Usage: "The number of seconds to wait between updates",
			Value: 3,
		},
		cli.IntFlag{
			Name:  "variant",
			Usage: "The number of the variant you'd like to use",
			Value: 0,
		},
	}

	err := app.Run(os.Args)

	if err != nil {
		log.Fatal(err)
	}
}

func tail(playlist string, count int, interval int, variant int) error {
	termSess := term.NewSession()

	if err := termSess.MakeRaw(); err != nil {
		return err
	}

	// Start the new terminal session
	termSess.Start()

	// Create a new HLS Session to manage the requests.
	hls, err := hls.NewSession(playlist)

	if err != nil {
		termSess.End()
		return err
	}

	for {
		if variant == 0 {
			variant, err = PollForVariant(termSess, hls)

			if err != nil {
				// (q)uit
				termSess.End()
				fmt.Println("error getting master playlist.")
				os.Exit(0)
			}
		}

		// Set the variant that was selected in the previous loop.
		hls.SetVariant(variant - 1)

		// Run the updates in a go routine but respect the pause state.
		go updateLoop(termSess, interval, count, hls)

		// Run the loop to poll input for commands.
		PollForInput(termSess)

		// Reset the variant so that we can prompt for variant selection if the user selects that option
		variant = 0
	}
}

// PollForInput will query the stdin to determine if someone has entered a command
func PollForInput(termSess *term.Session) {
	// Read the std input
	reader := bufio.NewReader(os.Stdin)

	// Loop and read the input waiting for keyboard input
	for {
		r, _, err := reader.ReadRune()

		if err != nil {
			break
		}

		switch r {
		case rune(112):
			// (p)ause
			termSess.Paused = true
		case rune(114):
			// (r)esume
			termSess.Paused = false
		case rune(99):
			// (c)hange variant
			termSess.Reset = true
			return
		case rune(113):
			// (q)uit
			termSess.End()
			os.Exit(0)
		}
	}
}

// PollForVariant will prompt the user to select a variant
func PollForVariant(termSess *term.Session, hls *hls.Session) (int, error) {
	width := termSess.GetCliWidth()

	// Get the Master and return the variant list.
	content := hls.GetMasterPlaylistOptions(width)

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
			termSess.End()
			os.Exit(0)
		case rune(114):
			// (r)efresh
			width = termSess.GetCliWidth()
			// Get the Master and return the variant list.
			content = hls.GetMasterPlaylistOptions(width)
			// Reprint the variant list.
			tools.PrintBuffer(content)
			// Continue to monitor user input
			continue
		default:
			index, err := strconv.Atoi(string(r))

			if err != nil {
				continue
			}

			if err != nil || index > len(hls.Master.Variants) || index == 0 {
				errMsg := fmt.Sprintf("%s\n%s%s\n", content, "Incorrect option provided, try again : ", err)
				tools.PrintBuffer(errMsg)
				continue
			}

			return index, nil
		}
	}
}

// updateLoop will query for updates at the supplied interval
func updateLoop(termSess *term.Session, interval int, count int, hls *hls.Session) {
	var variantInfo string
	var nextRun int64 = time.Now().Unix()
	var lastPauseState bool = termSess.Paused

	// Loop forever and request updates every n number of seconds.
	for {
		// Prevent maxing out the CPU.
		time.Sleep(time.Millisecond * 50)

		// Check timer and statechange. If we are still paused then don't update the screen.
		if nextRun > time.Now().Unix() && lastPauseState == termSess.Paused {
			continue
		}

		/**
		 *	Handle the reset here, we need to return so this go routine will die
		 * 	and we can start another one when the user selects a variant.
		 *  */
		if termSess.Reset {
			termSess.Reset = false
			termSess.Paused = false
			// clear the previous segments
			hls.Variant.Segments = make([][]string, 0)
			return
		}

		if !termSess.Paused {
			width := termSess.GetCliWidth()
			variantInfo = hls.GetVariantPrintData(width, count)
			tools.PrintBuffer(variantInfo)
		} else {

			// This will print only when the state changes to pause, reduce the wonkiness of redrawing the screen
			if lastPauseState != termSess.Paused {
				width := termSess.GetCliWidth()
				parts := strings.Split(variantInfo, "\r\n")
				end := parts[len(parts)-4]
				end = strings.ReplaceAll(end, "=", "")

				end = strings.Trim(end, " ")

				end = fmt.Sprintf("PAUSED @%s", end)

				parts[len(parts)-4] = tools.PadString(end, width, "=")

				// Trim the pause instructions.
				tools.PrintBuffer(strings.Join(parts, "\r\n"))
			}
		}

		lastPauseState = termSess.Paused
		nextRun = time.Now().Unix() + int64(interval)
	}
}
