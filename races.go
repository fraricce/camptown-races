package main

import (
	"flag"
	"fmt"
	// tl "github.com/JoelOtter/termloop"
	"github.com/dustinkirkland/golang-petname"
	"math/rand"
	"time"
)

var (
	words     = flag.Int("words", 1, "The number of words in the pet name")
	separator = flag.String("separator", " ", "The separator between words in the pet name")
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	fmt.Println(petname.Generate(*words, *separator))

	// game := tl.NewGame()
	// level := tl.NewBaseLevel(tl.Cell{
	// 	Bg: tl.ColorGreen,
	// 	Fg: tl.ColorBlack,
	// 	Ch: 'v',
	// })
	// level.AddEntity(tl.NewRectangle(0, 0, 10, 40, tl.ColorBlue))
	// game.Screen().SetLevel(level)
	// game.Start()
}
