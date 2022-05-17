package main

import (
	"errors"
	"path"
	"regexp"
	"strings"
)

/*
PosterFrameNamesForEncoding returns two strings - the first being the _source_ name that the poster frame for the given
encoding should be named, if it's present; and the second being the _destination_ name that the endpoint wants it to be
*/
func PosterFrameNamesForEncoding(encodingKey string, imageXtn string) (string, string, error) {
	xtnRipper := regexp.MustCompile("^(.*)\\.([^.]+)")
	fileBase := path.Base(encodingKey)

	parts := xtnRipper.FindAllStringSubmatch(fileBase, -1)
	if parts == nil {
		return "", "", errors.New("encoding name had no file extension or was malformed")
	}

	fileBaseNoXtn := parts[0][1]

	//we assume that the filename is delimited by _
	fileElemts := strings.Split(fileBaseNoXtn, "_")
	if len(fileElemts) < 1 {
		return "", "", errors.New("encoding name was malformed, had no _ characters")
	}
	//so, the "_{encoding}" is the _last_ element of 'fileElemts'

	leadingPortion := strings.Join(fileElemts[0:len(fileElemts)-1], "_")
	trailingPortion := fileElemts[len(fileElemts)-1]
	return path.Dir(encodingKey) + "/" + leadingPortion + "_00001_" + trailingPortion + imageXtn, leadingPortion + "_" + trailingPortion + "_poster" + imageXtn, nil
}
