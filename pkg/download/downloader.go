package download

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/alecthomas/kingpin"
	"github.com/apex/log"
	"github.com/develar/errors"
	"github.com/mitchellh/go-homedir"
	"github.com/zieckey/goini"
)

//noinspection SpellCheckingInspection
const (
	maxRedirects = 10
	minPartSize  = 5 * 1024 * 1024
	maxPartCount = 8
	userAgent    = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_3) AppleWebKit/604.5.6 (KHTML, like Gecko) Version/11.0.3 Safari/604.5.6"
)

func ConfigureCommand(app *kingpin.Application) {
	command := app.Command("download", "Download file.")
	fileUrl := command.Flag("url", "The URL.").Short('u').Required().String()
	output := command.Flag("output", "The output file.").Short('o').Required().String()
	sha512 := command.Flag("sha512", "The expected sha512 of file.").String()

	command.Action(func(context *kingpin.ParseContext) error {
		return errors.WithStack(Download(*fileUrl, *output, *sha512))
	})
}

func Download(url string, output string, sha512 string) error {
	dir := filepath.Dir(output)
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return errors.WithStack(err)
	}

	var waitGroup sync.WaitGroup
	recoverIfPanic := func(id int) {
		if e := recover(); e != nil {
			log.WithFields(log.Fields{
				"id":    id,
				"error": e,
			}).Debug("part download error")
		}
		waitGroup.Done()
	}

	downloadContext, cancel := context.WithCancel(context.Background())
	httpTransport := &http.Transport{Proxy: proxyFromEnvironmentAndNpm}

	actualLocation, err := follow(url, userAgent, output, httpTransport)
	if err != nil {
		return errors.WithStack(err)
	}

	if actualLocation.ContentLength < 0 {
		return errors.Errorf("Invalid content length: %d", actualLocation.ContentLength)
	}

	partDownloadClient := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: httpTransport,
	}

	if actualLocation.StatusCode == http.StatusOK {
		actualLocation.computeParts(minPartSize)
		waitGroup.Add(len(actualLocation.Parts))
		for index, part := range actualLocation.Parts {
			go func(index int, part *Part) {
				defer recoverIfPanic(index)
				part.download(downloadContext, actualLocation.Location, index, partDownloadClient)
			}(index, part)
		}
	}

	go onCancelSignal(cancel)

	waitGroup.Wait()
	for _, part := range actualLocation.Parts {
		if part.isFail {
			cancel()
			break
		}
	}

	actualLocation.deleteUnnecessaryParts()
	err = actualLocation.concatenateParts(sha512)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func follow(initialUrl, userAgent, outFileName string, transport *http.Transport) (*ActualLocation, error) {
	totalWritten := int64(0)

	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: transport,
	}
	currentUrl := initialUrl
	redirectsFollowed := 0
	for {
		log.WithFields(log.Fields{
			"initialUrl": initialUrl,
			"currentUrl": currentUrl,
		}).Debug("computing effective URL")
		req, err := http.NewRequest(http.MethodGet, currentUrl, nil)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		req.Header.Set("User-Agent", userAgent)
		actualLocation, err := func() (*ActualLocation, error) {
			response, err := client.Do(req)
			if response != nil {
				defer response.Body.Close()
			}

			if err != nil {
				return nil, errors.WithStack(err)
			}

			if !isRedirect(response.StatusCode) {
				actualLocation := &ActualLocation{
					Location:          currentUrl,
					SuggestedFileName: outFileName,
					AcceptRanges:      response.Header.Get("Accept-Ranges"),
					StatusCode:        response.StatusCode,
					ContentLength:     response.ContentLength,
				}

				if response.StatusCode == http.StatusOK {
					var length string
					if totalWritten > 0 && actualLocation.AcceptRanges != "" {
						remaining := response.ContentLength - totalWritten
						length = fmt.Sprintf("%d, %d remaining", response.ContentLength, remaining)
					} else if response.ContentLength < 0 {
						length = "unknown"
					} else {
						length = fmt.Sprintf("%d", response.ContentLength)
					}
					log.WithFields(log.Fields{
						"length": length,
						"content-type":    response.Header.Get("Content-Type"),
						"url": initialUrl,
					}).Debug("downloading")

					if actualLocation.AcceptRanges == "" {
						log.Warn("server doesn't support ranges")
					}
				}
				return actualLocation, nil
			}

			loc, err := response.Location()
			if err != nil {
				return nil, errors.WithStack(err)
			}
			currentUrl = loc.String()
			return nil, nil
		}()

		if err != nil {
			return nil, errors.WithStack(err)
		}
		if actualLocation != nil {
			return actualLocation, nil
		}

		redirectsFollowed++
		if redirectsFollowed > maxRedirects {
			return nil, errors.Errorf("maximum number of redirects (%d) followed", maxRedirects)
		}
	}
	return nil, nil
}

func onCancelSignal(cancel context.CancelFunc) {
	defer cancel()
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	sig := <-signals
	fmt.Println()
	log.Infof("%v: canceling...\n", sig)
}

func isRedirect(status int) bool {
	return status > 299 && status < 400
}

type temporary interface {
	Temporary() bool
}

func isTemporary(err error) bool {
	te, ok := err.(temporary)
	return ok && te.Temporary()
}

func proxyFromEnvironmentAndNpm(req *http.Request) (*url.URL, error) {
	result, err := http.ProxyFromEnvironment(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if result != nil {
		return result, nil
	}

	urlString, err := proxyFromNpm()
	if err != nil {
		log.WithError(err).Error("cannot detect npm proxy")
		return nil, nil
	}

	if len(urlString) == 0 {
		return nil, nil
	}

	parsed, err := url.Parse(urlString)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return parsed, nil
}

func proxyFromNpm() (string, error) {
	userHomeDir, err := homedir.Dir()
	if err != nil {
		return "", errors.WithStack(err)
	}

	ini := goini.New()
	err = ini.ParseFile(filepath.Join(userHomeDir, ".npmrc"))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", errors.WithStack(err)
	}

	v, ok := ini.Get("https-proxy")
	if !ok {
		v, _ = ini.Get("proxy")
	}
	return v, nil
}
