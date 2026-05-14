package hls

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestMasterGetVariantBounds(t *testing.T) {
	master := &Master{Variants: []*Variant{{URL: "first"}}}

	if _, err := master.GetVariant(-1); err == nil {
		t.Fatal("expected error for negative variant index")
	}

	if _, err := master.GetVariant(1); err == nil {
		t.Fatal("expected error for out-of-range variant index")
	}

	variant, err := master.GetVariant(0)
	if err != nil {
		t.Fatalf("expected valid variant: %v", err)
	}

	if variant.URL != "first" {
		t.Fatalf("expected first variant, got %q", variant.URL)
	}
}

func TestVariantPrintDataHandlesEmptyAndInvalidCounts(t *testing.T) {
	variant := &Variant{}

	if got := variant.GetHeaderTagsToPrint(); got != "" {
		t.Fatalf("expected empty header output, got %q", got)
	}

	if got := variant.GetSegmentsToPrint(5); got != "" {
		t.Fatalf("expected empty segment output, got %q", got)
	}

	variant.Segments = [][]string{{"#EXTINF:1", "segment.ts"}}
	if got := variant.GetSegmentsToPrint(-1); got != "" {
		t.Fatalf("expected empty segment output for negative count, got %q", got)
	}
}

func TestParseSegments(t *testing.T) {
	segments := parseSegments(strings.Join([]string{
		"#EXTM3U",
		"#EXT-X-TARGETDURATION:4",
		"#EXTINF:4.0,",
		"segment-1.ts",
		"#EXTINF:4.0,",
		"segment-2.ts",
	}, "\n"))

	if len(segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(segments))
	}

	if got := segments[0][len(segments[0])-1]; got != "segment-1.ts" {
		t.Fatalf("expected first segment source, got %q", got)
	}

	if got := segments[1][len(segments[1])-1]; got != "segment-2.ts" {
		t.Fatalf("expected second segment source, got %q", got)
	}
}

func TestParseVariantsResolvesMediaURIAndIgnoresMalformedAttributes(t *testing.T) {
	rootURL, err := url.Parse("https://example.com/live/master.m3u8")
	if err != nil {
		t.Fatal(err)
	}

	variants := parseVariants(rootURL, strings.Join([]string{
		"#EXTM3U",
		"#EXT-X-MEDIA:TYPE=AUDIO,NAME=\"English, Main\",URI=\"audio/prog.m3u8\",BROKEN",
	}, "\n"))

	if len(variants) != 1 {
		t.Fatalf("expected 1 variant, got %d", len(variants))
	}

	if got := variants[0].URL; got != "https://example.com/live/audio/prog.m3u8" {
		t.Fatalf("expected resolved media URL, got %q", got)
	}

	if got := variants[0].Resolution; got != "English, Main" {
		t.Fatalf("expected quoted NAME with comma to be preserved, got %q", got)
	}
}

func TestParseVariantsIgnoresSubtitleMedia(t *testing.T) {
	rootURL, err := url.Parse("https://example.com/live/master.m3u8")
	if err != nil {
		t.Fatal(err)
	}

	variants := parseVariants(rootURL, strings.Join([]string{
		"#EXTM3U",
		`#EXT-X-MEDIA:TYPE=SUBTITLES,NAME="English",URI="subs/en.m3u8"`,
		`#EXT-X-MEDIA:TYPE="SUBTITLES",NAME="Spanish",URI="subs/es.m3u8"`,
		`#EXT-X-MEDIA:TYPE = SUBTITLES,NAME="French",URI="subs/fr.m3u8"`,
		`#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,NAME="CC"`,
		`#EXT-X-MEDIA:TYPE=AUDIO,NAME="English",URI="audio/en.m3u8"`,
	}, "\n"))

	if len(variants) != 1 {
		t.Fatalf("expected only the audio media variant, got %d", len(variants))
	}

	if got := variants[0].URL; got != "https://example.com/live/audio/en.m3u8" {
		t.Fatalf("expected audio media URL, got %q", got)
	}
}

func TestParseVariantsResolvesStreamURIAndProcessesAttributes(t *testing.T) {
	rootURL, err := url.Parse("https://example.com/live/master.m3u8")
	if err != nil {
		t.Fatal(err)
	}

	variants := parseVariants(rootURL, strings.Join([]string{
		"#EXTM3U",
		`#EXT-X-STREAM-INF:BANDWIDTH=2560000,CODECS="avc1.64001f,mp4a.40.2",RESOLUTION=1920x1080`,
		"video/high.m3u8",
	}, "\n"))

	if len(variants) != 1 {
		t.Fatalf("expected 1 variant, got %d", len(variants))
	}

	variant := variants[0]
	if variant.URL != "https://example.com/live/video/high.m3u8" {
		t.Fatalf("expected resolved stream URL, got %q", variant.URL)
	}

	if variant.Bandwidth != 2560000 {
		t.Fatalf("expected bandwidth 2560000, got %d", variant.Bandwidth)
	}

	if variant.Codecs != "avc1.64001f,mp4a.40.2" {
		t.Fatalf("expected parsed codecs, got %q", variant.Codecs)
	}

	if variant.Resolution != "1920x1080" {
		t.Fatalf("expected parsed resolution, got %q", variant.Resolution)
	}
}

func TestParseVariantsSortsByBandwidth(t *testing.T) {
	rootURL, err := url.Parse("https://example.com/live/master.m3u8")
	if err != nil {
		t.Fatal(err)
	}

	variants := parseVariants(rootURL, strings.Join([]string{
		"#EXTM3U",
		`#EXT-X-STREAM-INF:BANDWIDTH=3000000,RESOLUTION=1920x1080`,
		"high.m3u8",
		`#EXT-X-STREAM-INF:BANDWIDTH=800000,RESOLUTION=854x480`,
		"low.m3u8",
		`#EXT-X-STREAM-INF:BANDWIDTH=1500000,RESOLUTION=1280x720`,
		"mid.m3u8",
	}, "\n"))

	if len(variants) != 3 {
		t.Fatalf("expected 3 variants, got %d", len(variants))
	}

	got := []int{variants[0].Bandwidth, variants[1].Bandwidth, variants[2].Bandwidth}
	want := []int{800000, 1500000, 3000000}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected bandwidth order %v, got %v", want, got)
		}
	}
}

func TestGetVariantListMarksSelectedAndAudioOnly(t *testing.T) {
	master := &Master{Variants: []*Variant{
		{URL: "audio.m3u8", Bandwidth: 64000},
		{URL: "video.m3u8", Resolution: "1280x720", Bandwidth: 1280000},
	}}

	list := master.GetVariantList(1)
	if !strings.Contains(list, "1) audio-only - 64 Kbps -> audio.m3u8") {
		t.Fatalf("expected audio-only variant in list, got %q", list)
	}

	if !strings.Contains(list, "\033[0;30;47m2) 1280x720 - 1280 Kbps -> video.m3u8\033[0m") {
		t.Fatalf("expected selected variant highlighting, got %q", list)
	}
}

func TestVariantProcessParsesQuotedCodecs(t *testing.T) {
	variant := &Variant{Tags: []string{`#EXT-X-STREAM-INF:BANDWIDTH=1280000,CODECS="avc1.4d401f,mp4a.40.2",RESOLUTION=1280x720,BROKEN`}}
	variant.Process()

	if variant.Bandwidth != 1280000 {
		t.Fatalf("expected bandwidth 1280000, got %d", variant.Bandwidth)
	}

	if variant.Codecs != "avc1.4d401f,mp4a.40.2" {
		t.Fatalf("expected quoted codecs with comma, got %q", variant.Codecs)
	}

	if variant.Resolution != "1280x720" {
		t.Fatalf("expected resolution, got %q", variant.Resolution)
	}
}

func TestMasterGetRejectsNon2xxStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	err := NewMaster(server.URL).Get()
	if err == nil {
		t.Fatal("expected error for non-2xx master response")
	}

	if !strings.Contains(err.Error(), "404") {
		t.Fatalf("expected status code in error, got %q", err.Error())
	}
}

func TestVariantGetRejectsNon2xxStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	err := (&Variant{URL: server.URL}).Get()
	if err == nil {
		t.Fatal("expected error for non-2xx variant response")
	}

	if !strings.Contains(err.Error(), "500") {
		t.Fatalf("expected status code in error, got %q", err.Error())
	}
}

func TestMasterGetLoadsVariantsFromHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, strings.Join([]string{
			"#EXTM3U",
			"#EXT-X-STREAM-INF:BANDWIDTH=1280000,RESOLUTION=1280x720",
			"variant.m3u8",
		}, "\n"))
	}))
	defer server.Close()

	master := NewMaster(server.URL + "/master.m3u8")
	if err := master.Get(); err != nil {
		t.Fatalf("expected master get success: %v", err)
	}

	if len(master.Variants) != 1 {
		t.Fatalf("expected 1 variant, got %d", len(master.Variants))
	}

	if got := master.Variants[0].URL; got != server.URL+"/variant.m3u8" {
		t.Fatalf("expected resolved variant URL, got %q", got)
	}
}

func TestVariantGetLoadsSegmentsFromHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, strings.Join([]string{
			"#EXTM3U",
			"#EXT-X-TARGETDURATION:4",
			"#EXTINF:4.0,",
			"segment.ts",
		}, "\n"))
	}))
	defer server.Close()

	variant := &Variant{URL: server.URL}
	if err := variant.Get(); err != nil {
		t.Fatalf("expected variant get success: %v", err)
	}

	if len(variant.Segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(variant.Segments))
	}
}

func TestVariantRefreshPreservesPreviousSegmentsAndWrapsErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	variant := &Variant{
		URL:      server.URL,
		Segments: [][]string{{"#EXTINF:4.0,", "old.ts"}},
	}

	err := variant.Refresh()
	if err == nil {
		t.Fatal("expected refresh error")
	}

	if !strings.Contains(err.Error(), "unable to get segments") || !strings.Contains(err.Error(), "503") {
		t.Fatalf("expected wrapped status error, got %q", err.Error())
	}

	if len(variant.previousSegments) != 1 || variant.previousSegments[0][1] != "old.ts" {
		t.Fatalf("expected previous segments to be retained, got %#v", variant.previousSegments)
	}
}

func TestSessionMethodsWithHTTPPlaylists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/master.m3u8":
			fmt.Fprint(w, strings.Join([]string{
				"#EXTM3U",
				"#EXT-X-STREAM-INF:BANDWIDTH=1280000,RESOLUTION=1280x720",
				"variant.m3u8",
			}, "\n"))
		case "/variant.m3u8":
			fmt.Fprint(w, strings.Join([]string{
				"#EXTM3U",
				"#EXT-X-TARGETDURATION:4",
				"#EXTINF:4.0,",
				"segment.ts",
			}, "\n"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	sess, err := NewSession(server.URL + "/master.m3u8")
	if err != nil {
		t.Fatalf("expected session creation: %v", err)
	}

	options, err := sess.GetMasterPlaylistOptions(80, 0, false)
	if err != nil {
		t.Fatalf("expected master options: %v", err)
	}

	if !strings.Contains(options, "Select a variant") || !strings.Contains(options, "1280x720") {
		t.Fatalf("expected rendered master options, got %q", options)
	}

	if err := sess.SetVariant(0); err != nil {
		t.Fatalf("expected variant selection: %v", err)
	}

	output := sess.GetVariantPrintData(80, 1)
	if !strings.Contains(output, "Segment Data") || !strings.Contains(output, "segment.ts") {
		t.Fatalf("expected rendered variant data, got %q", output)
	}
}

func TestAttributeHelpers(t *testing.T) {
	attrs := parseTagAttributes(`A=one,B="two,with,commas",BROKEN,C=three=four`)

	if attrs["A"] != "one" {
		t.Fatalf("expected A attribute, got %q", attrs["A"])
	}

	if attrs["B"] != "two,with,commas" {
		t.Fatalf("expected quoted comma value, got %q", attrs["B"])
	}

	if attrs["C"] != "three=four" {
		t.Fatalf("expected SplitN value, got %q", attrs["C"])
	}

	if _, ok := attrs["BROKEN"]; ok {
		t.Fatal("expected malformed attribute to be ignored")
	}

	if data, ok := tagAttributeData("#EXT-X-STREAM-INF:BANDWIDTH=1"); !ok || data != "BANDWIDTH=1" {
		t.Fatalf("expected tag data, got %q %v", data, ok)
	}

	if _, ok := tagAttributeData("#EXTM3U"); ok {
		t.Fatal("expected missing attribute data to fail")
	}
}

func TestSegmentHelpersAndPrinting(t *testing.T) {
	variant := &Variant{
		Segments: [][]string{
			{"#EXT-X-TARGETDURATION:4", "#EXTINF:4.0,", "old.ts"},
			{"#EXTINF:4.0,", "new.ts"},
		},
		previousSegments: [][]string{{"#EXTINF:4.0,", "old.ts"}},
	}

	if !segmentExists(variant.previousSegments, variant.Segments[0]) {
		t.Fatal("expected old segment to exist")
	}

	if segmentExists(variant.previousSegments, variant.Segments[1]) {
		t.Fatal("expected new segment not to exist")
	}

	if got := filterSegmentSource([]string{"#EXTINF:4.0,"}); got != "" {
		t.Fatalf("expected no segment source, got %q", got)
	}

	header := variant.GetHeaderTagsToPrint()
	if !strings.Contains(header, "#EXT-X-TARGETDURATION:4") {
		t.Fatalf("expected header tag output, got %q", header)
	}

	segments := variant.GetSegmentsToPrint(5)
	if !strings.Contains(segments, "old.ts") || !strings.Contains(segments, "new.ts") {
		t.Fatalf("expected segment output, got %q", segments)
	}
}
