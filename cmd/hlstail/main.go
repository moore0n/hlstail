package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/moore0n/hlstail/pkg/tools"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "hlstail"
	app.Version = "1.0.0"

	app.Usage = "Query and HLS playlist and then tail the new segments of a selected variant"

	app.Action = func(c *cli.Context) error {
		return tail(c.String("playlist"), c.Int("count"), c.Int("interval"))
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:     "playlist",
			Usage:    "The url of the master playlist",
			Required: true,
		},
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
	}

	err := app.Run(os.Args)

	if err != nil {
		log.Fatal(err)
	}
}

func tail(playlist string, count int, interval int) error {
	// Create a new HLS Session to manage the requests.
	hls := tools.NewHLSSession(playlist)

	// Get the Master and return the variant list.
	content, size := hls.GetMasterPlaylistOptions()

	// Show the variant list to the user
	tools.PrintBuffer(content)

	var selectedOption int

	// Loop until we have a valid option for a variant to tail.
	for {
		// Get which variant they want to tail.
		index, err := tools.GetOption()

		if err != nil || index > size || index < 1 {
			errMsg := fmt.Sprintf("%s\n%s%s\n", content, "Incorrect option provided, try again : ", err)
			tools.PrintBuffer(errMsg)
			continue
		}

		selectedOption = index - 1

		break
	}

	// Set the variant that was selected in the previous loop.
	hls.SetVariant(selectedOption)

	// Loop forever and request updates every n number of seconds.
	for {
		variantInfo := hls.GetVariantPrintData(count)

		tools.PrintBuffer(variantInfo)

		// Wait to get the next update.
		time.Sleep(interval * time.Second)
	}
}
