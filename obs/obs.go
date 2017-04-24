package obs

import (
	"os/exec"
)

type Obs struct {
	StreamUrl string
	StreamKey string
}

func New(streamUrl, streamKey string) Obs {
	return Obs{
		StreamUrl: streamUrl,
		StreamKey: streamKey,
	}
}

func (o Obs) Start() error {
	if err := o.setupConfig(); err != nil {
		return err
	}

	cmd := exec.Command("obs", "--profile", "main", "--startstreaming")

	if err := cmd.Run(); err != nil {
		return err
	}

	if err := o.cleanupConfig(); err != nil {
		return err
	}

	return nil
}
