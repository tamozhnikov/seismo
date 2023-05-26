package seishub

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"seismo"
	"sync"
	"time"
)

const (
	//defBaseAddr constant defines the basic default address for SEISHUB
	defBaseAddr = "http://seishub.ru/pipermail/seismic-report/"

	//defParal constant defines max number of go routings each of them gets one
	//message from seishub
	defParal = 10

	//avgMonthMsgNum constant defines average number of seismic messages per month
	//on SEISHUB. This constant is used to create slices with proper capacity.
	avgMonthMsgNum = 200
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
func (e *Extractor) ExtractMessages(ctx context.Context, from seismo.MonthYear, to seismo.MonthYear, paral int) ([]*seismo.Message, error) {
	monthNum := to.Diff(from) + 1
	if monthNum <= 0 {
		return nil, fmt.Errorf(`ExtractMessages: the "from" arg cannot be more than the "to" arg`)
	}

	if paral <= 0 {
		paral = defParal
	}

	//Result slice of messages
	msgs := make([]*seismo.Message, 0, avgMonthMsgNum*monthNum)
	links := make(chan string)

	var wg sync.WaitGroup
	wg.Add(paral)
	for i := 0; i < paral; i++ {
		go func() {
			defer func() {
				wg.Done()
			}()
			for l := range links {
				msg, err := extractMsg(ctx, l)
				if err != nil {
					log.Printf("extract message error: %v url: %s", err, l)
				} else {
					msgs = append(msgs, msg)
				}
			}
		}()
	}

	for m := from; !m.After(to); m.AddMonth(1) {
		sg := MonthYearPathSeg(m.Month, m.Year)
		monthLink, err := url.JoinPath(e.BaseAddr, sg)
		if err != nil {
			return nil, fmt.Errorf("ExtractMessages: %v", err)
		}

		namesPage, err := GetMsgNamesPage(ctx, monthLink, nil)
		if err != nil {
			return nil, fmt.Errorf("ExtractMessages: %v ", err)
		}

		for _, n := range parseMsgNames(namesPage) {
			lnk, err := url.JoinPath(monthLink, n)
			if err != nil {
				return nil, fmt.Errorf("ExtractMessages: %v", err)
			}
			links <- lnk
		}
	}
	close(links)
	wg.Wait()

	return msgs, nil
}

func extractMsg(ctx context.Context, url string) (m *seismo.Message, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("extractMsg: %w", err)
		}
	}()

	sm, err := getMsgPage(ctx, url, nil)
	if err != nil {
		return nil, err
	}

	m, err = ParseMsg(sm)
	if err != nil {
		return nil, err
	}

	return m, nil
}
