package utils

import (
	"io/ioutil"
	"os"
)

func Stream(file string) ([]byte, error) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil, err
	}

	c, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return c, nil
}
