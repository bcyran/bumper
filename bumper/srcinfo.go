package bumper

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

type FullVersion struct {
	Pkgver string
	Pkgrel string
}

type Srcinfo struct {
	Pkgbase string
	Url     string
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
			return &Srcinfo{}, fmt.Errorf("%w: missing/invalid '%s' value", ErrInvalidSrcinfo, fieldName)
		}
	}

	srcinfo := Srcinfo{
		Pkgbase: rawInfo["pkgbase"][0],
		Url:     rawInfo["url"][0],
		FullVersion: &FullVersion{
			Pkgver: rawInfo["pkgver"][0],
			Pkgrel: rawInfo["pkgrel"][0],
		},
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
		if strings.Contains(line, srcinfoSeparator) {
			parts := strings.SplitN(line, srcinfoSeparator, 2)
			fieldName, fieldValue := parts[0], parts[1]
			rawInfo[fieldName] = append(rawInfo[fieldName], fieldValue)
		}
	}

	return rawInfo, nil
}
