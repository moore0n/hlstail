package tools

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

// PrintBuffer prints the value of val to stdout
func PrintBuffer(val interface{}) {
	fmt.Printf("\033[1;1H\033[0J%v", val)
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
