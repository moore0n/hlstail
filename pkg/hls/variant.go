package hls

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

const streamInf = "#EXT-X-STREAM-INF:"

var headerTags = []string{
	"EXT-X-VERSION",
	"EXT-X-TARGETDURATION",
	"EXT-X-MEDIA-SEQUENCE",
	"EXT-X-DISCONTINUITY-SEQUENCE",
	"EXT-X-ENDLIST",
	"EXT-X-PLAYLIST-TYPE",
	"EXT-X-I-FRAMES-ONLY",
}

// Variant is a struct for storing data about a variant.
type Variant struct {
	Tags             []string
	URL              string
	Resolution       string
	Bandwidth        int
	Codecs           string
	Segments         [][]string
	previousSegments [][]string
	rawData          string
}

// Process will loop through the tags and populate convenience properties.
func (v *Variant) Process() {
	for _, tag := range v.Tags {
		if strings.Index(tag, streamInf) == 0 {
			raw := strings.ReplaceAll(tag, streamInf, "")

			// Split up the values based on their key=values which are comman delimited
			parts := strings.Split(raw, ",")

			for _, part := range parts {
				// Catch the edge case where codes have a comma in the middle of them.
				if strings.Index(part, "=") == -1 {
					v.Codecs = fmt.Sprintf("%s%s", v.Codecs, strings.ReplaceAll(part, "\"", ""))
					continue
				}

				kv := strings.Split(strings.Trim(part, " "), "=")

				fmt.Println(kv)

				key := kv[0]
				val := kv[1]

				switch key {
				case "BANDWIDTH":
					i, _ := strconv.Atoi(val)
					v.Bandwidth = i
				case "CODECS":
					v.Codecs = strings.ReplaceAll(val, "\"", "")
				case "RESOLUTION":
					v.Resolution = val
				}
			}
		}
	}
}

// Get makes the http request to get the latest data.
func (v *Variant) Get() error {
	data, err := http.Get(v.URL)

	if err != nil {
		return err
	}

	defer data.Body.Close()

	body, err := ioutil.ReadAll(data.Body)

	if err != nil {
		return err
	}

	v.rawData = string(body)
	v.Segments = parseSegments(v.rawData)

	return nil
}

// Refresh gets fresh segments that can be processed
func (v *Variant) Refresh() error {
	// Store the previous data.
	v.previousSegments = v.Segments

	// Get new information
	if err := v.Get(); err != nil {
		return errors.New("Unable to get segments")
	}

	return nil
}

// GetHeaderTagsToPrint returns a the header tags for printing.
func (v *Variant) GetHeaderTagsToPrint() string {
	// Get the first segment which would also hold the header data.
	headSegment := filterHeadTags(v.Segments[0])

	var previousHeadSegment []string

	if len(v.previousSegments) > 0 {
		previousHeadSegment = filterHeadTags(v.previousSegments[0])
	}

	// Build a buffer to manage appending the text.
	output := new(bytes.Buffer)

	for _, tag := range headSegment {
		exists := false

		for _, ptag := range previousHeadSegment {
			if ptag == tag {
				exists = true
			}
		}

		// Grey by default.
		color := "\033[38;5;250m"

		// If the tag and value don't match then we want to highlight it.
		if !exists {
			color = "\033[38;5;40m"
		}

		// Print the output, in green.
		fmt.Fprintf(output, "\r\n%s%s\033[0m", color, tag)
	}

	fmt.Fprintf(output, "\r\n\r\n")

	return output.String()
}

// GetSegmentsToPrint compiles the text list of segments to print.
func (v *Variant) GetSegmentsToPrint(count int) string {
	// Prevent out of range errors.
	if count > len(v.Segments) {
		count = len(v.Segments)
	}

	// Trim to the segments to the count that the user requested.
	segments := v.Segments[len(v.Segments)-count:]

	// Build a buffer to manage appending the text.
	output := new(bytes.Buffer)

	// Check the segments and colorize the new segments.
	for i := 0; i < len(segments); i++ {
		color := ""

		if !segmentExists(v.previousSegments, segments[i]) {
			color = "\033[38;5;40m"
		} else if i%2 == 0 {
			// Gray
			color = "\033[38;5;250m"
		}

		fmt.Fprintf(output, "\r\n%s%s\033[0m\r\n", color, strings.Join(segments[i], "\r\n"))
	}

	return output.String()
}

// Parse the segments into slices for each group of segment data.
func parseSegments(rawData string) [][]string {
	// Make a slice to store the segments to be printed.
	segments := make([][]string, 0)

	lines := strings.Split(rawData, "\n")

	var segment []string

	// Loop over the lines and create variants.
	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if line == "" {
			continue
		}

		// We've hit the ts line so we need to push this segment into the list.
		if strings.Index(line, "#") == 0 {
			segment = append(segment, line)
		} else {
			segment = append(segment, line)

			// Append to the master list of segments
			segments = append(segments, segment)

			// Create a new array.
			segment = make([]string, 0)
		}
	}

	return segments
}

// segmentExists Check if the elem exists in the prev list.
func segmentExists(prev [][]string, elem []string) bool {
	currentSegmentSource := filterSegmentSource(elem)

	// If there was an issue getting the source just return false
	if currentSegmentSource == "" {
		return false
	}

	for i := 0; i < len(prev); i++ {
		prevSegmentSource := filterSegmentSource(prev[i])

		if currentSegmentSource == prevSegmentSource {
			return true
		}
	}

	return false
}

// Use the first segment and pull out the header specific tags to print.
func filterHeadTags(segment []string) []string {

	result := make([]string, 0)

	// Loop over each line for this segment.
	for _, line := range segment {
		parts := strings.Split(line, ":")
		// ignore invalid tags.
		if len(parts) != 2 {
			continue
		}

		tag := strings.Replace(parts[0], "#", "", -1)

		// Compare this tag to each of the tags in the header tags slice.
		for _, allowedTag := range headerTags {
			if tag == allowedTag {
				result = append(result, line)
			}
		}
	}

	return result
}

func filterSegmentSource(segment []string) string {
	for _, val := range segment {
		if strings.Index(val, "#") < 0 {
			return val
		}
	}

	return ""
}
