package production

import (
	"fmt"
	"sync"
	"testing"
)

func TestWorkerConcurrentConfigurationAccess(t *testing.T) {
	worker := NewWorker(1, func() string { return "logs" }, make(chan bool, 1))
	var waitGroup sync.WaitGroup

	for i := 0; i < 50; i++ {
		i := i
		waitGroup.Add(2)
		go func() {
			defer waitGroup.Done()
			worker.SetName(fmt.Sprint(i))
			worker.SetGatheringMode(i%2 == 0)
			worker.manualMode.Store(i%2 != 0)
		}()
		go func() {
			defer waitGroup.Done()
			_ = worker.Name()
			_ = worker.gatheringMode.Load()
			_ = worker.manualMode.Load()
		}()
	}

	waitGroup.Wait()
}
