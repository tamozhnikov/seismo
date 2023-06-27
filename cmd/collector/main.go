package main

import (
	"context"
	"fmt"
	"log"
	"seismo"
	"seismo/pr/pseudo"
	"time"
)

func main() {
	var w seismo.Watcher = pseudo.NewHub("pseudo") //seishub.NewHub("seishub", "", 0)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	//ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := w.StartWatch(ctx, time.Date(2023, 6, 24, 12, 0, 0, 0, time.UTC), time.Second*2)
	if err != nil {
		log.Printf("Cannot start watching: %v", err)
		return
	}

	for m := range ch {
		fmt.Println(m)
	}
}
