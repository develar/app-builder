package remoteBuild

import (
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/develar/app-builder/pkg/log"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"github.com/json-iterator/go"
	"go.uber.org/zap"
)

func findBuildAgent(transport http.RoundTripper) (string, error) {
	result := os.Getenv("BUILD_AGENT_HOST")
	if result != "" {
		log.Debug("build agent host is set explicitly", zap.String("host", result))
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
			log.Warn("cannot get, wait", zap.Error(err), zap.Int("attempt", attemptNumber), zap.Int("waitTime", waitTime))
			time.Sleep(time.Duration(waitTime) * time.Second)
			continue
		}

		return result, nil
	}
}

func getBuildAgentEndpoint(client *http.Client, url string) (string, error) {
	response, err := client.Get(url)
	if err != nil {
		return "", err
	}

	if response.Body != nil {
		defer util.Close(response.Body)
	}

	if response.StatusCode != http.StatusOK {
		return "", errors.Errorf("cannot get %s: http error %d", url, response.StatusCode)
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
