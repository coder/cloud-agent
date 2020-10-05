package config

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

var (
	// ConfigDir is the directory where agent config is stored.
	ConfigDir = "coder-cloud"
)

func dir() (string, error) {
	conf, err := os.UserConfigDir()
	if runtime.GOOS == "darwin" {
		// No one uses macOS's ~/Library for CLI apps...
		// Sigh.
		// See https://github.com/golang/go/issues/29960#issuecomment-499842130
		conf, err = xdgConfigDir()
	}
	if err != nil {
		return "", err
	}

	return filepath.Join(conf, ConfigDir), nil
}

// Copied directly from stdlib's os.UserConfigDir.
func xdgConfigDir() (string, error) {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		dir = os.Getenv("HOME")
		if dir == "" {
			return "", errors.New("neither $XDG_CONFIG_HOME nor $HOME are defined")
		}
		dir += "/.config"
	}
	return dir, nil
}

// open opens a file in the configuration directory,
// creating all intermediate directories.
func open(path string, flag int, mode os.FileMode) (*os.File, error) {
	dir, err := dir()
	if err != nil {
		return nil, err
	}

	path = filepath.Join(dir, path)

	err = os.MkdirAll(filepath.Dir(path), 0750)
	if err != nil {
		return nil, err
	}

	return os.OpenFile(path, flag, mode)
}

func write(path string, mode os.FileMode, dat []byte) error {
	fi, err := open(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, mode)
	if err != nil {
		return err
	}
	defer fi.Close()
	_, err = fi.Write(dat)
	return err
}

func read(path string) ([]byte, error) {
	fi, err := open(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer fi.Close()
	return ioutil.ReadAll(fi)
}

func rm(path string) error {
	dir, err := dir()
	if err != nil {
		return err
	}

	return os.Remove(filepath.Join(dir, path))
}
