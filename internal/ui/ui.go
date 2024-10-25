package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

type App struct {
	fyne.App
}

func Setup() (*App, error) {
	var application App

	fyneApp := app.New()
	myWindow := fyneApp.NewWindow("Watchman")
	myWindow.Resize(fyne.NewSize(1000, 1000*0.66))
	myWindow.SetContent(widget.NewLabel("Hello"))
	myWindow.Show()

	application.App = fyneApp

	return &application, nil
}
