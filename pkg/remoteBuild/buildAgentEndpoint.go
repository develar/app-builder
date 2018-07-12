package remoteBuild

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/develar/app-builder/pkg/util"
	"github.com/json-iterator/go"
)

func findBuildAgent(transport *http.Transport) (string, error) {
	result := os.Getenv("BUILD_AGENT_HOST")
	if result != "" {
		log.WithField("host", result).Debug("build agent host is set explicitly")
		return addHttpsIfNeed(result), nil
	}

	routerUrl := addHttpsIfNeed(util.GetEnvOrDefault("BUILD_SERVICE_ROUTER_HOST", "https://service.electron.build"))
	// add random query param to prevent caching
	routerUrl += "/find-build-agent?no-cache=" + strconv.FormatInt(time.Now().Unix(), 32)

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	for attemptNumber := 0; ; attemptNumber++ {
		result, err := getBuildAgentEndpoint(client, routerUrl)
		if err != nil {
			if attemptNumber == 3 {
				return "", err
			}

			waitTime := 2 * (attemptNumber + 1)
			log.WithError(err).WithField("attempt", attemptNumber).WithField("waitTime", waitTime).Warn("cannot get, wait")
			time.Sleep(time.Duration(waitTime) * time.Second)
			continue
		}

		return result, nil
	}
}

func getBuildAgentEndpoint(client *http.Client, url string) (string, error) {
	response, err := client.Get(url)

	if response.Body != nil {
		defer response.Body.Close()
	}

	if err == nil {
		if response.StatusCode != http.StatusOK {
			err = fmt.Errorf("cannot get %s: http error %d", url, response.StatusCode)
		}
	}

	if err != nil {
		return "", err
	}

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	result := jsoniter.Get(bodyBytes, "endpoint")
	err = result.LastError()
	if err != nil {
		return "", err
	}

	return result.ToString(), nil
}

func addHttpsIfNeed(s string) string {
	if strings.HasPrefix(s, "http") {
		return s
	} else {
		return "https://" + s
	}
}
