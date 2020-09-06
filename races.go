package main

import (
	"flag"
	"log"
	"math/rand"
	"time"

	petname "github.com/dustinkirkland/golang-petname"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

var (
	words     = flag.Int("words", 1, "The number of words in the pet name")
	separator = flag.String("separator", " ", "The separator between words in the pet name")
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	name := petname.Generate(*words, *separator)
	p := widgets.NewParagraph()
	p.Text = name
	p.TextStyle.Fg = ui.ColorBlue
	p.SetRect(0, 0, 20, 5)

	ui.Render(p)

	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			break
		}
	}
}
