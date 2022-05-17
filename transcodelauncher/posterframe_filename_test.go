package main

import "testing"

func TestPosterFrameNamesForEncoding(t *testing.T) {
	before, after, err := PosterFrameNamesForEncoding("path/to/some/key_with_underscores_512.mp4", ".png")
	if before != "path/to/some/key_with_underscores_00001_512.png" {
		t.Error("Unexpected 'before' name, got ", before)
	}
	if after != "key_with_underscores_512_poster.png" {
		t.Error("Unexpected 'after' name, got ", after)
	}
	if err != nil {
		t.Error("Unexpected error returned: ", err)
	}

	_, _, err = PosterFrameNamesForEncoding("path/to/invalidlocation.xxx", ".png")
	if err != nil {
		t.Error("Should have errored for an invalid filename")
	}
}
