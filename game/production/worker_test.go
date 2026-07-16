package production

import (
	"fmt"
	"sync"
	"testing"
	"time"
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

func TestWorkerBeginWorkAllowsOnlyOneConcurrentStart(t *testing.T) {
	worker := &Worker{}
	const attempts = 50

	start := make(chan struct{})
	results := make(chan bool, attempts)
	var waitGroup sync.WaitGroup
	for i := 0; i < attempts; i++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			results <- worker.beginWork()
		}()
	}

	close(start)
	waitGroup.Wait()
	close(results)

	successes := 0
	for result := range results {
		if result {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("successful starts = %d, want 1", successes)
	}
}

func TestWorkerCanBeginWorkAgainAfterFinish(t *testing.T) {
	worker := &Worker{}
	if !worker.beginWork() {
		t.Fatal("first beginWork() = false, want true")
	}
	if worker.beginWork() {
		t.Fatal("second beginWork() = true while running, want false")
	}

	worker.finishWork()
	if !worker.beginWork() {
		t.Fatal("beginWork() after finishWork() = false, want true")
	}
}

func TestWaitForProductionCompletionReturnsWhenAlreadySuccessful(t *testing.T) {
	stopped := waitForProductionCompletion(make(chan bool), time.Hour, func() bool { return true })
	if stopped {
		t.Fatal("waitForProductionCompletion() = stopped, want completed")
	}
}

func TestWaitForProductionCompletionConsumesStopSignal(t *testing.T) {
	stopChan := make(chan bool, 1)
	stopChan <- true

	stopped := waitForProductionCompletion(stopChan, time.Hour, func() bool { return false })
	if !stopped {
		t.Fatal("waitForProductionCompletion() = completed, want stopped")
	}
}
