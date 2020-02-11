package main

import (
	"fmt"
	"log"
	"os"
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
	app.Version = "1.0.5"

	app.Usage = "Query an HLS playlist and then tail the new segments of a selected variant"

	app.UsageText = "[playlist]"

	app.Action = func(c *cli.Context) error {

		playlist := c.Args().Get(0)

		// Validate that we have a playlist value.
		if len(playlist) == 0 {
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

	width := termSess.GetCliWidth()

	// Create a new HLS Session to manage the requests.
	hls := hls.NewSession(termSess, playlist)

	// Get the Master and return the variant list.
	content, size := hls.GetMasterPlaylistOptions(width)

	if variant == 0 {

		// Show the variant list to the user
		tools.PrintBuffer(content)

		// Loop until we have a valid option for a variant to tail.
		for {
			// Get which variant they want to tail.
			index, err := tools.GetOption(termSess)

			if err != nil || index > size || index == 0 {
				errMsg := fmt.Sprintf("%s\n%s%s\n", content, "Incorrect option provided, try again : ", err)
				tools.PrintBuffer(errMsg)
				continue
			}

			// Handle the quit case.
			if index == -1 {
				termSess.End()
				return nil
			}

			variant = index - 1

			break
		}

	}

	// Set the variant that was selected in the previous loop.
	hls.SetVariant(variant)

	// Run the updates in a go routine but respect the pause state.
	go updateLoop(termSess, interval, count, hls)

	// Run the loop to poll input for pause.
	tools.CheckForPause(termSess)

	return nil
}

// updateLoop will query for updates at the supplied interval
func updateLoop(termSess *term.Session, interval int, count int, hls *hls.Session) {
	var variantInfo string
	var nextRun int64 = time.Now().Unix()
	var lastPauseState bool = termSess.Paused

	// Loop forever and request updates every n number of seconds.
	for {

		// Prevent maxing out the CPU.
		time.Sleep(time.Millisecond * 250)

		if nextRun > time.Now().Unix() && lastPauseState == termSess.Paused {
			lastPauseState = termSess.Paused
			continue
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
