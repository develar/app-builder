package util

import (
	"runtime"

	"github.com/apex/log"
	"github.com/develar/errors"
)

func MapAsync(taskCount int, taskProducer func(taskIndex int) (func() error, error)) error {
	return MapAsyncConcurrency(taskCount, runtime.NumCPU(), taskProducer)
}

func MapAsyncConcurrency(taskCount int, concurrency int, taskProducer func(taskIndex int) (func() error, error)) error {
	if taskCount == 0 {
		return nil
	}

	log.WithField("taskCount", taskCount).Debug("map async")

	errorChannel := make(chan error)
	doneChannel := make(chan bool, taskCount)
	quitChannel := make(chan bool)

	sem := make(chan bool, concurrency)
	for i := 0; i < taskCount; i++ {
		// wait semaphore
		sem <- true

		task, err := taskProducer(i)
		if err != nil {
			close(quitChannel)
			return errors.WithStack(err)
		}

		if task == nil {
			<-sem
			doneChannel <- true
			continue
		}

		go func(task func() error) {
			defer func() {
				// release semaphore, notify done
				<-sem
				doneChannel <- true
			}()

			// select waits on multiple channels, if quitChannel is closed, read will succeed without blocking
			// the default case in a select is run if no other case is ready
			select {
			case <-quitChannel:
				return

			default:
				err := task()
				if err != nil {
					errorChannel <- errors.WithStack(err)
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

