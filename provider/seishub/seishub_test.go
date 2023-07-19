package seishub

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"seismo/provider"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

const (
	maxInputSize = 10 * 1024
)

func Test_parseMsgNames(t *testing.T) {
	input := `<UL>
	<!--0 01677647756.21127- -->
	<LI><A HREF="021127.html">[Seismic-Report] ОПЕРАТИВНОЕ СООБЩЕНИЕ О СЕЙСМИЧЕСКОМ СОБЫТИИ (asb2023eescua)</A><A NAME="21127">&nbsp;</A><I>Система Автоматической Обработки</I>

	<!--0 01677648264.21128- -->
	<LI><A HREF="021128.html">[Seismic-Report] ОПЕРАТИВНОЕ СООБЩЕНИЕ О СЕЙСМИЧЕСКОМ СОБЫТИИ (asb2023eesfwx)</A><A NAME="21128">&nbsp;</A><I>Seismic Reporting Service</I>

	<!--0 01677648661.21129- -->
	<LI><A HREF="021129.html">[Seismic-Report] ОПЕРАТИВНОЕ СООБЩЕНИЕ О СЕЙСМИЧЕСКОМ СОБЫТИИ (asb2023eestol)</A><A NAME="21129">&nbsp;</A><I>Система Автоматической Обработки</I>

	<UL>
	<!--1 01677648661.21129-01677649445.21130- -->
	<LI><A HREF="021130.html">[Seismic-Report] ОПЕРАТИВНОЕ СООБЩЕНИЕ О СЕЙСМИЧЕСКОМ СОБЫТИИ (asb2023eestol)</A><A NAME="21130">&nbsp;</A><I>Seismic Reporting Service</I>

	</UL>
	<!--0 01677649452.21131- -->
	<LI><A HREF="021131.html">[Seismic-Report] ОПЕРАТИВНОЕ СООБЩЕНИЕ О СЕЙСМИЧЕСКОМ СОБЫТИИ (asb2023eeszrv)</A><A NAME="21131">&nbsp;</A><I>Система Автоматической Обработки</I>

	<!--0 01677651486.21132- -->
	<LI><A HREF="021132.html">[Seismic-Report] ОПЕРАТИВНОЕ СООБЩЕНИЕ О СЕЙСМИЧЕСКОМ СОБЫТИИ (asb2023eeufwa)</A><A NAME="21132">&nbsp;</A><I>Система Автоматической Обработки</I>

	<!--0 01677653722.21133- -->
	<LI><A HREF="021133.html">[Seismic-Report] ОПЕРАТИВНОЕ СООБЩЕНИЕ О СЕЙСМИЧЕСКОМ СОБЫТИИ (asb2023eeuooq)</A><A NAME="21133">&nbsp;</A><I>Seismic Reporting Service</I>

	<!--0 01677661237.21134- -->
	<LI><A HREF="021134.html">[Seismic-Report] ОПЕРАТИВНОЕ СООБЩЕНИЕ О СЕЙСМИЧЕСКОМ СОБЫТИИ (asb2023eezneb)</A><A NAME="21134">&nbsp;</A><I>Система Автоматической Обработки</I>

	<UL>
	<!--1 01677661237.21134-01677661736.21135- -->
	<LI><A HREF="021135.html">[Seismic-Report] ОПЕРАТИВНОЕ СООБЩЕНИЕ О СЕЙСМИЧЕСКОМ СОБЫТИИ (asb2023eezneb)</A><A NAME="21135">&nbsp;</A><I>Система Автоматической Обработки</I>

	<!--1 01677661237.21134-01677661915.21136- -->
	<LI><A HREF="021136.html">[Seismic-Report] ОПЕРАТИВНОЕ СООБЩЕНИЕ О СЕЙСМИЧЕСКОМ СОБЫТИИ (asb2023eezneb)</A><A NAME="21136">&nbsp;</A><I>Система Автоматической Обработки</I>

	</UL>
	<!--0 01677712513.21137- -->
	<LI><A HREF="021137.html">[Seismic-Report] ОПЕРАТИВНОЕ СООБЩЕНИЕ О СЕЙСМИЧЕСКОМ СОБЫТИИ (asb2023egbtsg)</A><A NAME="21137">&nbsp;</A><I>Система Автоматической Обработки</I>

	<!--0 01677732835.21138- -->
	<LI><A HREF="021138.html">[Seismic-Report] ОПЕРАТИВНОЕ СООБЩЕНИЕ О СЕЙСМИЧЕСКОМ СОБЫТИИ (asb2023egmwpr)</A><A NAME="21138">&nbsp;</A><I>Система Автоматической Обработки</I>
	</UL>`

	want := []string{
		"021127.html",
		"021128.html",
		"021129.html",
		"021130.html",
		"021131.html",
		"021132.html",
		"021133.html",
		"021134.html",
		"021135.html",
		"021136.html",
		"021137.html",
		"021138.html",
	}

	res := ParseMsgNames(input)

	if len(res) != len(want) {
		t.Fail()
	}

	if !cmp.Equal(res, want) {
		t.Fail()
	}

	buf := new(strings.Builder)
	buf.WriteString(fmt.Sprintf("\ngetMsgNames: \n\t want(%d) & result(%d):\n", len(want), len(want)))
	buf.WriteString(sideBySide(want, res))
	t.Logf(buf.String())
}

func sideBySide(left []string, right []string) string {
	maxLen := max(len(left), len(right))

	buf := new(strings.Builder)
	buf.WriteString("\n")
	for i := 0; i < maxLen; i++ {
		if i < len(left) {
			buf.WriteString(left[i])
		} else {
			buf.WriteString("ABSENT")
		}

		buf.WriteString("\t\t\t\t")

		if i < len(right) {
			buf.WriteString(right[i])
		} else {
			buf.WriteString("ABSENT")
		}

		buf.WriteString("\n")
	}

	return buf.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

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

func Test_ParseMsg(t *testing.T) {
	inputDataDir := "testdata/html/2022-February"
	wantDataDir := "testdata/json_msg/2022-February"

	inputFiles, err := os.ReadDir(inputDataDir)
	if err != nil {
		t.Fatalf("Cannot read input data directory content list: %s", inputDataDir)
	}

	for _, f := range inputFiles {
		inf, err := f.Info()
		if err != nil {
			t.Logf("Skiping. Cannot read info for %q\n", f.Name())
			continue
		}

		if f.IsDir() || inf.Size() > maxInputSize {
			t.Logf("Skiping. \"%s\" is a folder or too big.\n", f.Name())
			continue
		}

		inputBuf, err := os.ReadFile(path.Join(inputDataDir, f.Name()))
		if err != nil {
			t.Fatalf("Cannot read \"%s\": %v\n", f.Name(), err)
		}

		wantBuf, err := os.ReadFile(path.Join(wantDataDir, f.Name()+".json"))
		if err != nil {
			t.Fatalf("Cannot read \"%s\": %v\n", f.Name()+".json", err)
		}

		resultMsg, err := ParseMsg(string(inputBuf))
		if err != nil {
			t.Errorf("\nCannot parse \"%s\": %v\n", f.Name(), err)
		}

		var wantMsg provider.Message
		if err = json.Unmarshal(wantBuf, &wantMsg); err != nil {
			t.Fatalf("\nCannot unmarshal \"%s\"; error: %v", f.Name()+".json", err)
		}

		if *resultMsg != wantMsg {
			t.Errorf("\nParseMsg: \twant: %v\n\t result: %v\n", wantMsg, *resultMsg)
		}
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

func Test_monthYearPathSeg(t *testing.T) {
	var tests = []struct {
		input provider.MonthYear
		want  string
	}{
		{provider.MonthYear{1, 2022}, "2022-January"},
		{provider.MonthYear{2, 2022}, "2022-February"},
		{provider.MonthYear{3, 2022}, "2022-March"},
		{provider.MonthYear{4, 2022}, "2022-April"},
		{provider.MonthYear{5, 2022}, "2022-May"},
		{provider.MonthYear{6, 2022}, "2022-June"},
		{provider.MonthYear{7, 2022}, "2022-July"},
		{provider.MonthYear{8, 2022}, "2022-August"},
		{provider.MonthYear{9, 2022}, "2022-September"},
		{provider.MonthYear{10, 2022}, "2022-October"},
		{provider.MonthYear{11, 2022}, "2022-November"},
		{provider.MonthYear{12, 2022}, "2022-December"},
	}

	for _, test := range tests {
		if res := MonthYearPathSeg(test.input.Month, test.input.Year); res != test.want {
			t.Errorf("input: %s result: %s", test.input.String(), res)
		}
	}
}
