package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

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

func (part *Part) download(context context.Context, url string, index int, client *http.Client) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		panic(err)
	}

	req = req.WithContext(context)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Range", part.getRange())

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	total := part.End - part.Start
	if resp.StatusCode == http.StatusOK {
		if index > 0 {
			part.Skip = true
			return
		}
		total = resp.ContentLength
		part.End = total
	} else if resp.StatusCode != http.StatusPartialContent {
		return
	}

	messageCh := make(chan string, 1)
	failureCh := make(chan struct{})

	fail := func(err error) {
		close(failureCh)
		part.isFail = true
		errs := strings.Split(err.Error(), ":")
		messageCh <- errs[len(errs)-1]
	}

	partFile, err := os.OpenFile(part.Name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fail(err)
		return
	}

	defer partFile.Close()

	buf := make([]byte, 32*1024)
	for i := 0; i <= 3; i++ {
		if i > 0 {
			time.Sleep(2 * time.Second)
			messageCh <- fmt.Sprintf("Retrying (%d)", i)
			req.Header.Set("Range", part.getRange())
			resp, err = client.Do(req)
			if err != nil {
				if resp != nil {
					resp.Body.Close()
				}
				if i == 3 {
					fail(err)
				}
				continue
			}
		}

		err = part.writeToFile(partFile, resp, &buf)

		if err == nil || context.Err() != nil {
			if total <= 0 {
				//bar.Complete()
			}
			return
		}

		if i == 3 {
			fail(err)
		} else {
			messageCh <- "Error..."
		}
	}
}

func (part *Part) writeToFile(dst *os.File, resp *http.Response, buf *[]byte) (err error) {
	defer resp.Body.Close()

	reader := resp.Body
	for i := 0; i < 3; i++ {
		_, err := io.CopyBuffer(dst, reader, *buf)
		if err != nil && isTemporary(err) {
			time.Sleep(1e9)
			continue
		}
		break
	}

	return
}
