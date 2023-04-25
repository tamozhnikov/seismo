package seishub

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"seismo"
	"strconv"
	"strings"
	"time"
)

const (
	//defBaseAddr constant defines the basic default address for SEISHUB
	defBaseAddr = "http://seishub.ru/pipermail/seismic-report/"

	//avgMonthMsgNum constant returns average number of seismic messages per month
	//on SEISHUB. This constant is used to create slices with proper capacity.
	avgMonthMsgNum = 300
)

// Extractor provides getting seismic event messages
// (by parsing SEISHUB message pages in order to extract seismic event info)
// and also tracking (watching) the appearance of new messages.
// Extractor embeds an http.Client.
type Extractor struct {
	BaseAddr string
	http.Client
}

// NewExtractor returns a new SEISHUB Extractor for a given basic address (baseAddr)
// with specified timeout for the embedded http.Client.
func NewExtractor(baseAddr string, timeout time.Duration) *Extractor {
	if baseAddr == "" {
		baseAddr = defBaseAddr
	}

	if timeout == 0 {
		timeout = 60 * time.Second
	}

	return &Extractor{BaseAddr: baseAddr, Client: http.Client{Timeout: timeout}}
}

// Watch starts monitoring the appearance of new messages on SEISHUB
// and extracting such messages (message information).
// Returns a channel for getting new messages.
func (e *Extractor) Watch(ctx context.Context, from time.Time, checkPeriod time.Duration) (<-chan seismo.Message, error) {
	ch := make(chan seismo.Message)
	go func() {
		//TO DO: this function watches the seishub, extracts new messages and send them into the channel
	}()
	return ch, nil
}

// ExtractMessages returns seismic messages extracted from SEISHUB.
func (e *Extractor) ExtractMessages(ctx context.Context, from seismo.MonthYear, to seismo.MonthYear) ([]*seismo.Message, error) {
	fromDate, toDate := from.Date(), to.Date()
	if fromDate.After(toDate) {
		return nil, fmt.Errorf("ExtractMessages: the from arg cannot be more than the to arg")
	}

	msgs := make([]*seismo.Message, 0, avgMonthMsgNum*(int(fromDate.Sub(toDate).Hours())/24+1))

	//iterate months
	for my := fromDate; !my.After(toDate); my = my.AddDate(0, 1, 0) {
		sg := MonthYearPathSeg(my.Month(), my.Year())
		url, err := url.JoinPath(e.BaseAddr, sg)
		if err != nil {
			return nil, fmt.Errorf("ExtractMessages: %v", err)
		}

		namesPage, err := GetMsgNamesPage(ctx, url, nil)
		if err != nil {
			return nil, fmt.Errorf("ExtractMessages: %v ", err)
		}

		names := parseMsgNames(namesPage)

		//iterate msg page link-names for current month
		for _, n := range names {
			m, err := extractMsg(ctx, url, n)
			if err != nil {
				log.Printf("extract message error: %v url: %s, name: %s", err, url, n)
			} else {
				msgs = append(msgs, m)
			}
		}
	}
	return msgs, nil
}

func extractMsg(ctx context.Context, dir string, name string) (m *seismo.Message, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("extractMsg: %w", err)
		}
	}()

	sm, err := getMsgPage(ctx, dir, name, nil)
	if err != nil {
		return nil, err
	}

	m, err = parseMsg(sm)
	if err != nil {
		return nil, err
	}

	return m, nil
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
