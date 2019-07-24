package codesign

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/develar/app-builder/pkg/download"
	"github.com/develar/app-builder/pkg/log"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"github.com/develar/go-pkcs12"
	"github.com/json-iterator/go"
	"go.uber.org/zap"
)

func ConfigureCertificateInfoCommand(app *kingpin.Application) {
	command := app.Command("certificate-info", "Read information about code signing certificate")
	inFile := command.Flag("input", "input file").Short('i').Required().String()
	password := command.Flag("password", "password").Short('p').String()

	command.Action(func(context *kingpin.ParseContext) error {
		return readInfo(*inFile, *password)
	})
}

func readInfo(inFile string, password string) error {
	data, err := ioutil.ReadFile(inFile)
	if err != nil {
		if os.IsNotExist(err) {
			return writeError(err.Error())
		}
		return err
	}

	certificates, err := pkcs12.DecodeAllCerts(data, password)
	if err != nil {
		if err.Error() == "pkcs12: decryption password incorrect" {
			return writeError("password incorrect")
		}

		log.Warn("cannot decode PKCS 12 data using Go pure implementation, openssl will be used", zap.Error(err))
		certificates, err = readUsingOpenssl(inFile, password)
		if err != nil {
			if strings.Contains(err.Error(), "Mac verify error: invalid password?") {
				return writeError("password incorrect")
			}

			m := err.Error()
			if exitError, ok := errors.Cause(err).(*exec.ExitError); ok {
				m += "; error output:\n" + string(exitError.Stderr)
			}
			return writeError(m)
		}
	}

	if len(certificates) == 0 {
		return fmt.Errorf("no certificates")
	}

	var firstCert *x509.Certificate
certLoop:
	for _, cert := range certificates {
		for _, usage := range cert.ExtKeyUsage {
			if usage == x509.ExtKeyUsageCodeSigning {
				firstCert = cert
				break certLoop
			}
		}
	}

	if firstCert == nil {
		return fmt.Errorf("no certificates with ExtKeyUsageCodeSigning")
	}

	jsonWriter := jsoniter.NewStream(jsoniter.ConfigFastest, os.Stdout, 16*1024)
	jsonWriter.WriteObjectStart()

	util.WriteStringProperty("commonName", firstCert.Subject.CommonName, jsonWriter)

	// DN
	jsonWriter.WriteMore()
	util.WriteStringProperty("bloodyMicrosoftSubjectDn", BloodyMsString(firstCert.Subject.ToRDNSequence()), jsonWriter)

	jsonWriter.WriteObjectEnd()

	return util.FlushJsonWriterAndCloseOut(jsonWriter)
}

func writeError(error string) error {
	jsonWriter := jsoniter.NewStream(jsoniter.ConfigFastest, os.Stdout, 16*1024)
	jsonWriter.WriteObjectStart()
	util.WriteStringProperty("error", error, jsonWriter)
	jsonWriter.WriteObjectEnd()
	return util.FlushJsonWriterAndCloseOut(jsonWriter)
}

func readUsingOpenssl(inFile string, password string) ([]*x509.Certificate, error) {
	opensslPath := "openssl"
	if util.GetCurrentOs() == util.WINDOWS {
		vendor, err := download.DownloadWinCodeSign()
		if err != nil {
			return nil, err
		}

		opensslPath = filepath.Join(vendor, "openssl-ia32", "openssl.exe")
	}

	//noinspection SpellCheckingInspection
	pemData, err := util.Execute(exec.Command(opensslPath, "pkcs12", "-in", inFile, "-passin", "pass:"+password, "-nokeys"))
	if err != nil {
		return nil, err
	}

	var blocks []byte
	rest := pemData
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			log.Debug("PEM not parsed")
			break
		}

		blocks = append(blocks, block.Bytes...)
		if len(rest) == 0 {
			break
		}
	}


	result, err2 := x509.ParseCertificates(blocks)
	if err2 != nil {
		return nil, errors.WithStack(err2)
	}
	return result, nil
}

//noinspection SpellCheckingInspection
var attributeTypeNames = map[string]string{
	"2.5.4.6":  "C",
	"2.5.4.10": "O",
	"2.5.4.11": "OU",
	"2.5.4.3":  "CN",
	"2.5.4.5":  "SERIALNUMBER",
	"2.5.4.7":  "L",
	"2.5.4.8":  "ST",
	"2.5.4.9":  "STREET",
	"2.5.4.17": "POSTALCODE",
}

// *** MS uses "The RDN value has quotes" for AppX, see https://docs.microsoft.com/en-us/uwp/schemas/appxpackage/appxmanifestschema/element-identity
// standard escaping doesn't work and forbidden
func BloodyMsString(r pkix.RDNSequence) string {
	var s strings.Builder
	for i := 0; i < len(r); i++ {
		rdn := r[len(r)-1-i]
		if i > 0 {
			s.WriteRune(',')
		}
		for j, tv := range rdn {
			if j > 0 {
				s.WriteRune('+')
			}

			oidString := tv.Type.String()
			typeName, ok := attributeTypeNames[oidString]
			if !ok {
				derBytes, err := asn1.Marshal(tv.Value)
				if err == nil {
					s.WriteString(oidString)
					s.WriteString("=#")
					s.WriteString(hex.EncodeToString(derBytes))
					// no value escaping necessary
					continue
				}

				typeName = oidString
			}

			valueString := fmt.Sprint(tv.Value)
			escaped := make([]rune, 0, len(valueString))

			s.WriteString(typeName)
			s.WriteRune('=')

			isNeedToBeEscaped := false
			for _, c := range valueString {
				switch c {
				case ',', '+', '"', '\\', '<', '>', ';':
					isNeedToBeEscaped = true
				}

				if c == '"' {
					escaped = append(escaped, '"', c)
				} else {
					escaped = append(escaped, c)
				}
			}

			if isNeedToBeEscaped {
				s.WriteRune('"')
			}
			s.WriteString(string(escaped))
			if isNeedToBeEscaped {
				s.WriteRune('"')
			}
		}
	}
	return s.String()
}
