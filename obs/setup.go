package obs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mitchellh/go-homedir"
	"github.com/zachlatta/nostalgic-rewind/util"
	"io/ioutil"
	"text/template"
)

// Run `go generate` to regenerate the configData.go file (must be done any time
// a file changes in ./config/).

//go:generate go-bindata -o configData.go -pkg obs config/...

const (
	customConfigPath      = "config/"
	backupConfigExtension = ".bak"
)

func (o *Obs) setup() error {
	if err := o.setupTmpFiles(); err != nil {
		return err
	}

	if err := o.setupConfig(); err != nil {
		return err
	}

	return nil
}

func (o *Obs) setupTmpFiles() (err error) {
	o.NextButtonPressPath, err = createTmp("next-btn-countdown")
	if err != nil {
		return err
	}

	o.VoteBreakdownPath, err = createTmp("vote-breakdown")
	if err != nil {
		return err
	}

	o.MostRecentPressesPath, err = createTmp("most-recent-presses")
	if err != nil {
		return err
	}

	o.TotalPressesPath, err = createTmp("total-presses")
	if err != nil {
		return err
	}

	o.TotalUptimePath, err = createTmp("total-uptime")
	if err != nil {
		return err
	}

	// Set default values
	if err := o.drawDefaults(); err != nil {
		return err
	}

	return nil
}

func createTmp(prefix string) (filename string, err error) {
	file, err := ioutil.TempFile("", prefix)
	if err != nil {
		return "", err
	}
	defer file.Close()

	return file.Name(), nil
}

func (o Obs) setupConfig() error {
	config, err := o.ConfigPath()
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

	files := map[string]string{
		"global.ini":                       "process",
		"basic/profiles/main/basic.ini":    "process",
		"basic/profiles/main/service.json": "process",
		"basic/scenes/Main.json":           "process",
		"assets/controller.png":            "copy",
		"assets/footer.png":                "copy",
	}

	for path, action := range files {
		data, err := Asset(filepath.Join(customConfigPath, path))
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

		switch action {
		case "process":
			tmpl, err := template.New(path).Parse(string(data))
			if err != nil {
				return err
			}

			if err := tmpl.Execute(file, o); err != nil {
				return err
			}
		case "copy":
			if _, err := file.Write(data); err != nil {
				return err
			}
		default:
			return errors.New(fmt.Sprintf("encountered unknown action while processing %s: %s", path, action))
		}
	}

	return nil
}

func (o *Obs) cleanup() error {
	if err := o.cleanupTmpFiles(); err != nil {
		return err
	}

	if err := o.cleanupConfig(); err != nil {
		return err
	}

	return nil
}

func (o *Obs) cleanupTmpFiles() error {
	if err := os.Remove(o.NextButtonPressPath); err != nil {
		return err
	}
	o.NextButtonPressPath = ""

	if err := os.Remove(o.VoteBreakdownPath); err != nil {
		return err
	}
	o.VoteBreakdownPath = ""

	if err := os.Remove(o.MostRecentPressesPath); err != nil {
		return err
	}
	o.MostRecentPressesPath = ""

	if err := os.Remove(o.TotalPressesPath); err != nil {
		return err
	}
	o.TotalPressesPath = ""

	if err := os.Remove(o.TotalUptimePath); err != nil {
		return err
	}
	o.TotalUptimePath = ""

	return nil
}

func (o Obs) cleanupConfig() error {
	config, err := o.ConfigPath()
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

func (o Obs) ConfigPath() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return homedir.Expand("~/Library/Application Support/obs-studio")
	case "linux":
		return homedir.Expand("~/.config/obs-studio")
	default:
		return "", errors.New("unsupported OS")
	}
}
