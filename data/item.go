package data

import (
	ui "github.com/gizak/termui/v3"
	"github.com/sqshq/sampler/config"
	"os"
	"os/exec"
	"time"
)

const errorThreshold = 10

type Item struct {
	label           string
	sampleScript    string
	initScript      *string
	transformScript *string
	color           *ui.Color
	rateMs          int
	pty             bool
	basicShell      *BasicInteractiveShell
	ptyShell        *PtyInteractiveShell
}

func NewItems(cfgs []config.Item, rateMs int) []*Item {

	items := make([]*Item, 0)

	for _, i := range cfgs {
		item := &Item{
			label:           *i.Label,
			sampleScript:    *i.SampleScript,
			initScript:      i.InitScript,
			transformScript: i.TransformScript,
			color:           i.Color,
			rateMs:          rateMs,
			pty:             *i.Pty,
		}
		items = append(items, item)
	}
	return items
}

func (i *Item) nextValue(variables []string) (string, error) {

	if i.initScript != nil && i.basicShell == nil && i.ptyShell == nil {
		err := i.initInteractiveShell(variables)
		if err != nil {
			return "", err
		}
	}

	if i.basicShell != nil {
		return i.basicShell.execute()
	} else if i.ptyShell != nil {
		return i.ptyShell.execute()
	} else {
		return i.execute(variables, i.sampleScript)
	}
}

func (i *Item) execute(variables []string, script string) (string, error) {

	cmd := exec.Command("sh", "-c", script)
	enrichEnvVariables(cmd, variables)

	output, err := cmd.Output()

	if err != nil {
		return "", err
	}

	return string(output), nil
}

func (i *Item) initInteractiveShell(v []string) error {
	timeout := time.Duration(i.rateMs) * time.Millisecond * 3 / 4
	if i.pty {
		i.ptyShell = &PtyInteractiveShell{item: i, variables: v, timeout: timeout}
		return i.ptyShell.init()
	} else {
		i.basicShell = &BasicInteractiveShell{item: i, variables: v, timeout: timeout}
		return i.basicShell.init()
	}
}

func (i *Item) transform(sample string) (string, error) {

	if i.transformScript != nil && len(sample) > 0 {
		return i.execute([]string{"sample=" + sample}, *i.transformScript)
	}

	return sample, nil
}

func enrichEnvVariables(cmd *exec.Cmd, variables []string) {
	cmd.Env = os.Environ()
	for _, variable := range variables {
		cmd.Env = append(cmd.Env, variable)
	}
}
