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
	postLines := strings.Split(postWithHdr, "\n")
	headerLines, headerEndIdx, err := ExtractPostHeaderLines(postLines)

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
	if postLines[headerEndIdx + 1] != "somepostcontent" {
		t.Logf("header end index: %d\n", headerEndIdx)
		t.Fatal("something wrong with post")
	}
}

func TestDoNotCrashWhenNoHeader(t *testing.T) {
	postNoHdr := `this is a post that has no header`
	headerLines, headerEnd, err := ExtractPostHeaderLines(strings.Split(postNoHdr, "\n"))
	if err == nil {
		t.Log("there is no header, this should have thrown!")
		t.Fatal(headerLines)
	}
	if headerEnd != -1 {
		t.Log("shouldn't have assigned a value for headerEnd")
		t.Fatal(headerEnd)
	}
}
