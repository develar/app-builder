package util

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/develar/errors"
	"github.com/mitchellh/go-homedir"
	"github.com/zieckey/goini"
)

func ProxyFromEnvironmentAndNpm(req *http.Request) (*url.URL, error) {
	if os.Getenv("NO_PROXY") == "*" {
		return nil, nil
	}

	result, err := http.ProxyFromEnvironment(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if result != nil {
		return result, nil
	}

	result, err = proxyFromNpm()
	if err != nil {
		log.WithError(err).Error("cannot detect npm proxy")
		return nil, nil
	}
	return result, nil
}

func proxyFromNpm() (*url.URL, error) {
	userHomeDir, err := homedir.Dir()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	ini := goini.New()
	err = ini.ParseFile(filepath.Join(userHomeDir, ".npmrc"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.WithStack(err)
	}

	v, ok := ini.Get("https-proxy")
	if !ok {
		v, _ = ini.Get("proxy")
	}

	if len(v) == 0 || v == "false" || v == "true" {
		return nil, nil
	}

	parsed, err := url.Parse(v)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return parsed, nil
}
