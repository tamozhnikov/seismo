package main

import "testing"

func Test_parseMsgFiles(t *testing.T) {
	input := struct {
		inputDir string
		saveDir  string
	}{
		"testdata/2022-February",
		"testout/2022-February",
	}

	err := parseMsgFiles(input.inputDir, input.saveDir)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}
