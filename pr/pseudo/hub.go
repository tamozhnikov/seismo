package pseudo

import (
	"context"
	"fmt"
	"math/rand"
	"seismo"
	"strconv"
	"time"
)

type hubState interface {
	startWatch(ctx context.Context, from time.Time, checkPeriod time.Duration) (<-chan seismo.Message, error)
	stateInfo() seismo.WatcherStateInfo
}

// stoppedState implements a stopped Hub's behavior within the STATE PATTERN
type stoppedState struct {
	hub *Hub
}

func newStoppedState(h *Hub) *stoppedState {
	return &stoppedState{hub: h}
}

func (s *stoppedState) startWatch(ctx context.Context, from time.Time, checkPeriod time.Duration) (<-chan seismo.Message, error) {
	if checkPeriod < time.Second {
		return nil, fmt.Errorf("checkPeriod cannot be less than 1 sec")
	}
	h := s.hub
	h.setState(newRunState(h))
	o := make(chan seismo.Message)
	go h.generateMessages(ctx, o, checkPeriod)

	return o, nil
}

func (s *stoppedState) stateInfo() seismo.WatcherStateInfo {
	return seismo.Stopped
}

// runState implements a running Hub's behavior within the STATE PATTERN
type runState struct {
	hub *Hub
}

func newRunState(h *Hub) *runState {
	return &runState{hub: h}
}

func (r *runState) startWatch(ctx context.Context, from time.Time, checkPeriod time.Duration) (<-chan seismo.Message, error) {
	return nil, seismo.AlreadyRunErr{}
}

func (r *runState) stateInfo() seismo.WatcherStateInfo {
	return seismo.Run
}

// Hub emulates getting seismic event messages
// implementing the seismo.Watcher interface
// and creating new messages randomly
type Hub struct {
	//state implements the State pattern
	state hubState
}

func NewHub() *Hub {
	h := &Hub{}
	h.setState(newStoppedState(h))

	return h
}

func (h *Hub) setState(s hubState) {
	h.state = s
}

// StateInfo reports a current state of the hub
func (h *Hub) StateInfo() seismo.WatcherStateInfo {
	return h.state.stateInfo()
}

// StartWatch starts generating several (1 to 3) random seismic messages every checkPeriod,
// the FocusTime of every message corresponds to its generating moment,
// the from argument is ignored
// Returns a channel for getting these messages.
func (h *Hub) StartWatch(ctx context.Context, from time.Time, checkPeriod time.Duration) (<-chan seismo.Message, error) {
	o, err := h.state.startWatch(ctx, from, checkPeriod)
	return o, err
}

func (h *Hub) generateMessages(ctx context.Context, o chan<- seismo.Message, period time.Duration) {
	defer func() {
		close(o)
		h.setState(newStoppedState(h))
	}()
	for {
		if ctx.Err() != nil {
			return
		}
		msgs := createRandMsgs()
		for _, m := range msgs {
			o <- m
		}
		time.Sleep(period)
	}
}

// createRandMsgs returns a slice containing 1 to 3 messages
// with the same EventId
func createRandMsgs() []seismo.Message {
	rand.Seed(time.Now().UnixNano())
	num := rand.Intn(3) + 1
	msgs := make([]seismo.Message, 0, num)

	id := strconv.Itoa(rand.Int())
	lat := rand.Float64()*20.0 + 40.0
	long := rand.Float64()*30.0 + 70.0
	mag := rand.Float64()*6.0 + 0.1

	for i := 0; i < num; i++ {

		m := seismo.Message{}

		m.FocusTime = time.Now().UTC()
		m.Latitude = lat + lat*((rand.Float64()-0.5)/100.0)
		m.Longitude = long + long*((rand.Float64()-0.5)/100.0)
		m.Magnitude = mag
		m.EventId = id
		m.Type = seismo.RandEventType()
		m.Quality = seismo.RandEventQuality()

		msgs = append(msgs, m)
	}

	return msgs
}
