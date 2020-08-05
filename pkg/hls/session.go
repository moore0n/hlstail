package hls

import (
	"bytes"
	"fmt"
	"time"

	"github.com/moore0n/hlstail/pkg/tools"
)

// Session Stores state information
type Session struct {
	URL     string
	Master  *Master
	Variant *Variant
}

// NewSession return a new session
func NewSession(URL string) (*Session, error) {
	sess := &Session{
		URL: URL,
	}

	sess.Master = NewMaster(sess.URL)

	if err := sess.Master.Get(); err != nil {
		return nil, err
	}

	return sess, nil
}

// GetMasterPlaylistOptions return the possible playlist options.
func (sess *Session) GetMasterPlaylistOptions(width int) string {
	sess.Master = NewMaster(sess.URL)

	// Print the loading screen here before we make the request.
	tools.PrintLoading(width)

	if err := sess.Master.Get(); err != nil {
		fmt.Println("error getting master playlist.")
		return ""
	}

	output := new(bytes.Buffer)

	fmt.Fprint(output, tools.GetHeader(width, " Select a variant"), "\r\n")
	fmt.Fprint(output, sess.Master.GetVariantList())
	fmt.Fprint(output, "\r\n", tools.GetFooter(width, ""))

	fmt.Fprint(output, "\r\nactions: (q)uit (r)efresh\r\n")

	return output.String()
}

// SetVariant sets the variant used for requesting data
func (sess *Session) SetVariant(index int) {
	sess.Variant = sess.Master.Variants[index]
}

// GetVariantPrintData return the last n segments of a variant.
func (sess *Session) GetVariantPrintData(width int, count int) string {
	output := new(bytes.Buffer)

	fmt.Fprint(output, tools.GetHeader(width, " Segment Data"))

	if err := sess.Variant.Refresh(); err != nil {
		fmt.Fprint(output, err.Error())
	} else {
		fmt.Fprint(output, sess.Variant.GetHeaderTagsToPrint())
		fmt.Fprint(output, tools.GetSeparator(width, "-"))
		fmt.Fprint(output, sess.Variant.GetSegmentsToPrint(count))
	}

	fmt.Fprint(output, "\r\n", tools.GetFooter(width, time.Now().UTC().Format(time.RFC3339)))

	fmt.Fprint(output, "\r\nactions: (q)uit (p)ause (r)esume (c)hange variant\r\n")

	return output.String()
}
