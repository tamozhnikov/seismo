package seishub

import (
	"context"
	"os"
	"seismo"
	"testing"
	"time"
)

func Test_extractMsg(t *testing.T) {
	input := struct {
		dir  string
		name string
	}{
		"http://seishub.ru/pipermail/seismic-report/2023-March",
		"021128.html",
	}

	want := seismo.Message{EventId: "asb2023eesfwx"}
	want.FocusTime, _ = time.Parse("2006.01.02 03:04:05", "2023.03.01 05:13:16.43")
	want.Latitude = 54.71
	want.Longitude = 83.67
	want.Magnitude = 3.3
	want.EventType = "quarry blast"
	want.Quality = "наилучшее, обработано аналитиком"

	res, err := extractMsg(context.Background(), input.dir, input.name)
	if err != nil {
		t.Errorf("Test_extractMsg: \n\t error: %v", err)
	}

	if res == nil || *res != want {
		t.Errorf("extractMsg: \n\t result != want")
	}
}

func Test_parseMsg(t *testing.T) {
	input, err := os.ReadFile("testdata/msg_asb2023eesfwx.html")
	if err != nil {
		panic(err)
	}

	want := seismo.Message{EventId: "asb2023eesfwx"}
	want.FocusTime, _ = time.Parse("2006.01.02 03:04:05", "2023.03.01 05:13:16.43")
	want.Latitude = 54.71
	want.Longitude = 83.67
	want.Magnitude = 3.3
	want.EventType = "quarry blast"
	want.Quality = "наилучшее, обработано аналитиком"

	res, err := parseMsg(string(input))

	if err != nil {
		t.Errorf("parseMsg: \n\t error: %v", err)
	}

	if res == nil || *res != want {
		t.Errorf("parseMsg: \n\t result != want")
	}
}
