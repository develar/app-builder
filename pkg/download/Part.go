package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/apex/log"
	"github.com/develar/go-fs-util"
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

func (part *Part) download(context context.Context, url string, index int, client *http.Client) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req = req.WithContext(context)
	req.Header.Set("User-Agent", userAgent)
	if part.End > 0 {
		req.Header.Set("Range", part.getRange())
	}

	response, err := client.Do(req)
	if err != nil {
		return err
	}

	if response.StatusCode == http.StatusOK {
		if part.End > 0 {
			if index > 0 {
				part.Skip = true
				return nil
			}
			part.End = response.ContentLength
		}
	} else if response.StatusCode != http.StatusPartialContent {
		return nil
	}

	partFile, err := os.OpenFile(part.Name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	defer partFile.Close()

	buf := make([]byte, 32*1024)
	for i := 0; i <= 3; i++ {
		if i > 0 {
			time.Sleep(2 * time.Second)
			log.Infof("Retrying (%d)", i)
			if part.End > 0 {
				req.Header.Set("Range", part.getRange())
			}
			response, err = client.Do(req)
			if err != nil {
				if response != nil {
					err = fsutil.CloseAndCheckError(err, response.Body)
				}
				if i == 3 {
					return err
				}
				continue
			}
		}

		err = part.writeToFile(partFile, response, &buf)
		if err == nil || context.Err() != nil {
			//if total <= 0 {
			//	//bar.Complete()
			//}
			return nil
		}

		if i == 3 {
			return err
		}
	}

	return nil
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
