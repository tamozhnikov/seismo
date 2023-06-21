package seismo

import (
	"flag"
	"fmt"
	"time"
)

// MonthYear represents a month of a year
type MonthYear struct {
	Month time.Month
	Year  int
}

// String represents a value in month/year format
func (m *MonthYear) String() string {
	return fmt.Sprintf("%d/%d", m.Month, m.Year)
}

// Date creates UTC date with the 1st day of month
func (m *MonthYear) Date() time.Time {
	return time.Date(m.Year, m.Month, 1, 0, 0, 0, 0, time.UTC)
}

// After reports whether the MonthYear instant m is after u
func (m *MonthYear) After(u MonthYear) bool {
	if m.Year > u.Year || (m.Year == u.Year && m.Month > u.Month) {
		return true
	}

	return false
}

// AddMonth adds n months to the MonthYear instant value
func (m *MonthYear) AddMonth(n int) MonthYear {
	res := *m
	res.Year += n / 12
	md := int(m.Month) + n%12
	switch {
	case md > 12:
		res.Year++
		res.Month = time.Month(md - 12)
	case md <= 0:
		res.Year--
		res.Month = time.Month(12 + md)
	default:
		res.Month = time.Month(md)
	}

	return res
}

// Diff returns difference in months. A returned value is negative if u is after m.
func (m *MonthYear) Diff(u MonthYear) int {
	return (m.Year-u.Year)*12 + (int(m.Month) - int(u.Month))
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
