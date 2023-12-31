package seishub

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"seismo/provider"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const (
	maxInputSize = 10 * 1024
)

func Test_ParseMsgNames(t *testing.T) {
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
	buf.WriteString(fmt.Sprintf("\n Test_ParseMsgNames: \n\t want(%d) & result(%d):\n", len(want), len(want)))
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

func Test_ParseMsg(t *testing.T) {
	inputDataDir := "testdata/html/2022-February"
	wantDataDir := "testdata/json_msg/2022-February"

	inputFiles, err := os.ReadDir(inputDataDir)
	if err != nil {
		t.Fatalf("Test_ParseMsg: Cannot read input data directory content list: %s", inputDataDir)
	}

	for _, f := range inputFiles {
		inf, err := f.Info()
		if err != nil {
			t.Logf("Test_ParseMsg: Skiping. Cannot read info for %q\n", f.Name())
			continue
		}

		if f.IsDir() || inf.Size() > maxInputSize {
			t.Logf("Test_ParseMsg: Skiping. \"%s\" is a folder or too big.\n", f.Name())
			continue
		}

		inputBuf, err := os.ReadFile(path.Join(inputDataDir, f.Name()))
		if err != nil {
			t.Fatalf("Test_ParseMsg: Cannot read \"%s\": %v\n", f.Name(), err)
		}

		wantBuf, err := os.ReadFile(path.Join(wantDataDir, f.Name()+".json"))
		if err != nil {
			t.Fatalf("Test_ParseMsg: Cannot read \"%s\": %v\n", f.Name()+".json", err)
		}

		resultMsg, err := ParseMsg(string(inputBuf))
		if err != nil {
			t.Errorf("\nTest_ParseMsg: Cannot parse \"%s\": %v\n", f.Name(), err)
		}

		var wantMsg provider.Message
		if err = json.Unmarshal(wantBuf, &wantMsg); err != nil {
			t.Fatalf("\nTest_ParseMsg: Cannot unmarshal \"%s\"; error: %v", f.Name()+".json", err)
		}

		if *resultMsg != wantMsg {
			t.Errorf("\nTest_ParseMsg: \twant: %v\n\t result: %v\n", wantMsg, *resultMsg)
		}
	}
}

func Test_monthYearPathSeg(t *testing.T) {
	var tests = []struct {
		input provider.MonthYear
		want  string
	}{
		{provider.MonthYear{Month: 1, Year: 2022}, "2022-January"},
		{provider.MonthYear{Month: 2, Year: 2022}, "2022-February"},
		{provider.MonthYear{Month: 3, Year: 2022}, "2022-March"},
		{provider.MonthYear{Month: 4, Year: 2022}, "2022-April"},
		{provider.MonthYear{Month: 5, Year: 2022}, "2022-May"},
		{provider.MonthYear{Month: 6, Year: 2022}, "2022-June"},
		{provider.MonthYear{Month: 7, Year: 2022}, "2022-July"},
		{provider.MonthYear{Month: 8, Year: 2022}, "2022-August"},
		{provider.MonthYear{Month: 9, Year: 2022}, "2022-September"},
		{provider.MonthYear{Month: 10, Year: 2022}, "2022-October"},
		{provider.MonthYear{Month: 11, Year: 2022}, "2022-November"},
		{provider.MonthYear{Month: 12, Year: 2022}, "2022-December"},
	}

	for _, test := range tests {
		if res := MonthYearPathSeg(test.input.Month, test.input.Year); res != test.want {
			t.Errorf("input: %s result: %s", test.input.String(), res)
		}
	}
}
