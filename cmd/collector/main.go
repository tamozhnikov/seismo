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
	var w seismo.Watcher = pseudo.NewHub()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	ch, err := w.StartWatch(ctx, time.Now(), time.Second*2)
	if err != nil {
		log.Printf("Cannot start watching: %v", err)
		return
	}

	for m := range ch {
		fmt.Println(m)
	}
}
