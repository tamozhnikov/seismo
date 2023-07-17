// Package seismo/provider/seishub provides tools for getting information about
// seismic activity from the SEISHUB source (АСФ ФИЦ ЕГС РАН).
// For example, getting seismic event messages and
// lists of messages, parsing raw html messages and so on.
//
// Package seishub depends on the structure and content rules of SEISHUB's internet site.
//
// Each SEISHUB's message list contains the "link names" of all messages
// for the month. E.g., message list page can be addressed by
// "http://seishub.ru/pipermail/seismic-report/2022-April/" and may be
// considered as a folder (directory) of messages.
//
// "Link names" of messages are presented as "<0-Left-Augmented 6-position Number>.html",
// like "002364.html". To create the full link to the message, its name schould be joined
// to the address of its folder (message list).
//
// "Message number" is a number contained by a message name.
package seishub

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"seismo/provider"
	"strconv"
	"strings"
	"time"
)

// defClient is a package-level default http client, that can be
// used by package functions, having no specified client(s).
// The Timeout value is 60 sec.
var defClient = http.Client{Timeout: 60 * time.Second}

// NotFoundErr indicates that resource was not found.
// Can be useful to represent a 404 status as an error type.
type NotFoundErr struct {
	link string
}

func (e NotFoundErr) Error() string {
	return fmt.Sprintf("Not found %s", e.link)
}

// GetMsgPages returns a map of message pages (html code), where the key is
// a name of a message and nil.
// If the returned error is not nil, the returned map is nil.
//
// The "dir" parameter represents address of a message
// list page, e.g., "http://seishub.ru/pipermail/seismic-report/2022-April/".
func GetMsgPages(ctx context.Context, dir string) (map[string]string, error) {
	namesPage, err := GetMsgNamesPage(ctx, dir, nil)
	if err != nil {
		return nil, fmt.Errorf("GetMsgPages: %v ", err)
	}

	names := ParseMsgNames(namesPage)

	msgs := make(map[string]string, len(names))
	for _, n := range names {
		link, err := url.JoinPath(dir, n)
		if err != nil {
			log.Printf("GetMsgPage: dir %q, name %q: %V", dir, n, err)
		}
		m, err := GetMsgPage(ctx, link, nil)
		if err != nil {
			log.Printf("GetMsgPages: get message page: %v url: %s, name: %s", err, link, n)
		}
		msgs[n] = m
	}

	return msgs, nil
}

// GetMsgNamesPage returns an addressed by "dir" html page containting message names and an error.
// If the returned error is not nil, the returned string is empty.
//
// If the "cl" parameter is nil, the function uses the default package-level http client.
func GetMsgNamesPage(ctx context.Context, dir string, cl *http.Client) (string, error) {
	if cl == nil {
		cl = &defClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, dir, nil)
	if err != nil {
		return "", fmt.Errorf("GetMsgNamesPage: dir: %q error: %w", dir, err)
	}
	resp, err := cl.Do(req)
	if err != nil {
		return "", fmt.Errorf("GetMsgNamesPage: dir: %q error: %w", dir, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GetMsgNamesPage: error: %w", NotFoundErr{link: dir})
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return "", fmt.Errorf("GetMsgNamesPage: copy response body: dir: \"%s\" error: %w", dir, err)
	}

	return buf.String(), nil
}

// MonthYearPathSeg returns a string in "2022-April" format.
func MonthYearPathSeg(m time.Month, y int) string {
	return fmt.Sprintf("%d-%s", y, m.String())
}

// ParseMsgNames finds all message names on message list page (in its html code passed in s).
func ParseMsgNames(s string) []string {
	re := regexp.MustCompile(`\d+\.html`)
	return re.FindAllString(s, -1)
}

// GetMsgNames returns a slice of message names found on an addressed by "dir" page and an error.
// If the returned error is not nil, the returned slice is nil.
//
// If the "cl" parameter is nil, the function uses the default package-level http client.
func GetMsgNames(ctx context.Context, dir string, cl *http.Client) ([]string, error) {
	namesPage, err := GetMsgNamesPage(ctx, dir, cl)
	if err != nil {
		return nil, fmt.Errorf("GetMsgNames: %w ", err)
	}

	return ParseMsgNames(namesPage), nil
}

// GetMsgPage returns a message html page addressed by link and error.
// If the returned error is not nil, the returned string is empty.
//
// If the "cl" parameter is nil, the function uses the default package-level http client.
func GetMsgPage(ctx context.Context, link string, cl *http.Client) (string, error) {
	if cl == nil {
		cl = &defClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return "", fmt.Errorf("getMsgPage: link: %q error: %w", link, err)
	}

	resp, err := cl.Do(req)
	if err != nil {
		return "", fmt.Errorf("getMsgPage: link: %q error: %w", link, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("getMsgPage: error: %w", NotFoundErr{link: link})
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return "", fmt.Errorf("getMsgPage: copy response body %q: %w", link, err)
	}

	return buf.String(), nil
}

// ParseMsg returns a pointer to a seismic event message extracted from msg and an error.
// If the returned error is not nil, the returned message pointer is nil.
func ParseMsg(msg string) (*provider.Message, error) {
	var resMsg provider.Message

	//Parse EventId
	re := regexp.MustCompile(`EVENT PUBLIC ID:\s*\w+`)
	resMsg.EventId = strings.Trim(strings.TrimPrefix(re.FindString(msg), "EVENT PUBLIC ID:"), " \r\n")
	if resMsg.EventId == "" {
		return nil, fmt.Errorf("parseMsg: cannot parse EventId")
	}

	//Parse FocusTime; Parse format like 2023.03.01 05:13:16.43
	re = regexp.MustCompile(`ВРЕМЯ В ОЧАГЕ \(UTC\):\s*[0-9-:. ]+`)
	fTimeStr := strings.Trim(strings.TrimPrefix(re.FindString(msg), "ВРЕМЯ В ОЧАГЕ (UTC):"), " \r\n")
	fTimeStr = strings.ReplaceAll(fTimeStr, "-", ".")
	fTime, err := time.Parse("2006.01.02 15:04:5", fTimeStr)
	if err != nil {
		return nil, fmt.Errorf("parseMsg: parse FocusTime: %w", err)
	}
	resMsg.FocusTime = fTime

	//Parse Latitude
	re = regexp.MustCompile(`ШИРОТА:\s*[0-9-.]+`)
	ltd, err := strconv.ParseFloat(strings.Trim(strings.TrimPrefix(re.FindString(msg), "ШИРОТА:"), " \r\n"), 64)
	if err != nil {
		return nil, fmt.Errorf("parseMsg: parse Latitude: %w", err)
	}
	resMsg.Latitude = ltd

	//Parse Longitude
	re = regexp.MustCompile(`ДОЛГОТА:\s*[0-9-.]+`)
	lng, err := strconv.ParseFloat(strings.Trim(strings.TrimPrefix(re.FindString(msg), "ДОЛГОТА:"), " \r\n"), 64)
	if err != nil {
		return nil, fmt.Errorf("parseMsg: parse Longitude: %w", err)
	}
	resMsg.Longitude = lng

	//Parse Magnitude
	re = regexp.MustCompile(`МАГНИТУДА:\s*[0-9.]+`)
	valStr := strings.Trim(strings.TrimPrefix(re.FindString(msg), "МАГНИТУДА:"), " \r\n")
	if valStr != "" {
		mgn, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			return nil, fmt.Errorf("parseMsg: parse Magnitude: %w", err)
		}
		resMsg.Magnitude = mgn
	}

	//Parse EventType
	re = regexp.MustCompile(`ТИП СОБЫТИЯ:\s*[A-Za-z ]+`)
	resMsg.Type = defineEventType(strings.Trim(strings.TrimPrefix(re.FindString(msg), "ТИП СОБЫТИЯ:"), " \r\n"))

	//Parse Quality
	re = regexp.MustCompile(`ОЦЕНКА КАЧЕСТВА РЕШЕНИЯ:\s*[А-Яа-я, ]+`)
	resMsg.Quality = defineEventQuality(strings.Trim(strings.TrimPrefix(re.FindString(msg), "ОЦЕНКА КАЧЕСТВА РЕШЕНИЯ:"), " \r\n"))

	return &resMsg, nil
}

// GetMsg returns an event message, the html page of which addressed by "link" and an error.
// If the return error is not nil, the returned message is nil.
//
// If the "cl" parameter is nil, the function uses the default package-level http client.
func GetMsg(ctx context.Context, link string, cl *http.Client) (m *provider.Message, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("extractMsg: %w", err)
		}
	}()

	sm, err := GetMsgPage(ctx, link, cl)
	if err != nil {
		return nil, err
	}

	m, err = ParseMsg(sm)
	if err != nil {
		return nil, err
	}
	m.Link = link

	return m, nil
}

// defineEventType converts a passed string value to the corresponding EventType value.
func defineEventType(s string) provider.EventType {
	switch strings.ToLower(s) {
	case "quarry blast":
		return provider.QuarryBlast
	case "earthquake":
		return provider.EarthQuake
	default:
		return provider.UnknownType
	}
}

// defineEventQuality converts a passed string value to the corresponding EventQuality value.
func defineEventQuality(s string) provider.EventQuality {
	switch strings.ToLower(s) {
	case "наилучшее, обработано аналитиком":
		return provider.Excellent
	case "предварительная оценка":
		return provider.Preliminary
	case "хорошо":
		return provider.Good
	default:
		return provider.UnknownQuality
	}
}
