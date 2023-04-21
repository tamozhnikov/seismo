package seishub

import (
	"context"
	"fmt"
	"os"
	"seismo"
	"strings"
	"testing"
	"time"
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

	res := parseMsgNames(input)

	if len(res) != len(want) {
		t.Fail()
	}

	for i := 0; i < len(res); i++ {
		if want[i] != res[i] {
			t.Fail()
		}
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

func Test_getStrMsg(t *testing.T) {
	var input = struct {
		dir  string
		name string
	}{"http://seishub.ru/pipermail/seismic-report/2023-March", "021128.html"}

	c, err := os.ReadFile("testdata/msg_asb2023eesfwx.html")
	if err != nil {
		panic(err)
	}

	want := string(c)
	res, err := getStrMsg(context.Background(), input.dir, input.name)
	if err != nil {
		t.Errorf("\ngetMsg: \n\t error: %v", err)
	}

	if res != want {
		t.Errorf(("\ngetMsg \n\t result != want"))
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

func Test_ExtractMessages(t *testing.T) {
	input := "http://seishub.ru/pipermail/seismic-report/2023-March/"
	//input := "http://seishub.ru/pipermail/seismic-report/2022-March/"
	res, err := ExtractMessages(context.Background(), input)
	if err != nil {
		t.Errorf("\nExtractMessages: \n\t input: %s \n\t error: %v \n\t result count: %d", input, err, len(res))
	}
}

// func Test_monthYearPathSeg(t *testing.T) {
// 	var tests = []struct {
// 		input sd.MonthYear
// 		want  string
// 	}{
// 		{sd.MonthYear{1, 2022}, "2022-January"},
// 		{sd.MonthYear{2, 2022}, "2022-February"},
// 		{sd.MonthYear{3, 2022}, "2022-March"},
// 		{sd.MonthYear{4, 2022}, "2022-April"},
// 		{sd.MonthYear{5, 2022}, "2022-May"},
// 		{sd.MonthYear{6, 2022}, "2022-June"},
// 		{sd.MonthYear{7, 2022}, "2022-July"},
// 		{sd.MonthYear{8, 2022}, "2022-August"},
// 		{sd.MonthYear{9, 2022}, "2022-September"},
// 		{sd.MonthYear{10, 2022}, "2022-October"},
// 		{sd.MonthYear{11, 2022}, "2022-November"},
// 		{sd.MonthYear{12, 2022}, "2022-December"},
// 	}

// 	for _, test := range tests {
// 		if res := monthYearPathSeg(test.input); res != test.want {
// 			t.Errorf("input: %s result: %s", test.input.String(), res)
// 		}
// 	}
// }
