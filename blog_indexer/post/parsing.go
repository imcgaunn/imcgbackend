package post

import (
	"errors"
	"strings"
)

func ExtractPostHeaderLines(postLines []string) ([]string, int, error) {
	// fetch metadata from the beginning of the post
	// scan until you see prelude's bottom marker.
	lastHeaderRowIdx := -1
	for pos, str := range postLines {
		if len(str) < 3 {
			continue
		}
		if str[:3] == ">>>" {
			lastHeaderRowIdx = pos
			break
		}
	}
	if lastHeaderRowIdx > 0 {
		headerLines := postLines[:lastHeaderRowIdx]
		return headerLines, lastHeaderRowIdx, nil
	}
	return nil, lastHeaderRowIdx, errors.New("failed to extract header")
}

func ParseHeaderLines(lines []string) map[string]string {
	metaDataMap := make(map[string]string)
	for line := range lines {
		newMap, err := parseHeaderLine(lines[line])
		if err != nil {
			panic("encountered bad header line bye bye")
		}
		for newKey, val := range newMap {
			metaDataMap[newKey] = val
		}
	}
	return metaDataMap
}

func parseHeaderLine(line string) (map[string]string, error) {
	components := strings.Split(line, ":")
	if len(components) < 2 {
		return nil, errors.New("invalid header line")
	}
	metaData := map[string]string{
		strings.TrimSpace(components[0]): strings.TrimSpace(components[1]),
	}
	return metaData, nil
}
