package provider

import (
	"math/rand"
	"time"
)

const (
	//EventType values
	UnknownType EventType = 0
	EarthQuake  EventType = 1
	QuarryBlast EventType = 2

	//Quality int
	UnknownQuality EventQuality = 0
	Preliminary    EventQuality = 1
	Good           EventQuality = 2
	Excellent      EventQuality = 3
)

type EventType int

func RandEventType() EventType {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return EventType(r.Intn(int(QuarryBlast)))
}

type EventQuality int

func RandEventQuality() EventQuality {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return EventQuality(r.Intn(int(Excellent)))
}

// Message contains common information about seismic event
type Message struct {
	SourceId  string    `json:"source_id"`
	FocusTime time.Time `json:"focus_time"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	//Depth     float64   `json:"depth"`
	Magnitude float64      `json:"magnitude"`
	EventId   string       `json:"event_id"`
	Type      EventType    `json:"event_type"`
	Quality   EventQuality `json:"quality"`
	Link      string       `json:"link"`
}
