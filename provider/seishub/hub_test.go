package seishub

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"seismo/provider"
	"testing"
	"time"
)

func Test_ExtractMessages(t *testing.T) {
	c := provider.WatcherConfig{Id: "seishub", Timeout: 120, CheckPeriod: 2}
	ext, err := NewHub(c)
	if err != nil {
		t.Fatalf("ExtractMessages: error: %v", err)
	}
	msgs, err := ext.Extract(context.Background(),
		provider.MonthYear{Month: 6, Year: 2023}, provider.MonthYear{Month: 7, Year: 2023}, 7)
	if err != nil {
		t.Fatalf("ExtractMessages: error: %v", err)
	}
	t.Log(len(msgs))
}

func Test_findStartMsgNum(t *testing.T) {
	inputDataDir := "testdata/json_msg/2022-February"

	inputFiles, err := os.ReadDir(inputDataDir)
	if err != nil {
		t.Fatalf("Cannot read input data directory content list: %s", inputDataDir)
	}

	msgs := make([]*provider.Message, 0, len(inputFiles))

	for _, f := range inputFiles {
		inf, err := f.Info()
		if err != nil {
			t.Logf("Skiping. Cannot read info for %q\n", f.Name())
			continue
		}

		if f.IsDir() || inf.Size() > maxInputSize {
			t.Logf("Skiping. %q is a folder or too big.\n", f.Name())
			continue
		}

		inputBuf, err := os.ReadFile(path.Join(inputDataDir, f.Name()))
		if err != nil {
			t.Fatalf("Cannot read \"%s\": %v\n", f.Name(), err)
		}

		var m provider.Message
		if err = json.Unmarshal(inputBuf, &m); err != nil {
			t.Fatalf("\nCannot unmarshal \"%s\"; error: %v", f.Name()+".json", err)
		}
		m.Link = f.Name()

		msgs = append(msgs, &m)
	}

	tests := []struct {
		from time.Time
		want int
	}{
		{time.Date(2022, 2, 1, 5, 55, 10, 0, time.UTC), 17538},
		{time.Date(2022, 2, 1, 5, 55, 12, 0, time.UTC), 17538},
		{time.Date(2022, 2, 1, 5, 56, 0, 0, time.UTC), 17541},
		{time.Date(2022, 2, 28, 7, 23, 20, 0, time.UTC), 17791},
		{time.Date(2022, 2, 28, 12, 00, 0, 0, time.UTC), 17801},
	}

	for _, tst := range tests {
		res, err := findStartMsgNum(msgs, tst.from)
		if err != nil {
			t.Errorf("fincStartMsgNum: error: %v", err)
		}
		if res != tst.want {
			t.Errorf("findStartMsgNum: from: %v want: %d res:%d", tst.from, tst.want, res)
		}
	}

}

func Test_parseMsgNum(t *testing.T) {
	tests := []struct {
		s    string
		want int
	}{
		{"https://seishub.ru/pipermail/seismic-report/2023-March/021128.html", 21128},
		{"2023-March/021128.html", 21128},
		{"2023-March/021128.htmlrjenrjek", 21128},
		{"  021128.html   ", 21128},
		{"2023-March/021128", 0},
		{"2023-March/021128.rjenrjek", 0},
	}

	for _, test := range tests {
		res, err := parseMsgNum(test.s)

		if res != test.want {
			t.Errorf("parseMsgNum: s: %s, want: %d, res: %d, error: %v", test.s, test.want, res, err)
		} else if err != nil {
			t.Logf("parseMsgNum: s: %s, want: %d, res: %d, error: %v", test.s, test.want, res, err)
		}
	}
}

func Test_msgNumToName(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "000000.html"},
		{1, "000001.html"},
		{23423, "023423.html"},
	}

	for _, test := range tests {
		res := msgNumToName(test.n)
		if res != test.want {
			t.Errorf("n: %d res: %s want: %s", test.n, res, test.want)
		}
	}
}
