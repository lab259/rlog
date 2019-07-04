package main

import (
	"math/rand"
	"sync"
	"time"

	"github.com/lab259/rlog/v2"
)

func main() {
	rlog.WithField("scope", "system").Info("Starting system")
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(logger rlog.Logger) {
			defer wg.Done()
			ms := rand.Intn(50)
			logger.Infof("Going to sleep for %dms", ms)
			time.Sleep(time.Millisecond * time.Duration(ms))
			logger.Info("Exiting ...")
		}(rlog.WithFields(rlog.Fields{
			"i": i,
		}))
	}
	rlog.Info("Waiting routines ...")
	wg.Wait()
	rlog.Info("OK!")
}
