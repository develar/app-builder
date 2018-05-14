package download

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"path/filepath"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/apex/log"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"github.com/dustin/go-humanize"
)

//noinspection SpellCheckingInspection
const localCert = `-----BEGIN CERTIFICATE-----
MIIBiDCCAS+gAwIBAgIRAPHSzTRLcN2nElhQdaRP47IwCgYIKoZIzj0EAwIwJDEi
MCAGA1UEAxMZZWxlY3Ryb24uYnVpbGQubG9jYWwgcm9vdDAeFw0xNzExMTMxNzI4
NDFaFw0yNzExMTExNzI4NDFaMCQxIjAgBgNVBAMTGWVsZWN0cm9uLmJ1aWxkLmxv
Y2FsIHJvb3QwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAQVyduuCT2acuk2QH06
yal/b6O7eTTpOHk3Ucjc+ZZta2vC2+c1IKcSAwimKbTbK+nRxWWJl9ZYx9RTwbRf
QjD6o0IwQDAOBgNVHQ8BAf8EBAMCAQYwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4E
FgQUlm08vBe4CUNAOTQN5Z1RNTfJjjYwCgYIKoZIzj0EAwIDRwAwRAIgMXlT6YM8
4pQtnhUjijVMz+NlcYafS1CEbNBMaWhP87YCIGXUmu7ON9hRLanXzBNBlrtTQG+i
l/NT6REwZA64/lNy
-----END CERTIFICATE-----
`

//noinspection SpellCheckingInspection
const producitonCert = `-----BEGIN CERTIFICATE-----
MIIBfDCCASKgAwIBAgIRALPgq5u/CdmbIQpZHwbem+EwCgYIKoZIzj0EAwIwHjEc
MBoGA1UEAxMTZWxlY3Ryb24uYnVpbGQgcm9vdDAeFw0xNzExMTMxNzI4NDFaFw0x
ODA1MTUwNTI4NDFaMBkxFzAVBgNVBAMTDmVsZWN0cm9uLmJ1aWxkMFkwEwYHKoZI
zj0CAQYIKoZIzj0DAQcDQgAE/OpHedt85s+Hc3l2XYLq7HcIEjgFGyOkEnzOgC6s
c3g7rEQA8Mwu+95dvyBnF1pUl64xMwIis+yI0GhmwqsI9qNGMEQwEwYDVR0lBAww
CgYIKwYBBQUHAwEwDAYDVR0TAQH/BAIwADAfBgNVHSMEGDAWgBToOryQru1CWu3O
QhitiJNkakcSnTAKBggqhkjOPQQDAgNIADBFAiBTd9EjhmHWOSSZmHkZUKpZbi+/
RD/JjoycWkzXSsz0qQIhAI0VoY/BSrTOlPaZVROB5v8g4b+pcQcKhI2h5F27xN2j
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIBfTCCASOgAwIBAgIRAIdieK1+3C4abgOvQ7pVVqAwCgYIKoZIzj0EAwIwHjEc
MBoGA1UEAxMTZWxlY3Ryb24uYnVpbGQgcm9vdDAeFw0xNzExMTMxNzI4NDFaFw0x
ODExMTMxNzI4NDFaMB4xHDAaBgNVBAMTE2VsZWN0cm9uLmJ1aWxkIHJvb3QwWTAT
BgcqhkjOPQIBBggqhkjOPQMBBwNCAAR+4b6twzizN/z27yvwrCV5kinGUrfo+W7n
L/l28ErscNe1BDSyh/IYrnMWb1rDMSLGhvkgI9Cfex1whNPHR101o0IwQDAOBgNV
HQ8BAf8EBAMCAQYwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQU6Dq8kK7tQlrt
zkIYrYiTZGpHEp0wCgYIKoZIzj0EAwIDSAAwRQIgKSfjAQbYlY/S1wMLUi84r8QN
hhMnUwsOmlDan0xPalICIQDLIAXAIyArVtH38a4aizvhH8YeXrxzpJh3U8RolBZF
SA==
-----END CERTIFICATE-----
`

func ConfigureDownloadResolvedFilesCommand(app *kingpin.Application) {
	command := app.Command("download-resolved-files", "Download artifacts from electron-build-service.")
	files := command.Flag("file", "The file path.").Short('f').Required().Strings()
	sizes := command.Flag("size", "The file size.").Short('s').Required().Int64List()
	baseUrl := command.Flag("base-url", "The base URL.").Required().String()
	outDir := command.Flag("out", "The output directory.").Required().String()

	command.Action(func(context *kingpin.ParseContext) error {
		downloader := NewDownloader()

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(getCaCerts())

		downloader.transport.TLSClientConfig = &tls.Config{
			ServerName: "electron.build.local",
			RootCAs:    caCertPool,
		}

		for index, file := range *files {
			start := time.Now()
			size := (*sizes)[index]
			location := NewResolvedLocation(*baseUrl+"/"+file, size, filepath.Join(*outDir, file), true)
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
	})
}

func getCaCerts() []byte {
	if util.IsEnvTrue("USE_ELECTRON_BUILD_SERVICE_LOCAL_CA") {
		return []byte(localCert)
	} else {
		return []byte(producitonCert)
	}
}
