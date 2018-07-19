package post

import (
	"fmt"
	"strings"
	"testing"
)

func TestCanExtractHeaderLinesFromPost(t *testing.T) {
	postWithHdr := `key: value
pairs: are
great: things
>>>
somepostcontent
`
	headerLines, err := ExtractPostHeaderLines(strings.Split(postWithHdr, "\n"))
	if err != nil {
		t.Fatal("failed to extract perfectly valid header")
	}
	headerMap := ParseHeaderLines(headerLines)

	expectedMap := map[string]string{
		"pairs": "are",
		"great": "things",
		"key":   "value",
	}
	for k, v := range expectedMap {
		t.Log(fmt.Sprintf("k, v: [%s, %s]", k, v))
		if headerMap[k] != v {
			t.Log(headerMap[k])
			t.Fatal(fmt.Sprintf("headerMap didn't have expected result for %s", k))
		}
	}
}

func TestDoNotCrashWhenNoHeader(t *testing.T) {
	postNoHdr := `this is a post that has no header`
	headerLines, err := ExtractPostHeaderLines(strings.Split(postNoHdr, "\n"))
	if err == nil {
		t.Log("there is no header, this should have thrown!")
		t.Fatal(headerLines)
	}
}
