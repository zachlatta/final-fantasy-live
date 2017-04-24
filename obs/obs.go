package obs

import (
	"os/exec"
)

type Obs struct {
	StreamUrl string
	StreamKey string

	NextButtonPressPath   string
	VoteBreakdownPath     string
	MostRecentPressesPath string
	TotalPressesPath      string
	TotalUptimePath       string

	buttonPressCount  int
	mostRecentPresses []string
}

func New(streamUrl, streamKey string) Obs {
	return Obs{
		StreamUrl: streamUrl,
		StreamKey: streamKey,
	}
}

func (o *Obs) Start() error {
	if err := o.setup(); err != nil {
		return err
	}

	cmd := exec.Command("obs", "--profile", "main", "--startstreaming")

	if err := cmd.Run(); err != nil {
		return err
	}

	if err := o.cleanup(); err != nil {
		return err
	}

	return nil
}
