package obs

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/paked/nes/nes"
	"github.com/zachlatta/nostalgic-rewind/util"
	"strconv"
	"time"
)

func (o *Obs) drawDefaults() error {
	if err := o.UpdateNextButtonPress(0); err != nil {
		return err
	}
	if err := o.UpdateVoteBreakdown(map[int]int{}); err != nil {
		return err
	}
	if err := o.updateMostRecentPresses(); err != nil {
		return err
	}
	if err := o.updateTotalPresses(); err != nil {
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

func (o *Obs) UpdateVoteBreakdown(breakdown map[int]int) error {
	str := fmt.Sprintf(`Current votes:

  UP: %s  DOWN: %s
LEFT: %s RIGHT: %s
   B: %s     A: %s
`,
		padCount(breakdown[nes.ButtonUp]), padCount(breakdown[nes.ButtonDown]),
		padCount(breakdown[nes.ButtonLeft]), padCount(breakdown[nes.ButtonRight]),
		padCount(breakdown[nes.ButtonB]), padCount(breakdown[nes.ButtonA]),
	)

	return ioutil.WriteFile(o.VoteBreakdownPath, []byte(str), os.ModePerm)
}

func padCount(num int) string {
	const length = 5
	const padder = " " // Character to pad with

	return util.RightPad(strconv.Itoa(num), padder, length)
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

func (o *Obs) UpdateTotalUptime(startTime, currentTime time.Time) error {
	delta := currentTime.Sub(startTime)
	days := int(delta.Hours()) / int(time.Hour*24)
	hours := int(delta.Hours()) - int(days*24)
	minutes := int(delta.Minutes()) - int(hours*60)
	seconds := int(delta.Seconds()) - int(minutes*60)

	str := fmt.Sprintf(
		"Total uptime: %sD, %sH, %sM, %sS",
		padTime(days),
		padTime(hours),
		padTime(minutes),
		padTime(seconds),
	)

	return ioutil.WriteFile(o.TotalUptimePath, []byte(str), os.ModePerm)
}

func padTime(num int) string {
	length := 2
	padChar := "0"

	return util.LeftPad(strconv.Itoa(num), padChar, length)
}
