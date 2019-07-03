package util

import (
	"runtime"

	"github.com/develar/app-builder/pkg/log"
	"github.com/develar/errors"
	"go.uber.org/zap"
)

func MapAsync(taskCount int, taskProducer func(taskIndex int) (func() error, error)) error {
	return MapAsyncConcurrency(taskCount, runtime.NumCPU() + 1, taskProducer)
}

func MapAsyncConcurrency(taskCount int, concurrency int, taskProducer func(taskIndex int) (func() error, error)) error {
	if taskCount == 0 {
		return nil
	}

	log.Debug("map async", zap.Int("taskCount", taskCount))

	errorChannel := make(chan error, concurrency)
	doneChannel := make(chan bool, taskCount)
	quitChannel := make(chan struct{})
	sem := make(chan bool, concurrency)

	markDone := func() {
		// release semaphore, notify done
		doneChannel <- true
		select {
		case <-sem:
			return
		case <-errorChannel:
			break
		}
	}

	for i := 0; i < taskCount; i++ {
		// wait semaphore
		select {
		case <-errorChannel:
			break
		case sem <- true:
			// ok
		}

		task, err := taskProducer(i)
		if err != nil {
			close(quitChannel)
			return errors.WithStack(err)
		}

		if task == nil {
			markDone()
			continue
		}

		go func(task func() error) {
			defer markDone()

			// select waits on multiple channels, if quitChannel is closed, read will succeed without blocking
			// the default case in a select is run if no other case is ready
			select {
			case <-quitChannel:
				return

			default:
				err := task()
				if err != nil {
					// do not wrap - up to client to wrap if needed (to avoid later to discover cause)
					errorChannel <- err
				}
			}
		}(task)
	}

	finishedCount := 0
	for {
		select {
		case err := <-errorChannel:
			close(quitChannel)
			return err

		case <-doneChannel:
			finishedCount++
			if finishedCount == taskCount {
				return nil
			}
		}
	}
}
