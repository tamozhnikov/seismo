package seishub

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"seismo"
	"testing"
)

func Test_ExtractMessages(t *testing.T) {
	ext := NewHub("", 0)
	msgs, err := ext.Extract(context.Background(),
		seismo.MonthYear{Month: 6, Year: 2023}, seismo.MonthYear{Month: 7, Year: 2023}, 7)
	if err != nil {
		t.Fatalf("ExtractMessages: error: %v", err)
	}
	t.Log(len(msgs))
}

func Test_findStartMsgNum(t *testing.T) {
	inputDataDir := "testdata/json_msg/2022-February"

	inputFiles, err := ioutil.ReadDir(inputDataDir)
	if err != nil {
		t.Fatalf("Cannot read input data directory content list: %s", inputDataDir)
	}

	msgs := make([]*seismo.Message, 0, len(inputFiles))

	for _, f := range inputFiles {
		if f.IsDir() || f.Size() > maxInputSize {
			t.Logf("Skiping. \"%s\" is a folder or too big.\n", f.Name())
			continue
		}

		inputBuf, err := os.ReadFile(path.Join(inputDataDir, f.Name()))
		if err != nil {
			t.Fatalf("Cannot read \"%s\": %v\n", f.Name(), err)
		}

		var m seismo.Message
		if err = json.Unmarshal(inputBuf, &m); err != nil {
			t.Fatalf("\nCannot unmarshal \"%s\"; error: %v", f.Name()+".json", err)
		}

		msgs = append(msgs, &m)
	}

	// tests := []struct {
	// 	from time.Time
	// 	want int
	// }{
	// 	{time.Date(2022, 2, 1, 5, 56, 0, 0, time.UTC), 0},
	// }

	// for _, tst := range tests {
	// 	res := findStartMsgNum(msgs, tst.from)
	// 	if res != tst.want {
	// 		t.Errorf("findStartMsgNum: from: %v want: %d res:%d", tst.from, tst.want, res)
	// 	}
	// }
	//find00
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
