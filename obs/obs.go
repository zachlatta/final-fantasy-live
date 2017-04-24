package obs

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type Obs struct {
	StreamUrl string
	StreamKey string

	NextButtonPressPath   string
	MostRecentPressesPath string
	TotalPressesPath      string

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

func (o *Obs) UpdateNextButtonPress(secondsRemaining int) error {
	return ioutil.WriteFile(
		o.NextButtonPressPath,
		[]byte(fmt.Sprintf("Next button press in: %d seconds", secondsRemaining)),
		os.ModePerm,
	)
}

func (o *Obs) AddMostRecentPress(newPress string) error {
	o.mostRecentPresses = append([]string{strings.ToUpper(newPress)}, o.mostRecentPresses...)

	if len(o.mostRecentPresses) > 3 {
		o.mostRecentPresses = o.mostRecentPresses[0:3]
	}

	return o.updateMostRecentPresses()
}

func (o *Obs) updateMostRecentPresses() error {
	return ioutil.WriteFile(
		o.MostRecentPressesPath,
		[]byte(fmt.Sprintf("Most recent presses:\n%s", strings.Join(o.mostRecentPresses, ", "))),
		os.ModePerm,
	)
}

func (o *Obs) IncrementButtonPresses() error {
	o.buttonPressCount += 1

	return o.updateTotalPresses()
}

func (o *Obs) updateTotalPresses() error {
	return ioutil.WriteFile(
		o.TotalPressesPath,
		[]byte(fmt.Sprintf("Total presses: %d", o.buttonPressCount)),
		os.ModePerm,
	)
}
