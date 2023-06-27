package pseudo

import (
	"fmt"
	"testing"
)

// Test_createRandMsgs is ONLY for launching and
// debugging createRandMsgs. This test function
// DOES NOT compares result and expected results
// and expected values
func Test_createRandMsgs(t *testing.T) {
	//t.Errorf("")
	h := NewHub("pseudo")
	for i := 0; i < 10; i++ {
		msgs := h.createRandMsgs()

		for _, m := range msgs {
			fmt.Println(m)
		}

		fmt.Println("-------")
	}
}

// func Test_Hub_StartWatch(t *testing.T) {
// 	//t.Error("")
// 	var w seismo.Watcher = NewHub()
// 	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
// 	defer cancel()

// 	ch, err := w.StartWatch(ctx, time.Now(), time.Second)
// 	if err != nil {
// 		t.Fatalf("start watching error: %v", err)
// 	}

// 	_, err = w.StartWatch(ctx, time.Now(), time.Second)
// 	if err != nil {
// 		t.Errorf("starting watch error: %v", err)
// 	}

// 	for m := range ch {
// 		fmt.Println(m)
// 	}

// 	fmt.Println("!!!!!!")

// 	// ctx1, cancel1 := context.WithTimeout(context.Background(), 6*time.Second)
// 	// defer cancel1()

// 	ch1, err := w.StartWatch(ctx, time.Now(), time.Second)
// 	if err != nil {
// 		t.Errorf("starting watch error: %v", err)
// 	}

// 	for m := range ch1 {
// 		fmt.Println(m)
// 	}

// }
