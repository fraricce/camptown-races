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

var horses []horse
var horseNames string
var total int

func main() {

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	total = 0
	words := flag.Int("words", 1, "The number of words in the pet name")
	separator := flag.String("separator", " ", "The separator between words in the pet name")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 5; i++ {

		temp := petname.Generate(*words, *separator)
		horses = append(horses, horse{name: cam.ToCamel(temp), age: 2, strenght: 8})
		horseNames += horses[i].name + "\n"
	}

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", '1', gocui.ModNone, startRace); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)

	}
}

func layout(g *gocui.Gui) error {
	//maxX, maxY := g.Size()
	if v, err := g.SetView("hello", 0, 0, 30, 10); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, horseNames)
		v.Title = "Horses"
	}

	if v, err := g.SetView("hello2", 31, 0, 55, 10); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, "1. Start the raceeeesss")
		v.Title = "Commands"
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func startRace(g *gocui.Gui, v *gocui.View) error {

	for i := 0; i < 10; i++ {
		g.Update(func(g *gocui.Gui) error {
			v, err := g.View("hello")
			if err != nil {
				return err
			}
			fmt.Fprint(v, total)
			fmt.Println(total)
			total++
			time.Sleep(time.Duration(1) * time.Second)
			return nil
		})
	}
	return nil
}
