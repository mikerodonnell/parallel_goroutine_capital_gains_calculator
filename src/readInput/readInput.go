package readInput

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func ReadStringInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter text: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", errors.Wrapf(err, "reading string input")
	}

	return strings.TrimSpace(input), nil
}

func ReadNumberInput() (int64, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter a number: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return 0, errors.Wrapf(err, "reading numeric input")
	}

	asString := strings.TrimSpace(input)

	intValue, err := strconv.Atoi(asString)
	if err != nil {
		return 0, errors.Wrapf(err, "casting string to int")
	}

	return int64(intValue), nil
}
