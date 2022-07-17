package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pdavid31/ptop/internal/cpu"
	intOS "github.com/pdavid31/ptop/internal/os"
)

func main() {
	_, err := intOS.GetOS()
	if err != nil {
		log.Fatal(err)
	}

	c, err := cpu.New()
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	ticker := time.NewTicker(1 * time.Second)

	// create shutdown channel
	shutdown := make(chan os.Signal, 1)
	// redirect system signals to the shutdown
	// channel to gracefully shut down the goroutines
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	func() {
		for {
			select {
			case <-ticker.C:
				err := c.Update()
				if err != nil {
					log.Println(err)
					return
				}
			case <-shutdown:
				return
			default:
				fmt.Println(c)
			}
		}
	}()
}
