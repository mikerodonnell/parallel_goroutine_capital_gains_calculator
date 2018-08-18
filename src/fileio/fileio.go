package fileio

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"strings"
)

func ReadCSVDataRows(path string) ([][]string, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "trying to read file %s", path)
	}

	asString := string(bytes[:])

	allLines := strings.Split(asString, "\n")

	dataTokens := make([][]string, len(allLines)-1)
	for index, line := range allLines {
		// skip header row
		if index == 0 {
			continue
		}

		dataTokens[index-1] = strings.Split(line, ",")
	}

	return dataTokens, nil
}
