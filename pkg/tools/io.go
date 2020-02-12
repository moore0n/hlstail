package tools

import (
	"bufio"
	"bytes"
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

// PollForVariant will prompt the user to select a variant
func PollForVariant(termSess *term.Session, content string, size int) int {
	// Show the variant list to the user
	PrintBuffer(content)

	// Loop until we have a valid option for a variant to tail.
	for {
		// Get which variant they want to tail.
		index, err := GetOption(termSess)

		if err != nil || index > size || index == 0 {
			errMsg := fmt.Sprintf("%s\n%s%s\n", content, "Incorrect option provided, try again : ", err)
			PrintBuffer(errMsg)
			continue
		}

		// Handle the quit case.
		if index == -1 {
			termSess.End()
			os.Exit(0)
		}

		return index
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

// GetHeader returns the header string with padding.
func GetHeader(width int, txt string) string {
	output := new(bytes.Buffer)
	title := fmt.Sprintf("[hlstail] %s", txt)

	fmt.Fprint(output, PadString(title, width, "="), "\r\n")

	return output.String()
}

// GetFooter returns the footer with padding.
func GetFooter(width int, txt string) string {
	output := new(bytes.Buffer)

	if txt == "" {
		width = width + 2
	}

	fmt.Fprint(output, PadString(txt, width, "="), "\r\n")

	return output.String()
}

// LogToFile is a debug method used to do linear logging, output to stdout in raw
// mode makes this extremely difficult otherwise.
func LogToFile(val interface{}) {
	s := fmt.Sprintf("%s\n", val)

	f, err := os.OpenFile("output.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return
	}

	defer f.Close()

	f.WriteString(s)
}
