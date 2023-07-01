package seishub

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"seismo/provider"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	//defBaseAddr constant defines the basic default address for SEISHUB
	defBaseAddr = "http://seishub.ru/pipermail/seismic-report/"

	//defParal constant defines max number of go routings each of them gets one
	//message from seishub
	defParal = 7

	//avgMonthMsgNum constant defines average number of seismic messages per month
	//on SEISHUB. This constant is used to create slices with proper capacity.
	avgMonthMsgNum = 200
)

// hubState interface for implementing the STATE DESIGN PATTERN in the Hub structure
type hubState interface {
	startWatch(ctx context.Context, from time.Time, checkPeriod time.Duration) (<-chan provider.Message, error)
	stateInfo() provider.WatcherStateInfo
}

// stoppedState implements a stopped Hub's behavior within the STATE PATTERN
type stoppedState struct {
	hub *Hub
}

func newStoppedState(h *Hub) *stoppedState {
	return &stoppedState{hub: h}
}

func (s *stoppedState) startWatch(ctx context.Context, from time.Time, checkPeriod time.Duration) (<-chan provider.Message, error) {
	from = from.UTC()
	now := time.Now().UTC()
	if from.After(now) {
		return nil, fmt.Errorf(`Watching cannot be started in the future (the "from" arg cannot after the start time)`)
	}
	h := s.hub
	h.setState(newRunState(s.hub))
	o := make(chan provider.Message)
	sn := make(chan int, 1)
	go h.getStartMsgNum(ctx, sn, from, checkPeriod)
	go h.watch(ctx, o, sn, from, checkPeriod)

	return o, nil
}

func (s *stoppedState) stateInfo() provider.WatcherStateInfo {
	return provider.Stopped
}

// runState implements a running Hub's behavior within the STATE PATTERN
type runState struct {
	hub *Hub
}

func newRunState(h *Hub) *runState {
	return &runState{hub: h}
}

func (r *runState) startWatch(ctx context.Context, from time.Time, checkPeriod time.Duration) (<-chan provider.Message, error) {
	return nil, provider.AlreadyRunErr{}
}

func (r *runState) stateInfo() provider.WatcherStateInfo {
	return provider.Run
}

// Hub provides getting seismic event messages
// (by parsing SEISHUB message pages in order to extract seismic event info)
// and also tracking (watching) the appearance of new messages.
// Hub embeds an http.Client.
type Hub struct {
	id       string
	BaseAddr string
	http.Client

	//state implements the State pattern
	state hubState
}

// NewHub returns a new SEISHUB Hub in the stopped state for a given basic address (baseAddr)
// with a specified timeout for the embedded http.Client. If the "baseAddr" arg is an empty string
// or the timeout is 0, default values will be used.
func NewHub(id string, baseAddr string, timeout time.Duration) *Hub {
	if baseAddr == "" {
		baseAddr = defBaseAddr
	}

	if timeout == 0 {
		timeout = 60 * time.Second
	}

	h := &Hub{BaseAddr: baseAddr, Client: http.Client{Timeout: timeout}}
	h.SetId(id)
	h.setState(newStoppedState(h))

	return h
}

func (h *Hub) SetId(id string) {
	h.id = id
}

func (h *Hub) GetId() string {
	return h.id
}

func (h *Hub) setState(s hubState) {
	h.state = s
}

func (h *Hub) StateInfo() provider.WatcherStateInfo {
	return h.state.stateInfo()
}

// StartWatch starts monitoring the appearance of new messages on SEISHUB
// and extracting such messages (message information).
// Returns a channel for getting new messages.
// Can start watching only in the current month or before.
// Watching can't be started in future months.
// Returns an error in such case.
func (h *Hub) StartWatch(ctx context.Context, from time.Time, checkPeriod time.Duration) (<-chan provider.Message, error) {
	o, err := h.state.startWatch(ctx, from, checkPeriod)
	return o, err
}

func (h *Hub) watch(ctx context.Context, o chan<- provider.Message, sn <-chan int, from time.Time, checkPeriod time.Duration) {
	defer close(o)
	msgNum, ok := <-sn //Wait for the start message number
	if !ok {
		log.Println("watch: Start msg num channel has been closed. Return.")
		return
	}
	wt := time.NewTicker(checkPeriod)
	defer wt.Stop()

	month := provider.MonthYear{Month: from.Month(), Year: from.Year()}
	for {
		select {
		case <-wt.C:
			msg, err := h.checkMsg(ctx, &msgNum, &month)
			if err != nil {
				log.Printf("watch: %v\n", err)

			} else if msg != nil {
				o <- *msg
			}
		case <-ctx.Done():
			log.Println("watch: Canceled")
		}
	}

}

func (h *Hub) checkMsg(ctx context.Context, msgNum *int, month *provider.MonthYear) (*provider.Message, error) {
	msgName := msgNumToName(*msgNum)
	l, err := url.JoinPath(h.BaseAddr, MonthYearPathSeg(month.Month, month.Year), msgName)
	if err != nil {
		return nil, fmt.Errorf("checkMsg: %w", err)
	}

	msg, err := h.getMsgByLink(ctx, l)
	if err == nil { //err is EQUAL nil !!! Getting is succeful
		*msgNum++
		return msg, nil
	}

	if !errors.As(err, &NotFoundErr{}) { //all errors except NotFoundErr
		return nil, err
	}

	//NotFound error: check next month
	nextMonth := month.AddMonth(1)
	l, err = url.JoinPath(h.BaseAddr, MonthYearPathSeg(nextMonth.Month, nextMonth.Year), msgName)
	if err != nil {
		return nil, fmt.Errorf("checkMsg: %w", err)
	}

	msg, err = h.getMsgByLink(ctx, l)
	if err == nil { //err is EQUAL nil !!! a message has been found in the next month
		*msgNum++
		*month = nextMonth //move to the next month
		return msg, nil
	}

	if !errors.As(err, &NotFoundErr{}) { //all errors except NotFoundErr
		return nil, err
	}

	//The message has not been found
	return nil, nil
}

func (h *Hub) getStartMsgNum(ctx context.Context, sn chan<- int, from time.Time, checkPeriod time.Duration) {
	m := provider.MonthYear{Month: from.Month(), Year: from.Year()}
	wt := time.NewTicker(checkPeriod)
	defer func() {
		wt.Stop()
		close(sn)
	}()

	for {
		select {
		case <-wt.C:
			msgs, err := h.Extract(ctx, m, m, 0)
			if err != nil {
				log.Printf("getStartMsgNum: %v", err)
				return
			}

			if len(msgs) > 0 {
				n, err := findStartMsgNum(msgs, from)
				if err != nil {
					log.Printf("getStartMsgNum: %v", err)
					return
				}
				sn <- n
				return
			}
		case <-ctx.Done():
			log.Println("getStartMsgNum: Canceled")
			return
		}
	}
}

func findStartMsgNum(msgs []*provider.Message, from time.Time) (int, error) {
	//Create a meta-slice ordered by a message number (parsed from the link)
	//and find the first message with focus time more than the "from" arg
	//This logic is neccesary because the seishub DOESN'T ENSURE that a message
	//about a later event has a higher number

	type meta struct {
		num       int
		focusTime time.Time
	}

	ind := make([]*meta, 0, len(msgs))
	for _, m := range msgs {
		n, err := parseMsgNum(m.Link)
		if err != nil {
			return 0, fmt.Errorf("findStartMsgNum: %w", err)
		}
		ind = append(ind, &meta{num: n, focusTime: m.FocusTime})
	}

	sort.Slice(ind, func(i, j int) bool {
		return ind[i].num < ind[j].num
	})

	for _, v := range ind {
		if v.focusTime.After(from) || v.focusTime.Equal(from) {
			return v.num, nil
		}
	}

	return ind[len(ind)-1].num, nil
}

func (h *Hub) getMsgByLink(ctx context.Context, link string) (*provider.Message, error) {
	m, err := GetMsg(ctx, link, &h.Client)
	if err != nil {
		return nil, fmt.Errorf("getMsgByLink: link %s, error: %w", link, err)
	}
	m.SourceId = h.id

	return m, nil
}

// Extract returns seismic messages extracted from SEISHUB.
func (h *Hub) Extract(ctx context.Context,
	from provider.MonthYear, to provider.MonthYear, paral int) ([]*provider.Message, error) {

	monthNum := to.Diff(from) + 1
	if monthNum <= 0 {
		return nil, fmt.Errorf(`Extract: the "from" arg cannot be more than the "to" arg`)
	}

	if paral <= 0 {
		paral = defParal
	}

	//Result slice of messages
	msgs := make([]*provider.Message, 0, avgMonthMsgNum*monthNum)
	links := make(chan string)

	var wg sync.WaitGroup
	wg.Add(paral)
	for i := 0; i < paral; i++ {
		go func() {
			defer func() {
				wg.Done()
			}()
			for l := range links {
				msg, err := h.getMsgByLink(ctx, l)
				if err != nil {
					log.Printf("Extract: link: %q error: %v", l, err)
				} else {
					msgs = append(msgs, msg)
				}
			}
		}()
	}

	for m := from; !m.After(to); m = m.AddMonth(1) {
		sg := MonthYearPathSeg(m.Month, m.Year)
		monthLink, err := url.JoinPath(h.BaseAddr, sg)
		if err != nil {
			return nil, fmt.Errorf("Extract: %v ", err)
		}

		names, err := GetMsgNames(ctx, monthLink, &h.Client)
		if err != nil && errors.As(err, &NotFoundErr{}) {
			log.Printf("Extract: %v", err)
			continue
		} else if err != nil {
			return nil, fmt.Errorf("Extract: %v", err)
		}

		for _, n := range names {
			lnk, err := url.JoinPath(monthLink, n)
			if err != nil {
				return nil, fmt.Errorf("Extract: %v", err)
			}
			links <- lnk
		}
	}
	close(links)
	wg.Wait()

	return msgs, nil
}

func parseMsgNum(s string) (int, error) {
	re := regexp.MustCompile(`[0-9]+[.]html`)
	n, err := strconv.Atoi(strings.TrimRight(re.FindString(s), ".html"))
	if err != nil {
		return 0, fmt.Errorf("parseMsgNum: cannot parse %q error: %w", s, err)
	}
	return n, nil
}

// parseMsgNumbers parses message numbers from message names
// and returns a sorted (from min to max) slice of ints
func parseMsgNumbers(ss []string) ([]int, error) {
	nums := make([]int, 0, len(ss))
	for _, s := range ss {
		n, err := parseMsgNum(s)
		if err != nil {
			return nil, fmt.Errorf("parseMsgNumbers: %w", err)
		}

		nums = append(nums, n)
	}

	sort.Slice(nums, func(i, j int) bool {
		return nums[i] < nums[j]
	})

	return nums, nil
}

// msgNumToName creates a message page name like "000234.html"
// It panics if n is less than 0 or has more than 6 digits
func msgNumToName(n int) string {
	if n < 0 {
		panic("msgNumToName: the n arg is less than 0")
	}
	s := strconv.Itoa(n)
	return strings.Repeat("0", 6-len(s)) + s + ".html"
}
