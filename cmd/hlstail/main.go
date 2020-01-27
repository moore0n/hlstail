package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/moore0n/hlstail/pkg/tools"
	"github.com/urfave/cli"
)

type appState struct {
	Paused bool
}

func main() {
	app := cli.NewApp()
	app.Name = "hlstail"
	app.Version = "1.0.1"

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
	// Create a new HLS Session to manage the requests.
	hls := tools.NewHLSSession(playlist)

	// Get the Master and return the variant list.
	content, size := hls.GetMasterPlaylistOptions()

	if variant == 0 {

		// Show the variant list to the user
		tools.PrintBuffer(content)

		// Loop until we have a valid option for a variant to tail.
		for {
			// Get which variant they want to tail.
			index, err := tools.GetOption()

			if err != nil || index > size || index < 1 {
				errMsg := fmt.Sprintf("%s\n%s%s\n", content, "Incorrect option provided, try again : ", err)
				tools.PrintBuffer(errMsg)
				continue
			}

			variant = index - 1

			break
		}

	}

	// Set the variant that was selected in the previous loop.
	hls.SetVariant(variant)

	state := &appState{
		Paused: false,
	}

	go updateLoop(state, interval, count, hls)

	checkForPause(state)

	return nil
}

// updateLoop will query for updates at the supplied interval
func updateLoop(state *appState, interval int, count int, hls *tools.HLSSession) {
	var variantInfo string
	var nextRun int64 = time.Now().Unix()
	var lastPauseState bool = state.Paused

	// Loop forever and request updates every n number of seconds.
	for {

		// Prevent maxing out the CPU.
		time.Sleep(time.Millisecond * 250)

		if nextRun > time.Now().Unix() && lastPauseState == state.Paused {
			lastPauseState = state.Paused
			continue
		}

		if !state.Paused {
			variantInfo = hls.GetVariantPrintData(count)
			tools.PrintBuffer(variantInfo)
		} else {

			// This will print only when the state changes to pause, reduce the wonkiness of redrawing the screen
			if lastPauseState != state.Paused {
				width := tools.GetCliWidth()
				parts := strings.Split(variantInfo, "\n")
				end := parts[len(parts)-2]
				end = strings.ReplaceAll(end, "=", "")

				end = fmt.Sprintf("PAUSED @%s", end)

				parts[len(parts)-2] = tools.PadString(end, width, "=")

				tools.PrintBuffer(strings.Join(parts, "\n"))
			}
		}

		lastPauseState = state.Paused
		nextRun = time.Now().Unix() + int64(interval)
	}
}

// checkForPause will query the stdin to determine if someone has hit return to pause the tailing.
func checkForPause(state *appState) {
	// Read the std input
	reader := bufio.NewReader(os.Stdin)

	// Loop and read the input waiting for keyboard input
	for {
		r, err := reader.ReadString('\n')

		if err != nil {
			break
		}

		r = strings.ReplaceAll(r, "\n", "")

		if r == " " || r == "" {
			state.Paused = !state.Paused
		}
	}
}
