package obs

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mitchellh/go-homedir"
	"github.com/zachlatta/nostalgic-rewind/util"
	"text/template"
)

// Run `go generate` to regenerate the configData.go file (must be done any time
// a file changes in ./config/).

//go:generate go-bindata -o configData.go -pkg obs config/...

const (
	customConfigPath      = "config/"
	backupConfigExtension = ".bak"
)

func (o Obs) setupConfig() error {
	config, err := configPath()
	if err != nil {
		return err
	}

	configExists, err := util.FileExists(config)
	if err != nil {
		return err
	}

	if configExists {
		if err := os.Rename(config, config+backupConfigExtension); err != nil {
			return err
		}
	}

	toProcess := []string{
		"global.ini",
		"basic/profiles/main/basic.ini",
		"basic/profiles/main/service.json",
		"basic/scenes/Untitled.json",
	}

	for _, path := range toProcess {
		data, err := Asset(filepath.Join(customConfigPath, path))
		if err != nil {
			return err
		}

		tmpl, err := template.New(path).Parse(string(data))
		if err != nil {
			return err
		}

		fullPath := filepath.Join(config, path)
		dir := filepath.Join(config, filepath.Dir(path))

		exists, err := util.FileExists(dir)
		if err != nil {
			return err
		}

		if !exists {
			os.MkdirAll(dir, os.ModePerm)
		}

		file, err := os.Create(fullPath)
		if err != nil {
			return err
		}
		defer file.Close()

		if err := tmpl.Execute(file, o); err != nil {
			return err
		}
	}

	return nil
}

func (o Obs) cleanupConfig() error {
	config, err := configPath()
	if err != nil {
		return err
	}

	if err := os.RemoveAll(config); err != nil {
		return err
	}

	backupConfigExists, err := util.FileExists(config + backupConfigExtension)
	if err != nil {
		return err
	}

	if backupConfigExists {
		if err := os.Rename(config+backupConfigExtension, config); err != nil {
			return err
		}
	}

	return nil
}

func configPath() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return homedir.Expand("~/Library/Application Support/obs-studio")
	case "linux":
		return homedir.Expand("~/.config/obs-studio")
	default:
		return "", errors.New("unsupported OS")
	}
}
