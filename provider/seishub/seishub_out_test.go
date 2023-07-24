//go:build outtest
// +build outtest

package seishub

import (
	"context"
	"os"
	"seismo/provider"
	"testing"
	"time"
)

func Test_getMsgPage(t *testing.T) {
	var input = struct {
		url string
	}{"http://seishub.ru/pipermail/seismic-report/2023-March/021128.html"}

	c, err := os.ReadFile("testdata/html/msg_asb2023eesfwx.html")
	if err != nil {
		panic(err)
	}

	want := string(c)
	res, err := GetMsgPage(context.Background(), input.url, nil)
	if err != nil {
		t.Errorf("\ngetMsgPage: \n\t error: %v", err)
	}

	if res != want {
		t.Errorf(("\ngetMsgPage \n\t result != want"))
	}
}

func Test_GetMsg(t *testing.T) {
	input := struct {
		url string
	}{
		"http://seishub.ru/pipermail/seismic-report/2023-March/021128.html",
	}

	want := provider.Message{EventId: "asb2023eesfwx"}
	want.FocusTime, _ = time.Parse("2006.01.02 03:04:05", "2023.03.01 05:13:16.43")
	want.Latitude = 54.71
	want.Longitude = 83.67
	want.Magnitude = 3.3
	want.Type = provider.QuarryBlast
	want.Quality = provider.Excellent
	want.Link = "http://seishub.ru/pipermail/seismic-report/2023-March/021128.html"

	res, err := GetMsg(context.Background(), input.url, nil)
	if err != nil {
		t.Errorf("Test_extractMsg: \n\t error: %v", err)
	}

	if res == nil || *res != want {
		t.Errorf("extractMsg: \n\t result != want")
	}
}
