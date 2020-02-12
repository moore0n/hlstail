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

// Master is a struct for interacting with the master playlist.
type Master struct {
	url      string
	rawData  string
	Variants []*Variant
}

// NewMaster creates a new Master
func NewMaster(url string) *Master {
	return &Master{
		url: url,
	}
}

// Get loads the data into memory to be used later.
func (m *Master) Get() error {
	data, err := http.Get(m.url)

	if err != nil {
		return err
	}

	defer data.Body.Close()

	body, err := ioutil.ReadAll(data.Body)

	if err != nil {
		return err
	}

	m.rawData = string(body)

	rootURLParts := strings.Split(m.url, "/")
	rootURL := strings.Join(rootURLParts[:len(rootURLParts)-1], "/")

	m.Variants = parseVariants(rootURL, m.rawData)

	return nil
}

// GetVariant returns a Variant struct representing the variant's data.
func (m *Master) GetVariant(index int) (*Variant, error) {
	if index > len(m.Variants) || index < 0 {
		return nil, errors.New("index out of range")
	}

	variant := m.Variants[index]

	return variant, nil
}

// GetVariantList gets a printable list of variants
func (m *Master) GetVariantList() string {
	output := new(bytes.Buffer)

	for i, variant := range m.Variants {

		res := variant.Resolution

		if res == "" {
			res = "audio-only"
		}

		fmt.Fprintf(output, "%d) %s - %s -> %s\r\n", i+1, res, strconv.Itoa(int(variant.Bandwidth)), variant.URL)
	}

	return output.String()
}

func parseVariants(rootURL string, rawData string) []*Variant {
	// Make a slice to store the segments to be printed.
	variants := make([]*Variant, 0)

	lines := strings.Split(rawData, "\n")

	var variant = &Variant{}

	// Loop over the lines and create variants.
	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if line == "" {
			continue
		}

		// We've hit the ts line so we need to push this segment into the list.
		if strings.Index(line, "#") == 0 {
			// segments = append(segments, segment)
			variant.Tags = append(variant.Tags, line)
		} else {
			// We've hit the source and need to push the variant into the list.
			variant.URL = line

			if strings.Index(variant.URL, "http") == -1 {
				variant.URL = fmt.Sprintf("%s/%s", rootURL, variant.URL)
			}

			variant.Process()

			variants = append(variants, variant)

			// Create a new array.
			variant = &Variant{}
		}
	}

	return variants
}
