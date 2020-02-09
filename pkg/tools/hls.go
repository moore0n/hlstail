package tools

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ingest/manifest/hls"
	"github.com/ingest/manifest/hls/source"
)

// HLSSession Stores state information
type HLSSession struct {
	URL              string
	Master           *hls.MasterPlaylist
	Variant          *hls.Variant
	Source           hls.Source
	PreviousSegments [][]string
}

// NewHLSSession return a new session
func NewHLSSession(URL string) *HLSSession {
	return &HLSSession{
		URL:    URL,
		Source: source.HTTP(nil),
	}
}

// GetMasterPlaylistOptions return the possible playlist options.
func (sess *HLSSession) GetMasterPlaylistOptions(width int) (string, int) {
	var err error

	sess.Master, err = sess.Source.Master(context.Background(), sess.URL)

	if err != nil {
		// Handle Error
		fmt.Println("Trouble reading data.")
		os.Exit(0)
	}

	variantURLs := new(bytes.Buffer)

	fmt.Fprint(variantURLs, PadString("[hlstail] Select a variant", width, "="), "\r\n")
	fmt.Fprint(variantURLs, PadString("", width+2, " "), "\r\n")
	for i, variant := range sess.Master.Variants {
		if url, err := variant.AbsoluteURL(); err == nil {
			res := variant.Resolution

			if res == "" {
				res = "audio-only"
			}

			fmt.Fprintf(variantURLs, "%d) %s - %s -> %s\r\n", i+1, res, strconv.Itoa(int(variant.Bandwidth)), url)
		}
	}

	fmt.Fprint(variantURLs, PadString("", width+2, " "), "\r\n")
	fmt.Fprint(variantURLs, PadString("", width+2, "="), "\r\n")
	fmt.Fprint(variantURLs, "\r\nactions: (q)uit \r\n")

	return variantURLs.String(), len(sess.Master.Variants)
}

// SetVariant sets the variant used for requesting data
func (sess *HLSSession) SetVariant(index int) {
	sess.Variant = sess.Master.Variants[index]
}

// GetRawVariantData Get the raw variant data as a string
func (sess *HLSSession) GetRawVariantData() (string, error) {
	path, err := sess.Variant.AbsoluteURL()

	if err != nil {
		return "", err
	}

	resp, err := http.Get(path)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	return string(body), nil
}

// GetVariantPrintData return the last n segments of a variant.
func (sess *HLSSession) GetVariantPrintData(width int, count int) string {
	rawData, err := sess.GetRawVariantData()

	if err != nil {
		return fmt.Sprintf("Error getting variant data: %s", err.Error())
	}

	// Build the segments list
	segments := sess.buildSegmentList(rawData)

	if len(segments) == 0 {
		return "Unable to get segments"
	}

	if count > len(segments) {
		count = len(segments)
	}

	segments = segments[len(segments)-count:]

	segmentData := new(bytes.Buffer)

	fmt.Fprint(segmentData, PadString("[hlstail] Segment Data", width, "="))
	fmt.Fprint(segmentData, PadString("", width+2, " "))

	// Check the segments and colorize the new segments.
	for i := 0; i < len(segments); i++ {
		color := ""

		if !exists(sess.PreviousSegments, segments[i]) {
			color = "\033[38;5;40m"
		} else if i%2 == 0 {
			// Gray
			color = "\033[38;5;250m"
		}

		fmt.Fprintf(segmentData, "%s%s\033[0m\r\n", color, strings.Join(segments[i], "\r\n"))
	}

	now := time.Now()
	now = now.UTC()

	fmt.Fprint(segmentData, PadString("", width+2, " "), "\r\n")
	fmt.Fprint(segmentData, PadString(now.Format(time.RFC3339), width, "="), "\r\n")
	fmt.Fprint(segmentData, "\r\nactions: (q)uit (p)ause (r)esume \r\n")

	sess.PreviousSegments = segments

	return segmentData.String()
}

// buildSegmentList takes the raw hls string and chunks it into an array of segments.
// A segment is itself just an array of strings.
func (sess *HLSSession) buildSegmentList(rawData string) [][]string {
	// Make a slice to store the segments to be printed.
	segments := make([][]string, 0)

	lines := strings.Split(rawData, "\n")

	var segment = make([]string, 0)

	// Loop over the lines and create segments.
	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if line == "" {
			continue
		}

		segment = append(segment, line)

		// We've hit the ts line so we need to push this segment into the list.
		if strings.Index(line, "#") != 0 {
			segments = append(segments, segment)

			// Create a new array.
			segment = make([]string, 0)
		}
	}

	return segments
}

// exists Check if the elem exists in the prev list.
func exists(prev [][]string, elem []string) bool {
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

func filterSegmentSource(segment []string) string {
	for _, val := range segment {
		if strings.Index(val, "#") < 0 {
			return val
		}
	}

	return ""
}
