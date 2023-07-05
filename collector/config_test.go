package collector

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_ConfigFromFile(t *testing.T) {
	_, err := ConfigFromFile("")
	if err == nil {
		t.Errorf("configFromFile didn't rase an err on an empty file name")
	}

	want := DefaultConfig()
	bf, err := json.Marshal(want)
	if err != nil {
		panic(err)
	}

	fileName := "testdata/simple_conf.json"
	err = os.WriteFile(fileName, bf, os.ModePerm)
	if err != nil {
		panic(err)
	}

	res, err := ConfigFromFile(fileName)
	if err != nil {
		t.Errorf("configFromFile: name: %s error: %v", fileName, err)
	}

	if !cmp.Equal(want, res) {
		t.Errorf("configFromFile: name: %s want: %v res: %v", fileName, want, res)
	}
}
