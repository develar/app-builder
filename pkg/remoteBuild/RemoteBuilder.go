package remoteBuild

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/develar/app-builder/pkg/download"
	"github.com/develar/app-builder/pkg/log"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"github.com/dustin/go-humanize"
	"github.com/json-iterator/go"
	"go.uber.org/zap"
)

func ConfigureBuildCommand(app *kingpin.Application) {
	command := app.Command("remote-build", "")
	filesToPack := command.Flag("file", "").Required().Strings()
	buildResourcesDir := command.Flag("build-resource-dir", "").String()
	request := command.Flag("request", "").Required().String()
	output := command.Flag("output", "").Required().String()
	command.Action(func(context *kingpin.ParseContext) error {
		decodedRequest, err := base64.StdEncoding.DecodeString(*request)
		if err != nil {
			return err
		}

		err = newRemoteBuilder().build(string(decodedRequest), *filesToPack, *output, *buildResourcesDir)
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

func (t *RemoteBuilder) build(buildRequest string, filesToPack []string, outDir string, buildResourceDir string) error {
	var err error
	t.endpoint, err = findBuildAgent(t.transport)
	if err != nil {
		return err
	}

	client := &http.Client{
		Transport: t.transport,
		Timeout:   30 * time.Minute,
	}

	response, err := t.upload(buildRequest, filesToPack, buildResourceDir, client)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	resultEvent, rawResult, err := readEvents(response)
	if err != nil {
		return err
	}

	if resultEvent == nil {
		_, _ = os.Stdout.Write(rawResult)
		return nil
	}

	err = t.downloadArtifacts(resultEvent, outDir)
	if err != nil {
		return err
	}

	_, _ = os.Stdout.Write(rawResult)

	log.Info("found build service useful? Please donate (https://donorbox.org/electron-build-service)")
	return nil
}

func readEvents(response *http.Response) (*Event, []byte, error) {
	reader := bufio.NewReader(response.Body)
	for {
		encodedEvent, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil, []byte("{\"error\": \"remote build unexpectedly exited\"}"), nil
			} else {
				return nil, nil, err
			}
		}

		// exclude last \n
		encodedEvent = encodedEvent[0 : len(encodedEvent)-1]

		if log.IsDebugEnabled() {
			log.Debug("remote builder event", zap.ByteString("event", encodedEvent))
		}

		var event Event
		err = jsoniter.ConfigFastest.Unmarshal(encodedEvent, &event)
		if err != nil {
			return nil, nil, err
		}

		switch {
		case event.Status != "":
			log.Info("remote building", zap.String("status", strings.TrimSuffix(event.Status, "\n")))
		case event.Error != "":
			return nil, encodedEvent, nil
		case event.Files != nil:
			return &event, encodedEvent, nil
		default:
			log.Warn("unknown builder event", zap.ByteString("event", encodedEvent))
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
		err := downloader.DownloadResolved(&location, "", file.File)
		if err != nil {
			return errors.WithStack(err)
		}

		log.Info("file downloaded",
			zap.String("file", file.File),
			zap.String("size", humanize.Bytes(uint64(size))),
			zap.Duration("duration", time.Since(start).Round(time.Millisecond)),
		)
	}
	return nil
}

type Event struct {
	Status    string `json:"status"`
	Error     string `json:"error"`
	BaseUrl   string `json:"baseUrl"`
	Files     []File `json:"files"`
	FileSizes []int  `json:"fileSizes"`
}

type File struct {
	File string `json:"file"`
}

// compress and upload in the same time, directly to remote without intermediate local file
func (t *RemoteBuilder) upload(buildRequest string, filesToPack []string, buildResourceDir string, client *http.Client) (*http.Response, error) {
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

	if buildResourceDir != "" {
		fileInfo, err := os.Stat(buildResourceDir)
		if err == nil && fileInfo.IsDir() {
			tarArgs = append(tarArgs, buildResourceDir)
		}
	}

	tarCommand := exec.Command(util.Get7zPath(), tarArgs...)

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
	if err != nil {
		return nil, err
	}

	url := t.endpoint + "/v2/build"
	req, err := http.NewRequest(http.MethodPost, url, compressOutput)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("x-build-request", buildRequest)
	// only for stats purpose, not required for build
	req.Header.Set("x-zstd-compression-level", zstdCompressionLevel)

	_ = util.StartPipedCommands(tarCommand, compressCommand)
	response, err := client.Do(req)

	if err == nil {
		if response.StatusCode != http.StatusOK {
			_ = response.Body.Close()
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

	log.Info("uploaded to remote builder", zap.Duration("duration", time.Since(startTime).Round(time.Millisecond)))
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
