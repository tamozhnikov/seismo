package pseudo

import (
	"context"
	"fmt"
	"math/rand"
	"seismo/provider"
	"time"

	"github.com/google/uuid"
)

// hubState is implemented to provide a specific behavior within THE STATE PATTERN.
type hubState interface {
	startWatch(ctx context.Context, from time.Time) (<-chan provider.Message, error)
	stateInfo() provider.WatcherStateInfo
}

// stoppedState implements a stopped Hub's behavior within THE STATE PATTERN.
type stoppedState struct {
	hub *Hub
}

func newStoppedState(h *Hub) *stoppedState {
	return &stoppedState{hub: h}
}

func (s *stoppedState) startWatch(ctx context.Context, from time.Time) (<-chan provider.Message, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("cannot start with canceled context")
	}
	h := s.hub
	h.setState(newRunState(h))
	o := make(chan provider.Message)
	go h.generateMessages(ctx, o, from)

	return o, nil
}

func (s *stoppedState) stateInfo() provider.WatcherStateInfo {
	return provider.Stopped
}

// runState implements a running Hub's behavior within THE STATE PATTERN.
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

// Hub implements the provider.Watcher interface,
// emulates getting seismic event messages,
// creating new messages randomly.
type Hub struct {
	config provider.WatcherConfig

	//state implements THE STATE PATTERN
	state hubState
}

// NewHub returns a pointer to a new pseudo.Hub in the stopped state and an error.
//
// If the returned error is not nil, the returned pointer value is nil.
func NewHub(conf provider.WatcherConfig) (*Hub, error) {
	if conf.CheckPeriod < 1 {
		return nil, fmt.Errorf("NewHub: checkperiod cannot be less than 1 second")
	}

	h := &Hub{}
	h.config = conf

	h.setState(newStoppedState(h))

	return h, nil
}

// GetConfig returns configuration of the Hub.
func (h *Hub) GetConfig() provider.WatcherConfig {
	return h.config
}

func (h *Hub) setState(s hubState) {
	h.state = s
}

// StateInfo reports a current state of the Hub
func (h *Hub) StateInfo() provider.WatcherStateInfo {
	return h.state.stateInfo()
}

// StartWatch starts generating several (1 to 3) random seismic messages every checkPeriod.
//
// The methods returns a channel for fetching messages. If the returned error is not nil, the returned
// channel is nil.
//
// The FocusTime of every message corresponds to its generating moment minus an offset.
// The offset is calculated as the differrence between the moment the function is called
// and the value of the "from" argument.
func (h *Hub) StartWatch(ctx context.Context, from time.Time) (<-chan provider.Message, error) {
	o, err := h.state.startWatch(ctx, from)
	return o, err
}

func (h *Hub) generateMessages(ctx context.Context, o chan<- provider.Message, from time.Time) {
	defer func() {
		close(o)
		h.setState(newStoppedState(h))
	}()

	offset := time.Now().UTC().Sub(from)
	for {
		for _, m := range h.createRandMsgs(offset) {
			if ctx.Err() != nil {
				return
			}

			select {
			case o <- m:
			case <-ctx.Done():
				return
			}
		}
		time.Sleep(time.Duration(h.config.CheckPeriod) * time.Second)
	}
}

// createRandMsgs returns a slice containing 1 to 3 messages with the same EventId.
// The offset argument shifts every value of the FocusTime from the moment the message
// is generated.
func (h *Hub) createRandMsgs(offset time.Duration) []provider.Message {
	rand.Seed(time.Now().UnixNano())
	num := rand.Intn(3) + 1
	msgs := make([]provider.Message, 0, num)

	id := uuid.New().String()
	lat := rand.Float64()*20.0 + 40.0
	long := rand.Float64()*30.0 + 70.0
	mag := rand.Float64()*6.0 + 0.1

	for i := 0; i < num; i++ {

		m := provider.Message{}

		m.SourceId = h.config.Id
		m.FocusTime = time.Now().UTC().Add(-offset)
		m.Latitude = lat + lat*((rand.Float64()-0.5)/100.0)
		m.Longitude = long + long*((rand.Float64()-0.5)/100.0)
		m.Magnitude = mag
		m.EventId = id
		m.Type = provider.RandEventType()
		m.Quality = provider.RandEventQuality()

		msgs = append(msgs, m)
	}

	return msgs
}
