package seishub

import (
	"context"
	"fmt"
	"net/http"
	"seismo"
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
	// // fromDate, toDate := from.Date(), to.Date()
	// //  fromDate.After()
	// if from.After(to) {
	// 	return nil, fmt.Errorf(`ExtractMessages: the "from" arg cannot be more than the "to" arg`)
	// }

	// msgs := avgMonthMsgNum * (int(toDate.Sub(fromDate).Hours())/(24*28) + 1)

	// //msgs := make([]*seismo.Message, 0, msgCap)

	// //iterate months
	// for my := fromDate; !my.After(toDate); my = my.AddDate(0, 1, 0) {
	// 	sg := MonthYearPathSeg(my.Month(), my.Year())
	// 	url, err := url.JoinPath(e.BaseAddr, sg)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("ExtractMessages: %v", err)
	// 	}

	// 	namesPage, err := GetMsgNamesPage(ctx, url, nil)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("ExtractMessages: %v ", err)
	// 	}

	// 	names := parseMsgNames(namesPage)

	// 	//iterate msg page link-names for current month
	// 	for _, n := range names {
	// 		m, err := extractMsg(ctx, url, n)
	// 		if err != nil {
	// 			log.Printf("extract message error: %v url: %s, name: %s", err, url, n)
	// 		} else {
	// 			msgs = append(msgs, m)
	// 		}
	// 	}
	// }
	// return msgs, nil
	return nil, nil
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

	m, err = ParseMsg(sm)
	if err != nil {
		return nil, err
	}

	return m, nil
}
