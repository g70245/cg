package container

import (
	"fmt"
	"sync"
	"testing"
)

func TestRobotGameDirectoryConcurrentAccess(t *testing.T) {
	robot := &robot{}
	var waitGroup sync.WaitGroup

	for i := 0; i < 50; i++ {
		i := i
		waitGroup.Add(2)
		go func() {
			defer waitGroup.Done()
			robot.setGameDir(fmt.Sprintf("logs-%d", i))
		}()
		go func() {
			defer waitGroup.Done()
			_ = robot.getGameDir()
		}()
	}

	waitGroup.Wait()
}
