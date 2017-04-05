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

var errInvalidKey = errors.New("could not find the given key")

// Get returns the value of a sysctl flag or an error
func Get(name string) (string, error) {
	path := path.Join(sysctlDir, strings.Replace(name, ".", "/", -1))
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", errInvalidKey
	}
	return strings.TrimSpace(string(data)), nil
}

// Set sets a sysctl flag with given value, or return an error
func Set(name string, value string) error {
	path := path.Join(sysctlDir, strings.Replace(name, ".", "/", -1))
	return ioutil.WriteFile(path, []byte(value), 0640)
}
