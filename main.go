package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/encoding"
	"github.com/gdamore/tcell/v2/views"
	"github.com/mattn/go-runewidth"
)

//const version = "0.0.1"

//var myView views.View

type position struct {
	pos    int
	minPos int
	maxPos int
}

func (p *position) scrollDown(v *views.ViewPort) {
	_, hh := v.Size()
	if p.pos < p.maxPos {
		p.pos++
		if p.pos > hh-3 {
			v.ScrollDown(1)
		}
	}
}

func (p *position) scrollUp(v *views.ViewPort) {
	//_, hh := v.Size()
	_, y1, _, _ := v.GetVisible()
	//_, yy1, _, _ := v.GetPhysical()
	if p.pos > p.minPos {
		p.pos--
		if p.pos == y1+2 {
			v.ScrollUp(1) // v.viewy
		}
	}

}

type Session struct { // NOTE why string and not a string?
	name string
}

func (s *Session) Name() string { return s.name }

// ListBox is a scrollable window listing contents.
type ListBox struct {
	//*Window
	view  views.View
	style tcell.Style
	//list   []Drawer
	list   []rune
	pos    position // index of a current content in the list
	offset int      // index of a top position to display
	//title  string
	lower  int
	column int
}

// NewListBox creates a new list box specified coordinates and sizes.
func NewListBox() *ListBox {
	return &ListBox{
		//view:  v,
		style: tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset),
		//Window: NewWindow(x, y, width, height),
		//list:   []Drawer{},
		list:   []rune{},
		pos:    position{pos: 0, minPos: 0, maxPos: 0},
		offset: 0,
		//title:  title,
		lower:  0,
		column: 1,
	}
}

func emitStr(v views.View, x, y int, style tcell.Style, str string) {
	for _, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		v.SetContent(x, y, c, comb, style)
		x += w
	}
}

func (l *ListBox) drawContent() {
	style := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset).Reverse(false)
	v := l.view
	for i, s := range l.list {
		//fmt.Printf("%s\n", s.name)
		if i == l.pos.pos {
			style = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset).Reverse(true)
		} else {
			style = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset).Reverse(false)
		}
		//emitStr(v, 0, i, style, s)
		v.SetContent(0, i, s, nil, style)
	}
}

// SetView sets the View object used for the text bar.
func (l *ListBox) SetView(view views.View) {
	l.view = view
}

func (l *ListBox) MakeSessionList() {
	//for i := 0; i < 100; i++ {
	for _, s := range "012345678901234567890123456789012345678901234567890123456789012345678901234567890" {
		l.list = append(l.list, s)
	}
	l.pos.maxPos = len(l.list) - 1 // FIXME
}

func (l *ListBox) Draw() {
	v := l.view
	if v == nil {
		fmt.Println("no v")
		return
	}
	// clear TODO
	l.drawContent()
}

type App struct {
	quitq  chan struct{}
	screen tcell.Screen
	eventq chan tcell.Event
	lview  *views.ViewPort
	rview  *views.ViewPort
	l2view *views.ViewPort
	//bar    *views.TextBar
	text *views.Text //Area
	box  *ListBox
	//box2 *ListBox

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

	//var myView views.View
	w, h := a.screen.Size()

	a.lview = views.NewViewPort(a.screen, 0, 0, 5, 3)
	a.l2view = views.NewViewPort(a.screen, 0, h/2+1, 5, 3)
	a.rview = views.NewViewPort(a.screen, w/2, 0, 10, 10)

	a.box = NewListBox()
	a.box.SetView(a.lview)
	a.box.MakeSessionList()

	a.text = views.NewText() //Area()
	a.text.SetView(a.l2view)

	a.quitq = make(chan struct{})
	a.eventq = make(chan tcell.Event)

	return nil

}

func (a *App) Quit() {
	close(a.quitq)
	a.screen.Fini()
	os.Exit(0)
}

func drawNumbers(v, v2 *views.ViewPort, pos position) {
	x1, y1, x2, y2 := v.GetVisible()
	xx1, yy1, xx2, yy2 := v.GetPhysical()

	ax1 := fmt.Sprintf("%d", x1)
	ay1 := fmt.Sprintf("%d", y1)
	ax2 := fmt.Sprintf("%d", x2)
	ay2 := fmt.Sprintf("%d", y2)

	axx1 := fmt.Sprintf("%d", xx1)
	ayy1 := fmt.Sprintf("%d", yy1)
	axx2 := fmt.Sprintf("%d", xx2)
	ayy2 := fmt.Sprintf("%d", yy2)

	pos1 := fmt.Sprintf("%d", pos.pos)
	posmax := fmt.Sprintf("%d", pos.maxPos)

	style := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	numberStrings := []string{"GetVisible:", ax1, ay1, ax2, ay2, " ", "GetPhysical:", axx1, ayy1, axx2, ayy2, " ", "pos, maxpos:", pos1, posmax}

	for i, s := range numberStrings {
		emitStr(v2, 0, i, style, s)
	}

}

func (a *App) Draw() {
	a.Lock()         // NOTE needed?
	a.screen.Clear() // widget.clear? TODO

	w, h := a.screen.Size()
	a.lview.SetSize(w/2, h/2)
	a.rview.SetSize(w/2, h)

	a.l2view.SetSize(w/2, h)

	//theText := "testing\nmore testing\nand even more\n"
	//a.text.SetText(theText)
	//a.text.SetStyleAt(4, style) // rune index, style
	//a.text.Draw()
	//a.lview.MakeVisible(0, h/2)
	//func (v *ViewPort) SetContentSize(width, height int, locked bool)
	//a.lview.SetSize(w/2, h/2)
	a.box.Draw()
	//a.box2.Draw()
	drawNumbers(a.lview, a.rview, a.box.pos)

	//a.box2.Draw()
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
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyEscape {
			a.Quit()
			return true
		} else if ev.Key() == tcell.KeyDown {
			a.box.pos.scrollDown(a.lview)
		} else if ev.Key() == tcell.KeyUp {
			a.box.pos.scrollUp(a.lview)
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
