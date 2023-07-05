package pseudo

import (
	"context"
	"fmt"
	"seismo/provider"
	"testing"
	"time"
)

// Test_createRandMsgs is ONLY for launching and
// debugging createRandMsgs. This test function
// DOES NOT compares result and expected results
// and expected values
func Test_createRandMsgs(t *testing.T) {
	//t.Errorf("")
	c := provider.WatcherConfig{Id: "pseudo", CheckPeriod: 1}
	h, _ := NewHub(c)
	for i := 0; i < 10; i++ {
		msgs := h.createRandMsgs()

		for _, m := range msgs {
			fmt.Println(m)
		}

		fmt.Println("-------")
	}
}

func Test_Hub_StartWatch(t *testing.T) {
	t.Error("")
	conf := provider.DefaultWatcherConfig()
	w, _ := NewHub(conf)
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	ch, err := w.StartWatch(ctx, time.Now())
	if err != nil {
		t.Fatalf("start watching error: %v", err)
	}

	_, err = w.StartWatch(ctx, time.Now())
	if err != nil {
		t.Errorf("starting watch error: %v", err)
	}

	for m := range ch {
		fmt.Println(m)
	}

	fmt.Println("!!!!!!")

	fmt.Printf("State %s", w.StateInfo())

	_, err = w.StartWatch(ctx, time.Now())
	if err != nil {
		t.Errorf("starting watch error: %v", err)
	}

	fmt.Println("!!!!!!")
}
