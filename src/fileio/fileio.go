package fileio

import (
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
)

func ReadLinesFromFile(path string) ([]string, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "reading file %s", path)
	}

	asString := string(bytes[:])

	return strings.Split(asString, "\n"), nil
}
