package tools

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/moore0n/hlstail/pkg/term"
)

// PrintBuffer prints the value of val to stdout
func PrintBuffer(val interface{}) {
	fmt.Printf("\033[1;1H\033[0J%v", val)
}

// GetOption returns a selected option from stdin
func GetOption(termSess *term.Session) (int, error) {
	option := 0

	// Read the std input
	reader := bufio.NewReader(os.Stdin)

	// Loop and read the input waiting for keyboard input
	for {
		r, _, err := reader.ReadRune()

		if err != nil {
			return option, err
		}

		/**
		* Ignore the escape sequence that could come back in the read rune.
		* 27 == ESC
		* 91 == [
		 */
		if r == rune(27) || r == rune(91) {
			continue
		}

		switch r {
		case rune(113):
			// (q)uit
			termSess.End()
			os.Exit(0)
		}

		// If the value provided is not an int continue until a valid value is provided.
		if option, err = strconv.Atoi(string(r)); err != nil {
			continue
		}

		break
	}

	return option, nil
}

// PadString returns a new padded string
func PadString(content string, width int, char string) string {
	strLength := len(content)

	// -2 is a space on either side.
	paddingRoom := (width - strLength - 2) / 2

	rawPadding := []string{}

	for i := 0; i < paddingRoom; i++ {
		rawPadding = append(rawPadding, char)
	}

	if content != "" {
		content = fmt.Sprintf(" %s ", content)
	}

	padding := strings.Join(rawPadding, "")

	return fmt.Sprintf("%s%s%s", padding, content, padding)
}

// CheckForPause will query the stdin to determine if someone has hit return to pause the tailing.
func CheckForPause(termSess *term.Session) {
	// Read the std input
	reader := bufio.NewReader(os.Stdin)

	// Loop and read the input waiting for keyboard input
	for {
		r, _, err := reader.ReadRune()

		if err != nil {
			break
		}

		/**
		* Ignore the escape sequence that could come back in the read rune.
		* 27 == ESC
		* 91 == [
		 */
		if r == rune(27) || r == rune(91) {
			continue
		}

		switch r {
		case rune(112):
			// (p)ause
			termSess.Paused = true
		case rune(114):
			// (r)esume
			termSess.Paused = false
		case rune(113):
			// (q)uit
			termSess.End()
			os.Exit(0)
		}
	}
}
