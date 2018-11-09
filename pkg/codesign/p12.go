package codesign

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/develar/errors"
	"github.com/json-iterator/go"
	"github.com/lotus-wu/go-pkcs12"
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
		return err
	}

	jsonWriter := jsoniter.NewStream(jsoniter.ConfigFastest, os.Stdout, 16*1024)

	_, certificates, err := pkcs12.DecodeAll(data, password)
	if err != nil {
		jsonWriter.WriteObjectStart()
		jsonWriter.WriteObjectField("error")
		jsonWriter.WriteString(err.Error())
		jsonWriter.WriteObjectEnd()
		return flushJsonWriterAndCloseOut(jsonWriter)
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

	jsonWriter.WriteObjectStart()

	jsonWriter.WriteObjectField("commonName")
	jsonWriter.WriteString(firstCert.Subject.CommonName)

	// DN
	jsonWriter.WriteMore()
	jsonWriter.WriteObjectField("bloodyMicrosoftSubjectDn")
	jsonWriter.WriteString(BloodyMsString(firstCert.Subject.ToRDNSequence()))

	jsonWriter.WriteObjectEnd()

	return flushJsonWriterAndCloseOut(jsonWriter)
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

func flushJsonWriterAndCloseOut(jsonWriter *jsoniter.Stream) error {
	err := jsonWriter.Flush()
	if err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(os.Stdout.Close())
}
