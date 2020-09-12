package main

import (
	"flag"
	"log"
	"math/rand"
	"time"

	petname "github.com/dustinkirkland/golang-petname"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	cam "github.com/iancoleman/strcase"
)

type horse struct {
	name     string
	age      int
	strenght int
}

func main() {

	horses := []horse{}
	words := flag.Int("words", 1, "The number of words in the pet name")
	separator := flag.String("separator", " ", "The separator between words in the pet name")

	for i := 0; i < 5; i++ {
		flag.Parse()
		rand.Seed(time.Now().UTC().UnixNano())
		temp := petname.Generate(*words, *separator)
		horses = append(horses, horse{name: cam.ToCamel(temp), age: 2, strenght: 8})
	}

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}

	defer ui.Close()

	p := widgets.NewParagraph()
	p.Title = "Horses for the next race"
	horseNames := ""
	for i := 0; i < 5; i++ {
		horseNames += horses[i].name + "\n"
	}
	p.Text = horseNames
	p.SetRect(0, 0, 50, 10)

	instructionP := widgets.NewParagraph()
	instructionP.Title = "Commands"
	instructionP.Text = "1. Start the race"
	instructionP.SetRect(0, 20, 50, 12)

	ui.Render(p, instructionP)

	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			break
		}
	}
}
