package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/apex/log"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/go-fs-util"
	"github.com/pkg/errors"
)

const maxAttemptNumber = 3

type Part struct {
	Name string

	Start int64
	End   int64

	Skip   bool
	isFail bool
}

func (part *Part) getRange() string {
	return fmt.Sprintf("bytes=%d-%d", part.Start, part.End-1)
}

func (part *Part) download(context context.Context, url string, index int, client *http.Client) error {
	// request cannot be reused because Range header is set
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	request = request.WithContext(context)
	request.Header.Set("User-Agent", userAgent)
	if part.End > 0 {
		request.Header.Set("Range", part.getRange())
	}

	response, err := part.doRequest(request, client, index)
	if err != nil {
		return err
	}
	if response == nil {
		return nil
	}

	partFile, err := os.OpenFile(part.Name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fsutil.CloseAndCheckError(err, response.Body)
	}

	defer util.Close(partFile)

	buf := make([]byte, 32*1024)
	for attemptNumber := 0; ; attemptNumber++ {
		if attemptNumber != 0 {
			time.Sleep(2 * time.Second)
			log.Infof("retrying (%d)", attemptNumber)
			response, err = part.doRequest(request, client, index)
			if err != nil {
				if response != nil {
					err = fsutil.CloseAndCheckError(err, response.Body)
				}
				if attemptNumber == maxAttemptNumber {
					return errors.WithStack(err)
				}
				continue
			}
		}

		written, err := writeToFile(partFile, response, &buf)
		if err == nil || request.Context().Err() != nil {
			return nil
		}

		if attemptNumber == maxAttemptNumber {
			return errors.WithStack(err)
		}

		if part.End > 0 {
			part.Start += written
			_, err = partFile.Seek(part.Start, io.SeekStart)
			if err != nil {
				return errors.WithStack(err)
			}
			request.Header.Set("Range", part.getRange())
		} else {
			_, err = partFile.Seek(0, io.SeekStart)
			if err != nil {
				return errors.WithStack(err)
			}
		}
	}
}

func (part *Part) doRequest(request *http.Request, client *http.Client, index int) (*http.Response, error) {
	response, err := client.Do(request)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if response.StatusCode == http.StatusPartialContent {
		return response, nil
	} else if response.StatusCode == http.StatusOK {
		if part.End > 0 {
			if index > 0 {
				part.Skip = true
				util.Close(response.Body)
				return nil, nil
			}
			part.End = response.ContentLength
		}
		return response, nil
	} else {
		util.Close(response.Body)
		return nil, errors.WithStack(fmt.Errorf("part download request failed with status code %d", response.StatusCode))
	}
}

func writeToFile(file *os.File, response *http.Response, buffer *[]byte) (int64, error) {
	defer util.Close(response.Body)
	return io.CopyBuffer(file, response.Body, *buffer)
}
