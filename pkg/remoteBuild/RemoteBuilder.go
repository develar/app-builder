package remoteBuild

import (
	"bufio"
	"encoding/base64"
	"fmt"
		"net/http"
	"os"
	"os/exec"
	"path/filepath"
		"strings"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/apex/log"
	"github.com/develar/app-builder/pkg/download"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"github.com/dustin/go-humanize"
	"github.com/json-iterator/go"
)

func ConfigureBuildCommand(app *kingpin.Application) {
	command := app.Command("remote-build", "")
	filesToPack := command.Flag("file", "").Required().Strings()
	request := command.Flag("request", "").Required().String()
	output := command.Flag("output", "").Required().String()
	command.Action(func(context *kingpin.ParseContext) error {
		decodedRequest, err := base64.StdEncoding.DecodeString(*request)
		if err != nil {
			return err
		}

		err = newRemoteBuilder().build(string(decodedRequest), *filesToPack, *output)
		if err != nil {
			return err
		}
		return nil
	})
}

type RemoteBuilder struct {
	endpoint string

	transport *http.Transport
}

func newRemoteBuilder() *RemoteBuilder {
	transport := &http.Transport{
		Proxy:           util.ProxyFromEnvironmentAndNpm,
		TLSClientConfig: getTls(),
	}
	return &RemoteBuilder{
		transport: transport,
	}
}

func (t *RemoteBuilder) build(buildRequest string, filesToPack []string, outDir string) error {
	var err error
	t.endpoint, err = findBuildAgent(t.transport)
	if err != nil {
		return err
	}

	client := &http.Client{
		Transport: t.transport,
		Timeout:   30 * time.Minute,
	}

	response, err := t.upload(buildRequest, filesToPack, client)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	resultEvent, rawResult, err := readEvents(response)
	if err != nil {
		return err
	}

	if resultEvent == nil {
		os.Stdout.Write(rawResult)
		return nil
	}

	err = t.downloadArtifacts(resultEvent, outDir)
	if err != nil {
		return err
	}

	os.Stdout.Write(rawResult)
	return nil
}

func readEvents(response *http.Response) (*Event, []byte, error) {
	reader := bufio.NewReader(response.Body)
	for {
		encodedEvent, err := reader.ReadBytes('\n')
		if err != nil {
			return nil, nil, err
		}

		// exclude last \n
		encodedEvent = encodedEvent[0:len(encodedEvent) - 1]

		if util.IsDebugEnabled() {
			log.WithField("event", string(encodedEvent)).Debug("remote builder event")
		}

		var event Event
		err = jsoniter.ConfigFastest.Unmarshal(encodedEvent, &event)
		if err != nil {
			return nil, nil, err
		}

		if event.Status != "" {
			log.WithField("status", event.Status).Info("remote building")
		} else if event.Error != "" {
			return nil, encodedEvent, nil
		} else if event.Files != nil {
			return &event, encodedEvent, nil
		} else {
			log.WithField("event", string(encodedEvent)).Warn("unknown builder event")
		}
	}
}

func (t *RemoteBuilder) downloadArtifacts(resultEvent *Event, outDir string) error {
	downloader := download.NewDownloaderWithTransport(t.transport)
	baseUrl := t.endpoint + resultEvent.BaseUrl
	for index, file := range resultEvent.Files {
		start := time.Now()
		size := (resultEvent.FileSizes)[index]
		location := download.NewResolvedLocation(baseUrl+"/"+file.File, int64(size), filepath.Join(outDir, file.File), true)
		err := downloader.DownloadResolved(&location, "")
		if err != nil {
			return errors.WithStack(err)
		}

		log.WithFields(&log.Fields{
			"file":     file,
			"size":     humanize.Bytes(uint64(size)),
			"duration": fmt.Sprintf("%v", time.Since(start).Round(time.Millisecond)),
		}).Info("file downloaded")
	}
	return nil
}

type Event struct {
	Status string `json:"status"`
	Error string `json:"error"`
	BaseUrl string `json:"baseUrl"`
	Files []File `json:"files"`
	FileSizes []int `json:"fileSizes"`
}

type File struct {
	File string `json:"file"`
}

// compress and upload in the same time, directly to remote without intermediate local file
func (t *RemoteBuilder) upload(buildRequest string, filesToPack []string, client *http.Client) (*http.Response, error) {
	zstd, err := download.GetZstd()
	if err != nil {
		return nil, err
	}

	zstdCompressionLevel := getZstdCompressionLevel(t.endpoint)

	//noinspection SpellCheckingInspection
	tarArgs := []string{"a", "dummy", "-ttar", "-so"}

	cwd, err := os.Getwd()
	if err == nil {
		for _, value := range filesToPack {
			d := value
			if !filepath.IsAbs(value) {
				d = filepath.Join(cwd, value)
			}
			tarArgs = append(tarArgs, filepath.Clean(d))
		}
	} else {
		tarArgs = append(tarArgs, filesToPack...)
	}

	tarCommand := exec.Command(util.GetEnvOrDefault("SZA_PATH", "7za"), tarArgs...)

	tarCommand.Stderr = os.Stderr
	tarOutput, err := tarCommand.StdoutPipe()
	if err != nil {
		return nil, err
	}

	compressCommand := exec.Command(zstd, "-"+zstdCompressionLevel, "--long", "-T0")
	compressCommand.Stderr = os.Stderr
	compressCommand.Stdin = tarOutput
	compressOutput, err := compressCommand.StdoutPipe()
	if err != nil {
		return nil, err
	}

	log.Info("compressing and uploading to remote builder")
	startTime := time.Now()
	err = util.StartPipedCommands(tarCommand, compressCommand)

	url := t.endpoint + "/v2/build"
	req, err := http.NewRequest(http.MethodPost, url, compressOutput)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("x-build-request", buildRequest)
	// only for stats purpose, not required for build
	req.Header.Set("x-zstd-compression-level", zstdCompressionLevel)

	util.StartPipedCommands(tarCommand, compressCommand)
	response, err := client.Do(req)

	if err == nil {
		if response.StatusCode != http.StatusOK {
			response.Body.Close()
			err = fmt.Errorf("cannot get %s: http error %d", url, response.StatusCode)
		}
	}

	if err != nil {
		return nil, err
	}

	err = util.WaitPipedCommand(tarCommand, compressCommand)
	if err != nil {
		response.Body.Close()
		return nil, err
	}

	log.WithFields(&log.Fields{
		"duration": fmt.Sprintf("%v", time.Since(startTime).Round(time.Millisecond)),
	}).Info("uploaded to remote builder")

	return response, nil
}

func getZstdCompressionLevel(endpoint string) string {
	result := os.Getenv("BUILD_SERVICE_ZSTD_COMPRESSION")
  if result != "" {
    return result
  }

  // 18 - 40s
  // 17 - 30s
  // 16 - 20s
  if strings.HasPrefix(endpoint, "https://127.0.0.1:") || strings.HasPrefix(endpoint, "https://localhost:") || strings.HasPrefix(endpoint, "[::1]:") {
  	return "3"
	} else {
		return "16"
	}
}