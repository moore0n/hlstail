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

	if err := sess.Master.Get(); err != nil {
		fmt.Println("error getting master playlist.")
		return ""
	}

	output := new(bytes.Buffer)

	fmt.Fprint(output, tools.GetHeader(width, "Select a variant"), "\r\n")
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

	fmt.Fprint(output, tools.GetHeader(width, "Segment Data"))
	fmt.Fprint(output, sess.Variant.GetSegmentsToPrint(count))

	now := time.Now()
	now = now.UTC()

	fmt.Fprint(output, "\r\n", tools.GetFooter(width, now.Format(time.RFC3339)))

	fmt.Fprint(output, "\r\nactions: (q)uit (p)ause (r)esume (c)hange variant\r\n")

	return output.String()
}
