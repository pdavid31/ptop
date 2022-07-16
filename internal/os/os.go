package os

import (
	"fmt"
	"io/ioutil"
	"path"
	"regexp"
)

const (
	osRegexString = "PRETTY_NAME=\"(.*)\""
)

var (
	osRegex = regexp.MustCompile(osRegexString)

	osReleaseFile = path.Join("/", "etc", "os-release")
)

func GetOS() (string, error) {
	content, err := ioutil.ReadFile(osReleaseFile)
	if err != nil {
		return "", err
	}

	match := osRegex.FindStringSubmatch(string(content))
	if len(match) < 1 {
		return "", fmt.Errorf("Could not identify operating system")
	}

	return match[1], nil
}
