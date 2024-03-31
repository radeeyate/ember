package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var (
	isHeating    bool
	wg           sync.WaitGroup
	stopWorkChan chan struct{}
)

func getCPUTemp() string {
	out, err := os.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	if err != nil {
		fmt.Println(err)
		return "Unknown error"
	}

	temp := strings.Replace(string(out), "\n", "", -1)

	tempInt, err := strconv.Atoi(temp)
	if err != nil {
		fmt.Println(err)
		return "Unknown error"
	}

	return fmt.Sprintf("%.1f°C", float64(tempInt)/1000.0)
}

func getDeviceInfo() string {
	if runtime.GOOS == "linux" {
		return "CPU Cores: " + fmt.Sprint(runtime.NumCPU()) +
			"\nArchitecture: " + runtime.GOARCH +
			"\nOS: " + cases.Title(language.English).String(string(runtime.GOOS)) +
			"\nGo Version: " + runtime.Version() +
			"\nCPU Temp: " + getCPUTemp()
	}
	return "CPU Cores: " + fmt.Sprint(runtime.NumCPU()) +
		"\nArchitecture: " + runtime.GOARCH +
		"\nOS: " + cases.Title(language.English).String(string(runtime.GOOS)) +
		"\nGo Version: " + runtime.Version()
}

func doWork() {
	var t float64 = math.MaxFloat64
	for {
		select {
		case <-stopWorkChan:
			return
		default:
			t /= 2.0
			if t > math.SmallestNonzeroFloat32 {
				t = math.MaxFloat64
			}
		}
	}
}

func main() {
	getCPUTemp()
	runtime.GOMAXPROCS(runtime.NumCPU())

	a := app.New()
	w := a.NewWindow("Hand Warmer")

	howItWorksContent := widget.NewLabel(`This app does computationally expensive work by continuously dividing ` +
		`a very large number (math.MaxFloat64) by 2.0. ` +
		`When the number gets too small (smaller than the smallest non-zero float32), ` +
		`it's reset back to the largest number. ` +
		`This loop keeps the CPU busy and therefore heats up your device.`)

	howItWorksContent.Wrapping = fyne.TextWrapWord
	howItWorksItem := widget.NewAccordionItem("How it Works", howItWorksContent)

	deviceInfoAccordionItem := widget.NewAccordionItem("Device Info", widget.NewLabel(getDeviceInfo()))

	label := widget.NewLabel("Hand Warmer Off")
	label.Wrapping = fyne.TextWrapWord

	button := widget.NewButton("Toggle On/Off Hand Warmer", func() {
		if isHeating {
			label.SetText("Cooling down...")
			isHeating = false
			close(stopWorkChan)
		} else {
			label.SetText("Heating up...")
			isHeating = true
			wg.Add(runtime.NumCPU())
			stopWorkChan = make(chan struct{})
			for i := 0; i < runtime.NumCPU(); i++ {
				go doWork()
			}
		}
	})

	aboutAccordion := widget.NewAccordion(howItWorksItem)
	devInfoAccordion := widget.NewAccordion(deviceInfoAccordionItem)
	w.SetContent(container.NewVBox(label, button, aboutAccordion, devInfoAccordion))
	w.ShowAndRun()
	wg.Wait()
}
