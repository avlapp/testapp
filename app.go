package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/encoding"
	"github.com/gdamore/tcell/views"
)

//const version = "0.0.1"

type App struct {
	quitq  chan struct{}
	screen tcell.Screen
	eventq chan tcell.Event
	lview  *views.ViewPort
	bar    *views.TextBar

	sync.Mutex
}

func (a *App) Init() error {
	encoding.Register()
	if screen, err := tcell.NewScreen(); err != nil {
		return err
	} else if err = screen.Init(); err != nil {
		return err
	} else {
		screen.SetStyle(tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset))
		a.screen = screen
	}

	a.lview = views.NewViewPort(a.screen, 0, 1, -1, -1)
	a.bar = views.NewTextBar()
	a.bar.SetView(a.lview)

	a.quitq = make(chan struct{})
	a.eventq = make(chan tcell.Event)

	return nil

}

func (a *App) Quit() {
	close(a.quitq)
	a.screen.Fini()
	os.Exit(0)
}

func (a *App) Draw() {
	a.Lock() // NOTE needed?
	sbwarn := tcell.StyleDefault.
		Background(tcell.ColorAqua).
		Foreground(tcell.ColorRed)

	a.bar.SetStyle(sbwarn)
	a.bar.SetLeft("test!", sbwarn)
	a.bar.Draw()
	a.screen.Show()
	a.Unlock() // NOTE needed?
}

func (a *App) Run() error {
	go a.EventPoller()

loop:
	for {
		a.Draw()
		select {
		case <-a.quitq:
			break loop
		case ev := <-a.eventq:
			a.HandleEvent(ev)
		}
	}
	return nil
}

func (a *App) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventResize:
		a.lview.Resize(0, 1, -1, -1)
		//g.level.HandleEvent(ev) // NOTE from proxima5
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyEscape {
			a.Quit()
			return true
		}
	}
	return true
}

func (a *App) EventPoller() {
	for {
		select {
		case <-a.quitq:
			return
		default:
		}
		ev := a.screen.PollEvent()
		if ev == nil {
			return
		}
		select {
		case <-a.quitq:
			return
		case a.eventq <- ev:
		}
	}
}

func main() {
	app := &App{}
	if err := app.Init(); err != nil {
		fmt.Printf("Failed to init app: %v\n", err)
		os.Exit(1)
	}
	if err := app.Run(); err != nil {
		fmt.Printf("Failed to run app: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("This was the app.\n")

}
