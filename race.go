package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	petname "github.com/dustinkirkland/golang-petname"
	cam "github.com/iancoleman/strcase"
	"github.com/jroimartin/gocui"
)

type horse struct {
	name     string
	age      int
	strenght int
	pos      int
	fallen   bool
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
		force := rand.Intn(9) + 1
		horses = append(horses, horse{name: cam.ToCamel(temp), age: 2, strenght: force, pos: 1, fallen: false})
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

	for i := 1; i <= 5; i++ {
		stride := rand.Intn(3-1) + 1
		if stride < 0 {
			stride = 0
		}
		h := strconv.Itoa(i) + ". " + PadRight(horses[i-1].name, " ", 9)

		len := ""

		for j := 0; j < horses[i-1].pos; j++ {
			len += "."
		}

		if horses[i-1].pos > 1 {
			for s := 0; s < stride; s++ {
				len += "."
			}
		}

		horses[i-1].pos++
		horses[i-1].pos += stride

		if !(horses[i-1].fallen) {
			h += len
		}

		fmt.Fprintln(v, h)
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

	if v, err := g.SetView("raceField", 0, 0, 79, 10); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Camptown Races"
		s := make([]string, 3)
		dice := rand.Intn(3)
		s[0] = " Stratford Racecourse"
		s[1] = " Wolverhampton Racecourse"
		s[2] = " Cheltenham Racecourse"

		fmt.Fprintln(v, "Welcome to"+s[dice]+"!")
		renderHorses(v)
	}

	if v, err := g.SetView("command", 0, 11, 22, 20); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, "s. Start the race")
		v.Title = "Commands"
	}

	if v, err := g.SetView("quotations", 23, 11, 79, 20); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, "s. Start the race")
		v.Title = "Quotations"
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
				v, err := g.View("raceField")
				if err != nil {
					return err
				}
				v.Clear()
				renderHorses(v)
				return nil
			})

			if ctr == 15 {
				ctr = 0
				<-done
			}
		}
	}
}
