package hls

import (
	"bytes"
	"fmt"
	"time"

	"github.com/moore0n/hlstail/pkg/term"
	"github.com/moore0n/hlstail/pkg/tools"
)

// Session Stores state information
type Session struct {
	TermSession      *term.Session
	URL              string
	Master           *Master
	Variant          *Variant
	PreviousSegments [][]string
}

// NewSession return a new session
func NewSession(termSession *term.Session, URL string) *Session {
	return &Session{
		TermSession: termSession,
		URL:         URL,
	}
}

// GetMasterPlaylistOptions return the possible playlist options.
func (sess *Session) GetMasterPlaylistOptions(width int) (string, int) {
	sess.Master = NewMaster(sess.URL)

	if err := sess.Master.Get(); err != nil {
		fmt.Println("error getting master playlist.")
		sess.TermSession.End()
		return "", 0
	}

	output := new(bytes.Buffer)

	fmt.Fprint(output, tools.PadString("[hlstail] Select a variant", width, "="), "\r\n")
	fmt.Fprint(output, tools.PadString("", width+2, " "), "\r\n")

	fmt.Fprint(output, sess.Master.GetVariantList())

	fmt.Fprint(output, tools.PadString("", width+2, " "), "\r\n")
	fmt.Fprint(output, tools.PadString("", width+2, "="), "\r\n")

	fmt.Fprint(output, "\r\nactions: (q)uit \r\n")

	return output.String(), len(sess.Master.Variants)
}

// SetVariant sets the variant used for requesting data
func (sess *Session) SetVariant(index int) {
	sess.Variant = sess.Master.Variants[index]
}

// GetVariantPrintData return the last n segments of a variant.
func (sess *Session) GetVariantPrintData(width int, count int) string {
	output := new(bytes.Buffer)

	fmt.Fprint(output, tools.PadString("[hlstail] Segment Data", width, "="))
	fmt.Fprint(output, tools.PadString("", width+2, " "))

	fmt.Fprint(output, sess.Variant.GetSegmentsToPrint(count))

	now := time.Now()
	now = now.UTC()

	fmt.Fprint(output, tools.PadString("", width+2, " "), "\r\n")
	fmt.Fprint(output, tools.PadString(now.Format(time.RFC3339), width, "="), "\r\n")

	fmt.Fprint(output, "\r\nactions: (q)uit (p)ause (r)esume (c)hange variant\r\n")

	return output.String()
}
