package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	petname "github.com/dustinkirkland/golang-petname"
	cam "github.com/iancoleman/strcase"
	"github.com/jroimartin/gocui"
)

type horse struct {
	name     string
	age      int
	strenght int
}

var (
	done   = make(chan struct{})
	ctr    = 0
	horses []horse
)

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	words := flag.Int("words", 1, "The number of words in the pet name")
	separator := flag.String("separator", " ", "The separator between words in the pet name")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 5; i++ {
		temp := petname.Generate(*words, *separator)
		horses = append(horses, horse{name: cam.ToCamel(temp), age: 2, strenght: 8})
	}

	g.SetManagerFunc(layout)

	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func renderHorses(v *gocui.View) error {
	for i := 0; i < 5; i++ {
		h := PadRight(horses[i].name, " ", 12)
		fmt.Fprintln(v, PadRight(h, ".", 50))

	}
	fmt.Fprintln(v, ctr)
	return nil
}

func PadRight(str, pad string, lenght int) string {
	for {
		str += pad
		if len(str) > lenght {
			return str[0:lenght]
		}
	}
}

func layout(g *gocui.Gui) error {
	if v, err := g.SetView("hello", 0, 0, 55, 10); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		renderHorses(v)
	}

	if v, err := g.SetView("hello2", 56, 0, 70, 10); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, "s. Start the race")
		v.Title = "Commands"
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
			ctr++

			g.Update(func(g *gocui.Gui) error {
				v, err := g.View("hello")
				if err != nil {
					return err
				}
				v.Clear()
				renderHorses(v)
				return nil
			})

			if ctr == 10 {
				ctr = 0
				<-done
			}
		}
	}
}
