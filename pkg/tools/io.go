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
	paddingRoom := width / 2
	strLength := len(content)

	if strLength != 0 {
		// -2 is a space on either side.
		paddingRoom = (width - strLength - 2) / 2
	}

	rawPadding := []string{}

	for i := 0; i < paddingRoom; i++ {
		rawPadding = append(rawPadding, char)
	}

	if content != "" {
		content = fmt.Sprintf(" %s ", content)
	}

	padding := strings.Join(rawPadding, "")

	result := fmt.Sprintf("%s%s%s", padding, content, padding)

	// In the case that we have an odd width or odd content we want to make sure
	// it truly stretches the width. It should only ever be off by 1.
	if len(result) != width {
		result = fmt.Sprintf("%s%s", result, char)
	}

	return result
}

// GetHeader returns the header string with padding.
func GetHeader(width int, txt string) string {
	output := new(bytes.Buffer)
	title := fmt.Sprintf("[hlstail]%s", txt)

	fmt.Fprint(output, PadString(title, width, "="), "\r\n")

	return output.String()
}

// GetFooter returns the footer with padding.
func GetFooter(width int, txt string) string {
	output := new(bytes.Buffer)

	fmt.Fprint(output, PadString(txt, width, "="), "\r\n")

	return output.String()
}

// GetSeparator returns a content separator
func GetSeparator(width int, char string) string {
	output := new(bytes.Buffer)

	fmt.Fprint(output, PadString("", width, char), "\r\n")

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

// PrintLoading takes over the screen with loading feedback
func PrintLoading(width int) {
	output := new(bytes.Buffer)

	fmt.Fprint(output, GetHeader(width, ""))
	fmt.Fprint(output, "\r\nLoading...\r\n\r\n")
	fmt.Fprint(output, GetFooter(width, ""))

	PrintBuffer(output.String())
}
