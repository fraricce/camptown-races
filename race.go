package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jroimartin/gocui"
)

var (
	done = make(chan struct{})
	ctr  = 0
)

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {
	if v, err := g.SetView("ctr", 2, 2, 12, 4); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, "0")
	}
	return nil
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 's', gocui.ModNone, start); err != nil {
		return err
	}
	return nil
}

func start(g *gocui.Gui, v *gocui.View) error {
	go counter(g)
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	close(done)
	return gocui.ErrQuit
}

func counter(g *gocui.Gui) {

	for {
		select {
		case <-done:
			return
		case <-time.After(500 * time.Millisecond):
			n := ctr
			ctr++

			g.Update(func(g *gocui.Gui) error {
				v, err := g.View("ctr")
				if err != nil {
					return err
				}
				v.Clear()
				fmt.Fprintln(v, n)
				return nil
			})

			if ctr == 10 {
				ctr = 0
				<-done
			}
		}
	}
}
