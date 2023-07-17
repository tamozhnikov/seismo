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
	//DefConnStr constant defines the basic default address for SEISHUB
	DefConnStr = "http://seishub.ru/pipermail/seismic-report/"

	//defParal constant defines max number of go routings each of them fetched
	//messages from seishub
	defParal = 7

	//avgMonthMsgNum constant defines average number of seismic messages per month
	//on SEISHUB. This constant is used to create slices with proper capacity.
	avgMonthMsgNum = 200
)

// hubState is implemented to provide a specific behavior within THE STATE PATTERN.
type hubState interface {
	startWatch(ctx context.Context, from time.Time) (<-chan provider.Message, error)
	stateInfo() provider.WatcherStateInfo
}

// stoppedState implements a stopped Hub's behavior within THE STATE PATTERN
type stoppedState struct {
	hub *Hub
}

func newStoppedState(h *Hub) *stoppedState {
	return &stoppedState{hub: h}
}

// startWatch implements the behaivor of Hub.StartWatch in the "stopped" state,
// i.e starts monitoring the appearance of new messages on SEISHUB
// and extracting such messages (message information).
//
// The method returns a channel for fetching messages (message channel).
// If the returned error is not nil, the returned channel is nil.
//
// Can start watching only in the current month or before.
// Watching can't be started in future months. Returns an error in such case.
func (s *stoppedState) startWatch(ctx context.Context, from time.Time) (<-chan provider.Message, error) {
	// To start watching it is necessary to get the number of the start message,
	// that is, the message that will be fetched first. Therefore, the function
	// starts 2 go-routines: the first go-routine gets the start message number
	// and sends it to the second go-routine through a channel, the second go-routine
	// is waiting for this number and starts watching immediately after receiving the number.
	//
	// The function is designed to return a message channel to the user as soon as possible, and
	// since the search for the start message number may take some time, a separate go-routine is
	// used for this purpose.

	from = from.UTC()
	now := time.Now().UTC()
	if from.After(now) {
		return nil, fmt.Errorf(`watching cannot be started in the future (the "from" arg cannot be after the start time)`)
	}
	h := s.hub
	h.setState(newRunState(s.hub))
	o := make(chan provider.Message) //output channel for fetched messages
	sn := make(chan int, 1)          //channel to transfer the start message number from getStartMsgNum() to watch()
	go h.getStartMsgNum(ctx, sn, from, time.Duration(h.config.CheckPeriod)*time.Second)
	go h.watch(ctx, o, sn, from, time.Duration(h.config.CheckPeriod)*time.Second)

	return o, nil
}

func (s *stoppedState) stateInfo() provider.WatcherStateInfo {
	return provider.Stopped
}

// runState implements a running Hub's behavior within THE STATE PATTERN
type runState struct {
	hub *Hub
}

func newRunState(h *Hub) *runState {
	return &runState{hub: h}
}

func (r *runState) startWatch(ctx context.Context, from time.Time) (<-chan provider.Message, error) {
	return nil, provider.AlreadyRunErr{}
}

func (r *runState) stateInfo() provider.WatcherStateInfo {
	return provider.Run
}

// Hub provides getting SEISHUB's seismic event messages
// and also tracking (watching) the appearance of new messages.
//
// Hub embeds an http.Client.
type Hub struct {
	config provider.WatcherConfig
	http.Client

	//state implements the State pattern
	state hubState
}

// NewHub returns a pointer to a new seishub.Hub in the stopped state
// configured by "conf" values and an error.
//
// If the returned error is not nil, the returned pointer is nil.
func NewHub(conf provider.WatcherConfig) (*Hub, error) {
	if conf.CheckPeriod < 1 {
		return nil, fmt.Errorf("NewHub: checkperiod cannot be less than 1 (second)")
	}

	if conf.Timeout < 1 {
		return nil, fmt.Errorf("NewHub: timeout cannot be less than 1 (second)")
	}

	if conf.ConnStr == "" {
		conf.ConnStr = DefConnStr
	}

	h := &Hub{config: conf, Client: http.Client{Timeout: time.Duration(conf.Timeout) * time.Second}}

	h.setState(newStoppedState(h))

	return h, nil
}

func (h *Hub) GetConfig() provider.WatcherConfig {
	return h.config
}

func (h *Hub) setState(s hubState) {
	h.state = s
}

func (h *Hub) StateInfo() provider.WatcherStateInfo {
	return h.state.stateInfo()
}

// StartWatch starts monitoring the appearance of new messages on SEISHUB
// and extracting such messages (message information).
//
// The method returns a channel for fetching messages. If the returned error is not nil, the returned
// channel is nil.
//
// Can start watching only in the current month or before.
// Watching can't be started in future months.Returns an error in such case.
func (h *Hub) StartWatch(ctx context.Context, from time.Time) (<-chan provider.Message, error) {
	o, err := h.state.startWatch(ctx, from)
	return o, err
}

// watch waits the start message number from the "sn" channel, than the function checks for new messages
// with a frequency of "checkPeriod" and send into the o channel.
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

// checkMsg checks for a message with the message number "msgNum" in "month"
// and increments "msgNum" if the message has been found. Otherwise the function
// checks for a message with this number in the next month and increments
// "msgNum" and "month" if the message has been found.
//
// The function returns a pointer to provider.Message and error. If error is not nil,
// the message pointer is nil.
// If the message is not found, the message pointer is nil.
func (h *Hub) checkMsg(ctx context.Context, msgNum *int, month *provider.MonthYear) (*provider.Message, error) {
	msgName := msgNumToName(*msgNum)
	l, err := url.JoinPath(h.config.ConnStr, MonthYearPathSeg(month.Month, month.Year), msgName)
	if err != nil {
		return nil, fmt.Errorf("checkMsg: %w", err)
	}

	msg, err := h.getMsgByLink(ctx, l)
	if err == nil { //err is EQUAL to nil !!! Getting is succeful
		*msgNum++
		return msg, nil
	}

	if !errors.As(err, &NotFoundErr{}) { //all errors except NotFoundErr
		return nil, err
	}

	//NotFound error: check next month
	nextMonth := month.AddMonth(1)
	l, err = url.JoinPath(h.config.ConnStr, MonthYearPathSeg(nextMonth.Month, nextMonth.Year), msgName)
	if err != nil {
		return nil, fmt.Errorf("checkMsg: %w", err)
	}

	msg, err = h.getMsgByLink(ctx, l)
	if err == nil { //err is EQUAL to nil !!! a message has been found in the next month
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

// getStartMsgNum finds a message number according to the logic described below
// and sends it into the "sn" channel.
//
// The function fetches all messages for the month of the "from" parameter and
// searches among the messages for the one that has the minimum number and
// FocusTime of which is after (or equal to) the value of the "from" parameter.
// This logic is neccesary because SEISHUB DOESN'T ENSURE that a message
// about a later event has a higher number.
// If no messages are fetched, i.e., there are no messages in the specified
// (as a rule current) month yet, fetching will be repeated with a frequency
// of "checkPeriod" until at least one message is received.
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

// findStartMsgNum searches among the messages for the one that has the minimum
// number and FocusTime of which is after (or equal to) the value of the "from"
// parameter.
// This logic is neccesary because SEISHUB DOESN'T ENSURE that a message
// about a later event has a higher number.
//
// The function returns a number and an error. If the returned error is not nil,
// the returned number is 0.
func findStartMsgNum(msgs []*provider.Message, from time.Time) (int, error) {
	//Create a "meta" slice ordered by a message number (parsed from the link)
	//and find the first message with focus time more than the "from" arg

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
	m.SourceId = h.config.Id

	return m, nil
}

// Extract returns a slice of seismic messages extracted from SEISHUB for the period defined
// with the "from" and "to" parameters and error.
// If the returned error is not nil, the returned slice is nil.
//
// The "paral" parameter defines a number of go-routines to process message links (getting messages).
// if "paral" is less (or equal to) 0, the default falue is used.
//
// Attention! The function does not guarantee immediate termination by context cancellation.
func (h *Hub) Extract(ctx context.Context,
	from provider.MonthYear, to provider.MonthYear, paral int) ([]*provider.Message, error) {
	links := make(chan string)
	defer close(links)

	monthNum := to.Diff(from) + 1
	if monthNum <= 0 {
		return nil, fmt.Errorf(`Extract: the "from" arg cannot be more than the "to" arg`)
	}

	if paral <= 0 {
		paral = defParal
	}

	//Result slice of messages
	msgs := make([]*provider.Message, 0, avgMonthMsgNum*monthNum)

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
		monthLink, err := url.JoinPath(h.config.ConnStr, sg)
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
			l, err := url.JoinPath(monthLink, n)
			if err != nil {
				return nil, fmt.Errorf("Extract: %v", err)
			}
			links <- l
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

// msgNumToName creates a message page name like "000234.html".
// It panics if n is less than 0 or has more than 6 digits.
func msgNumToName(n int) string {
	if n < 0 {
		panic("msgNumToName: the n arg is less than 0")
	}
	s := strconv.Itoa(n)
	return strings.Repeat("0", 6-len(s)) + s + ".html"
}
