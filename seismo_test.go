package seismo

import (
	"testing"
)

func Test_MonthYear_After(t *testing.T) {
	tests := []struct {
		base MonthYear
		arg  MonthYear
		want bool
	}{
		{MonthYear{10, 2003}, MonthYear{9, 2003}, true},
		{MonthYear{10, 2003}, MonthYear{10, 2002}, true},
		{MonthYear{10, 2003}, MonthYear{12, 2002}, true},
		{MonthYear{10, 2003}, MonthYear{3, 2002}, true},

		{MonthYear{10, 2003}, MonthYear{10, 2003}, false},
		{MonthYear{10, 2003}, MonthYear{11, 2003}, false},
		{MonthYear{10, 2003}, MonthYear{11, 2004}, false},
		{MonthYear{10, 2003}, MonthYear{10, 2004}, false},
		{MonthYear{10, 2003}, MonthYear{3, 2004}, false},
	}

	for _, test := range tests {
		if r := test.base.After(test.arg); r != test.want {
			t.Errorf("Test MonthYear.After: base: %v, arg: %v : want: %v, res: %v", test.base, test.arg, test.want, r)
		}
	}
}

func Test_MonthYear_AddMonth(t *testing.T) {
	tests := []struct {
		base MonthYear
		arg  int
		want MonthYear
	}{
		{MonthYear{10, 2003}, 1, MonthYear{11, 2003}},
		{MonthYear{10, 2003}, 2, MonthYear{12, 2003}},
		{MonthYear{10, 2003}, 3, MonthYear{1, 2004}},
		{MonthYear{10, 2003}, 5, MonthYear{3, 2004}},
		{MonthYear{10, 2003}, 11, MonthYear{9, 2004}},
		{MonthYear{10, 2003}, 12, MonthYear{10, 2004}},
		{MonthYear{10, 2003}, 22, MonthYear{8, 2005}},
		{MonthYear{10, 2003}, 24, MonthYear{10, 2005}},
		{MonthYear{10, 2003}, 120, MonthYear{10, 2013}},

		{MonthYear{10, 2003}, -1, MonthYear{9, 2003}},
		{MonthYear{10, 2003}, -2, MonthYear{8, 2003}},
		{MonthYear{10, 2003}, -9, MonthYear{1, 2003}},
		{MonthYear{10, 2003}, -10, MonthYear{12, 2002}},
		{MonthYear{10, 2003}, -11, MonthYear{11, 2002}},
		{MonthYear{10, 2003}, -12, MonthYear{10, 2002}},
		{MonthYear{10, 2003}, -22, MonthYear{12, 2001}},
		{MonthYear{10, 2003}, -24, MonthYear{10, 2001}},
		{MonthYear{10, 2003}, -120, MonthYear{10, 1993}},
	}

	for _, test := range tests {
		preBase := test.base
		if test.base.AddMonth(test.arg); test.base != test.want {
			t.Errorf("Test MonthYear.AddMonth: base: %v, arg: %d, want: %v, res: %v", preBase, test.arg, test.want, test.base)
		}
	}
}
