package seishub

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"seismo"
	"strconv"
	"strings"
	"time"
)

var client = http.Client{Timeout: 60 * time.Second}

const defBaseAddr = "http://seishub.ru/pipermail/seismic-report/"

type Processor struct {
	BaseAddr string
	http.Client
}

func NewProcessor(baseAddr string, timeout time.Duration) *Processor {
	if baseAddr == "" {
		baseAddr = defBaseAddr
	}

	if timeout == 0 {
		timeout = 60 * time.Second
	}

	return &Processor{BaseAddr: baseAddr, Client: http.Client{Timeout: timeout}}
}

func (p *Processor) Watch(ctx context.Context, from time.Time, checkPeriod time.Duration) (<-chan seismo.Message, error) {
	ch := make(chan seismo.Message)
	go func() {
		//TO DO: this function watches the seishub, extracts new messages and send them into the channel
	}()
	return ch, nil
}

func ExtractMessages(ctx context.Context, url string) ([]*seismo.Message, error) {

	namesPage, err := GetMsgNamesPage(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("ExtractMessages: %v ", err)
	}

	names := parseMsgNames(namesPage)

	msgs := make([]*seismo.Message, 0, len(names))
	for _, n := range names {
		m, err := extractMsg(ctx, url, n)
		if err != nil {
			log.Printf("extract message error: %v url: %s, name: %s", err, url, n)
		} else {
			msgs = append(msgs, m)
		}
	}

	return msgs, nil
}

func GetMsgPages(ctx context.Context, url string) (map[string]string, error) {
	namesPage, err := GetMsgNamesPage(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("GetMsgPages: %v ", err)
	}

	names := parseMsgNames(namesPage)

	msgs := make(map[string]string, len(names))
	for _, n := range names {
		m, err := getStrMsg(ctx, url, n)
		if err != nil {
			log.Printf("GetMsgPages: get message page: %v url: %s, name: %s", err, url, n)
		}
		msgs[n] = m
	}

	return msgs, nil
}

// GetMsgNamesPage retrieves raw html-page containting the link names of messages
func GetMsgNamesPage(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("GetMsgNamesPage: \"%s\": %w", url, err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("GetMsgNamesPage: \"%s\": %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GetMsgNamesPage: \"%s\": %s", url, resp.Status)
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return "", fmt.Errorf("GetMsgNamesPage: copy response body \"%s\": %w", url, err)
	}

	return buf.String(), nil
}

func parseMsgNames(s string) []string {
	re := regexp.MustCompile(`\d+\.html`)
	return re.FindAllString(s, -1)
}

func extractMsg(ctx context.Context, dir string, name string) (m *seismo.Message, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("extractMsg: %w", err)
		}
	}()

	sm, err := getStrMsg(ctx, dir, name)
	if err != nil {
		return nil, err
	}

	m, err = parseMsg(sm)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func getStrMsg(ctx context.Context, dir string, name string) (string, error) {
	url, err := url.JoinPath(dir, name)
	if err != nil {
		return "", fmt.Errorf("getStrMsg: dir arg %q, name erg %q: %w", dir, name, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("getStrMsg: url %q: %w", url, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("getStrMsg: get %q: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("getStrMsg: get %q: %s", url, resp.Status)
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return "", fmt.Errorf("getStrMsg: copy response body %q: %w", url, err)
	}

	return buf.String(), nil
}

func parseMsg(msg string) (*seismo.Message, error) {
	var resMsg seismo.Message

	//Parse EventId (obligatiry)
	re := regexp.MustCompile(`EVENT PUBLIC ID:\s*\w+`)
	resMsg.EventId = strings.Trim(strings.TrimPrefix(re.FindString(msg), "EVENT PUBLIC ID:"), " \r\n")
	if resMsg.EventId == "" {
		return nil, fmt.Errorf("parseMsg: cannot parse EventId")
	}

	//Parse FocusTime (obligatory); Parse format like 2023.03.01 05:13:16.43
	re = regexp.MustCompile(`ВРЕМЯ В ОЧАГЕ \(UTC\):\s*[0-9-:. ]+`)
	fTimeStr := strings.Trim(strings.TrimPrefix(re.FindString(msg), "ВРЕМЯ В ОЧАГЕ (UTC):"), " \r\n")
	fTimeStr = strings.ReplaceAll(fTimeStr, "-", ".")
	fTime, err := time.Parse("2006.01.02 15:04:5", fTimeStr)
	if err != nil {
		return nil, fmt.Errorf("parseMsg: parse FocusTime: %w", err)
	}
	resMsg.FocusTime = fTime

	//Parse Latitude (obligatory)
	re = regexp.MustCompile(`ШИРОТА:\s*[0-9-.]+`)
	ltd, err := strconv.ParseFloat(strings.Trim(strings.TrimPrefix(re.FindString(msg), "ШИРОТА:"), " \r\n"), 64)
	if err != nil {
		return nil, fmt.Errorf("parseMsg: parse Latitude: %w", err)
	}
	resMsg.Latitude = ltd

	//Parse Longitude (obligatory)
	re = regexp.MustCompile(`ДОЛГОТА:\s*[0-9-.]+`)
	lng, err := strconv.ParseFloat(strings.Trim(strings.TrimPrefix(re.FindString(msg), "ДОЛГОТА:"), " \r\n"), 64)
	if err != nil {
		return nil, fmt.Errorf("parseMsg: parse Longitude: %w", err)
	}
	resMsg.Longitude = lng

	//Parse Magnitude (optional)
	re = regexp.MustCompile(`МАГНИТУДА:\s*[nan0-9.]+`)
	mgn, err := strconv.ParseFloat(strings.Trim(strings.TrimPrefix(re.FindString(msg), "МАГНИТУДА:"), " \r\n"), 64)
	if err != nil {
		return nil, fmt.Errorf("parseMsg: parse Magnitude: %w", err)
	}
	resMsg.Magnitude = mgn

	//Parse EventType (optional)
	re = regexp.MustCompile(`ТИП СОБЫТИЯ:\s*[A-Za-z ]+`)
	resMsg.EventType = strings.Trim(strings.TrimPrefix(re.FindString(msg), "ТИП СОБЫТИЯ:"), " \r\n")

	//Parse Quality (obligatory)
	re = regexp.MustCompile(`ОЦЕНКА КАЧЕСТВА РЕШЕНИЯ:\s*[А-Яа-я, ]+`)
	resMsg.Quality = strings.Trim(strings.TrimPrefix(re.FindString(msg), "ОЦЕНКА КАЧЕСТВА РЕШЕНИЯ:"), " \r\n")
	if resMsg.Quality == "" {
		return nil, fmt.Errorf("parseMsg: cannot parse Qualtity")
	}

	return &resMsg, nil
}
