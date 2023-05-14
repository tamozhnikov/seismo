package seismo

import (
	"flag"
	"fmt"
	"time"
)

// Message represents contains common information about seismic event
type Message struct {
	EventId   string    `json:"event_id"`
	FocusTime time.Time `json:"focus_time"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	//Depth     float64   `json:"depth"`
	Magnitude float64 `json:"magnitude"`
	EventType string  `json:"event_type"`
	Quality   string  `json:"quality"`
}

// MonthYear represents a month of a year
type MonthYear struct {
	Month time.Month
	Year  int
}

// String represents a value in month/year format
func (my *MonthYear) String() string {
	return fmt.Sprintf("%d/%d", my.Month, my.Year)
}

// Date creates UTC date with the 1st day of month
func (my *MonthYear) Date() time.Time {
	return time.Date(my.Year, my.Month, 1, 0, 0, 0, 0, time.UTC)
}

type monthYearFlag struct {
	MonthYear
}

func (f *monthYearFlag) Set(s string) error {
	var year int
	var month time.Month
	fmt.Sscanf(s, "%d/%d", &month, &year)
	if month >= 1 && month <= 12 && year >= 1 && year <= 9999 {
		f.Month = month
		f.Year = year
		return nil
	}

	return fmt.Errorf("incorrect month/date format or value %q", s)
}

func MonthYearFlag(name string, value MonthYear, usage string) *MonthYear {
	f := monthYearFlag{value}
	flag.CommandLine.Var(&f, name, usage)
	return &f.MonthYear
}
