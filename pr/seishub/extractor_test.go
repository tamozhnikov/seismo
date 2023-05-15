package seishub

import (
	"context"
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

// func Test_ExtractMessages(t *testing.T) {
// 	input := "http://seishub.ru/pipermail/seismic-report/2023-March/"
// 	//input := "http://seishub.ru/pipermail/seismic-report/2022-March/"
// 	res, err := ExtractMessages(context.Background(), input)
// 	if err != nil {
// 		t.Errorf("\nExtractMessages: \n\t input: %s \n\t error: %v \n\t result count: %d", input, err, len(res))
// 	}
// }
