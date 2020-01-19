package tools

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// PrintBuffer prints the string to stdout
func PrintBuffer(text string) {
	fmt.Printf("\033[2J\033[1;1H%s\n", text)
}

// GetOption returns a selected option from stdin
func GetOption() (int, error) {
	option := 0

	// Read the std input
	reader := bufio.NewReader(os.Stdin)

	// Loop and read the input waiting for keyboard input
	for {
		input, err := read(reader)

		if input == "" {
			continue
		}

		if err != nil {
			return option, err
		}

		// Replace the new line character
		input = strings.TrimRight(input, "\n")

		if option, err = strconv.Atoi(input); err != nil {
			return option, err
		}

		break
	}

	return option, nil
}

// Read a line from the buffer
func read(r *bufio.Reader) (string, error) {
	text, err := r.ReadString('\n')

	if err != nil {
		return "", errors.Wrap(err, "Error reading from stdin")
	}

	return text, nil
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

	return fmt.Sprintf("%s%s%s\n", padding, content, padding)
}

// GetCliWidth returns the available screen space
func GetCliWidth() int {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin

	out, err := cmd.Output()

	if err != nil {
		fmt.Println("Error : ", err.Error())
		os.Exit(1)
	}

	parts := strings.Split(string(out), " ")

	rawWidth := strings.ReplaceAll(parts[1], "\n", "")

	width, err := strconv.Atoi(rawWidth)

	if err != nil {
		fmt.Println("Error : ", err.Error())
		os.Exit(1)
	}

	return width
}
