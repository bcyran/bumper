package pack

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

const srcinfoSeparator = " = "

var requiredFields = []string{"pkgbase", "pkgname", "pkgver", "pkgrel", "url"}

type rawSrcinfo map[string][]string

var ErrInvalidSrcinfo = errors.New("invalid .SRCINFO")

type Version string

func (v Version) GetVersionStr() string {
	return string(v)
}

type FullVersion struct {
	Pkgver Version
	Pkgrel string
}

type Srcinfo struct {
	Pkgbase string
	URL     string
	Source  []string
	*FullVersion
}

// ParseSrcinfo creates Srcinfo struct from .SRCINFO file at given path.
func ParseSrcinfo(path string) (*Srcinfo, error) {
	rawInfo, err := rawParseSrcinfo(path)
	if err != nil {
		return &Srcinfo{}, nil
	}

	for _, fieldName := range requiredFields {
		if len(rawInfo[fieldName]) != 1 {
			return &Srcinfo{}, fmt.Errorf("%w: %s missing/invalid '%s' value", ErrInvalidSrcinfo, path, fieldName)
		}
	}

	srcinfo := Srcinfo{
		Pkgbase: rawInfo["pkgbase"][0],
		URL:     rawInfo["url"][0],
		FullVersion: &FullVersion{
			Pkgver: Version(rawInfo["pkgver"][0]),
			Pkgrel: rawInfo["pkgrel"][0],
		},
	}

	if sourceValues, hasSourceValues := rawInfo["source"]; hasSourceValues {
		srcinfo.Source = sourceValues
	}

	return &srcinfo, nil
}

func rawParseSrcinfo(path string) (rawSrcinfo, error) {
	rawInfo := make(rawSrcinfo)

	file, err := os.Open(path)
	if err != nil {
		return rawInfo, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fieldName, fieldValue, separatorFound := strings.Cut(line, srcinfoSeparator)
		if separatorFound {
			fieldName, fieldValue = strings.TrimSpace(fieldName), strings.TrimSpace(fieldValue)
			rawInfo[fieldName] = append(rawInfo[fieldName], fieldValue)
		}
	}

	return rawInfo, nil
}
