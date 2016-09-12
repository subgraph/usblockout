package sysctl

import (
	"errors"
	"io/ioutil"
	"path"
	"strings"
)

const (
	sysctlDir = "/proc/sys/"
)

var invalidKeyError = errors.New("could not find the given key")

func Get(name string) (string, error) {
	path := path.Join(sysctlDir, strings.Replace(name, ".", "/", -1))
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", invalidKeyError
	}
	return strings.TrimSpace(string(data)), nil
}

func Set(name string, value string) error {
	path := path.Join(sysctlDir, strings.Replace(name, ".", "/", -1))
	return ioutil.WriteFile(path, []byte(value), 0640)
}
