package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Deichindianer/quick-ssm-state/internal/mainScreen"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"

	ui "github.com/gizak/termui/v3"
)

func main() {
	var err error

	if err = ui.Init(); err != nil {
		exit(1, err)
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		exit(1, err)
	}

	ssmClient := ssm.NewFromConfig(cfg)

	ms, err := mainScreen.New(ssmClient)
	if err != nil {
		exit(1, err)
	}
	UIBusyloop(ms)
}

func UIBusyloop(ms *mainScreen.MainScreen) {
	defer recoverPanic()
	ui.Render(ms.Grid)
	uiEvents := ui.PollEvents()
	var previousKey string
	selectedAssociation := ms.AssociationList.Rows[0]
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				exit(0, nil)
			case "<Down>", "j":
				ms.AssociationList.ScrollDown()
			case "<Up>", "k":
				ms.AssociationList.ScrollUp()
			case "<Home>":
				ms.AssociationList.ScrollTop()
			case "g":
				if previousKey == "g" {
					previousKey = ""
					ms.AssociationList.ScrollTop()
				}
			case "<End>", "G":
				ms.AssociationList.ScrollBottom()
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				ms.StatusBarChart.BarWidth = (payload.Width - 10) / 2 / 4
				ms.Grid.SetRect(0, 0, payload.Width, payload.Height)
				ui.Clear()
			case "r":
				if err := ms.AssociationList.Reload(); err != nil {
					exit(1, err)
				}
			case "<Enter>":
				selectedAssociation = ms.AssociationList.Rows[ms.AssociationList.SelectedRow]
				if err := ms.StatusBarChart.Reload(selectedAssociation); err != nil {
					exit(1, err)
				}
				if err := ms.TargetList.Reload(selectedAssociation); err != nil {
					exit(1, err)
				}
			}
			previousKey = e.ID
			ui.Render(ms.Grid)
		case <-ticker.C:
			if err := ms.StatusBarChart.Reload(selectedAssociation); err != nil {
				exit(1, err)
			}
			if err := ms.TargetList.Reload(selectedAssociation); err != nil {
				exit(1, err)
			}
			ui.Render(ms.Grid)
		}
	}
}

func recoverPanic() {
	r := recover()
	var err error
	if r != nil {
		switch x := r.(type) {
		case string:
			err = errors.New(x)
		case error:
			err = x
		default:
			err = errors.New("unknown panic")
		}
		exit(1, err)
	}
}

func exit(exitCode int, err error) {
	ui.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	os.Exit(exitCode)
}
