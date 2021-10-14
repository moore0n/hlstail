package hls

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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

	rootURL, err := url.Parse(m.url)
	if err != nil {
		return err
	}

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
func (m *Master) GetVariantList(selectedIndex int) string {
	output := new(bytes.Buffer)

	for i, variant := range m.Variants {

		res := variant.Resolution

		if res == "" {
			res = "audio-only"
		}

		if i == selectedIndex {
			fmt.Fprintf(output, "\033[0;30;47m%d) %s - %s -> %s\033[0m\r\n", i+1, res, strconv.Itoa(int(variant.Bandwidth)), variant.URL)
		} else {
			fmt.Fprintf(output, "%d) %s - %s -> %s\r\n", i+1, res, strconv.Itoa(int(variant.Bandwidth)), variant.URL)
		}
	}

	return output.String()
}

func parseVariants(rootURL *url.URL, rawData string) []*Variant {
	// Make a slice to store the variants to be printed.
	variants := make([]*Variant, 0)

	lines := strings.Split(rawData, "\n")

	var variant = &Variant{}

	// Loop over the lines and create variants.
	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if line == "" {
			continue
		}

		if strings.Index(line, "#") == 0 {
			variant.Tags = append(variant.Tags, line)

			// If this is a media tag then we need to parse it now rather than waiting for the source line.
			if strings.Index(line, "#EXT-X-MEDIA") == 0 {

				// Get the portion after the media tag
				data := strings.Split(line, ":")

				// If we don't have something we can parse, just continue
				if len(data) != 2 {
					continue
				}

				// Break out the key / value pairs
				parts := strings.Split(data[1], ",")

				for _, part := range parts {

					kv := strings.Split(part, "=")

					switch kv[0] {
					case "URI":

						variant.URL = strings.ReplaceAll(kv[1], "\"", "")

						if strings.Index(variant.URL, "http") == -1 {
							variant.URL = fmt.Sprintf("%s/%s", rootURL, variant.URL)
						}
					case "NAME":
						variant.Resolution = kv[1]
					default:
						// Ignore any fields we don't about for now.
						break
					}
				}

				variants = append(variants, variant)

				// Create a new variant.
				variant = &Variant{}

			}
		} else {

			// Parse the url, and then apply the protocol and host from the parent.
			u, err := url.Parse(line)
			if err != nil {
				variant.URL = "invalid-url"
			}

			// Here we want to fill in the blanks if the provided url is relative.
			if u.Host == "" {
				u.Host = rootURL.Host
				rootPath := strings.Split(rootURL.Path, "/")
				u.Path = fmt.Sprintf("%s/%s", strings.Join(rootPath[:len(rootPath)-1], "/"), u.Path)
			}

			if u.Scheme == "" {
				u.Scheme = rootURL.Scheme
			}

			variant.URL = u.String()

			variant.Process()

			variants = append(variants, variant)

			// Create a new variant.
			variant = &Variant{}
		}
	}

	return variants
}
