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

	//Quality values

	UnknownQuality EventQuality = 0
	Preliminary    EventQuality = 1
	Good           EventQuality = 2
	Excellent      EventQuality = 3
)

// EventType represents the type of a sesmic event.
type EventType int

// RandEventType creates a random value of the EventType type.
func RandEventType() EventType {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return EventType(r.Intn(int(QuarryBlast)))
}

// EventQuality represents quality of a seismic event assessment.
type EventQuality int

// RandEventQuality creates a random value of the EventQuality type.
func RandEventQuality() EventQuality {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return EventQuality(r.Intn(int(Excellent)))
}

// Message contains common information about a seismic event.
type Message struct {
	// SourceId specifies the string identifier of the message source,
	// that reported the event.
	SourceId string `json:"source_id" bson:"source_id"`

	// FocusTime specifies the UTC-time of the seismic event at its epicenter.
	FocusTime time.Time `json:"focus_time" bson:"focus_time"`

	// Latitude specifies the latitude of the event epicenter.
	Latitude float64 `json:"latitude" bson:"latitude"`

	// Longitude specifies the longitide of the event epicenter.
	Longitude float64 `json:"longitude" bson:"longitude"`

	// Magnitude specifies the magnitude of the event.
	Magnitude float64 `json:"magnitude" bson:"magnitude"`

	// EventId specifies the identifier which was assigned the the event
	// by the message source. EventId must be unique for the source.
	EventId string `json:"event_id" bson:"event_id"`

	// Type specifies the type of the seismic event.
	Type EventType `json:"event_type" bson:"event_type"`

	// Quality specifies quality of the message data.
	Quality EventQuality `json:"quality" bson:"quality"`

	// Link specifies a url of the message, i.e. an address,
	// from which the message can be fetched again. Optional.
	Link string `json:"link" bson:"link"`
}
