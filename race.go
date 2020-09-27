package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	petname "github.com/dustinkirkland/golang-petname"
	cam "github.com/iancoleman/strcase"
	"github.com/jroimartin/gocui"
)

type horse struct {
	name     string
	jockey   string
	age      int
	strenght int
	pos      int
	fallen   bool
	winner   bool
	finisher bool
	place    int
}

type placeInfo struct {
	raceCourse string
	city       string
	county     string
	country    string
	weather    int // 0 sunny, 1 rain
}

type raceInfo struct {
	name          string
	category      string // National Hunt
	branch        string // ex. hurdles
	lengthFurlong float32
}

var (
	done       = make(chan struct{})
	won        = -1
	ctr        = 0
	finishLine = 30
	horses     []horse
	place      placeInfo
	race       raceInfo
	comments   = make([]string, 0)
	arrivalIdx = 0
	words      = flag.Int("words", 1, "The number of words in the generated name")
	separator  = flag.String("separator", " ", "The separator between words in the generated name")
)

func initGame() {
	done = make(chan struct{})
	won = -1
	ctr = 0
	finishLine = 30
	horses = nil
	comments = make([]string, 0)
	arrivalIdx = 0
}

func main() {

	initGame()

	g, err := gocui.NewGui(gocui.OutputNormal)

	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	horses = generateHorses()
	place = generatePlace()
	race = generateRace()

	g.SetManagerFunc(layout)

	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func generateRace() raceInfo {
	return raceInfo{name: "Cathedral Stakes", category: "Flat", branch: "", lengthFurlong: 6}
}

func generatePlace() placeInfo {
	weather := 1
	weatherFactor := rand.Intn(10)
	if weatherFactor >= 5 {
		weather = 0
	} else {
		if weatherFactor >= 3 {
			weather = 1
		} else {
			weather = 2
		}
	}

	place := placeInfo{city: "Salisbury", raceCourse: "Salisbury Racecourse", county: "Wiltshire", country: "England", weather: weather}
	return place
}

func generateHorses() []horse {

	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 5; i++ {
		temp := petname.Generate(*words, *separator)
		force := rand.Intn(9) + 1
		year := rand.Intn(4) + 1
		horses = append(horses, horse{name: cam.ToCamel(temp), age: year, strenght: force, pos: 1, fallen: false, winner: false, finisher: false, place: 0})
	}

	return horses
}

func moveHorses() {
	for i := 0; i < 5; i++ {
		strideFactor := horses[i].strenght + 1
		stride := rand.Intn(strideFactor-1) + 1

		if stride >= 8 {
			stride = 3
		} else if stride < 8 && stride >= 5 {
			stride = 2
		} else if stride < 4 && stride > 1 {
			stride = 1
		}

		if !(horses[i].fallen) && horses[i].pos <= finishLine {
			//horses[i-1].pos++
			horses[i].pos += stride

			fallFactor := 0
			if place.weather == 0 {
				fallFactor = 80
			} else if place.weather == 1 {
				fallFactor = 60
			} else if place.weather == 2 {
				fallFactor = 35
			}

			fall := rand.Intn(fallFactor) + 1
			if fall == 2 {
				horses[i].fallen = true
			}
		}

	}
}

func renderHorses(v *gocui.View) error {

	for i := 0; i < 5; i++ {
		h := strconv.Itoa(i+1) + ". " + PadRight(horses[i].name, " ", 9)

		footPrint := ""

		maxExtent := horses[i].pos
		if maxExtent > finishLine {
			maxExtent = finishLine
		}

		for j := 0; j < maxExtent; j++ {
			footPrint += "."
		}

		h += footPrint

		if horses[i].fallen {
			h += "X"
			_, found := Find(comments, horses[i].name+" has fallen. The jockey is well.")
			if !found {
				comments = append(comments, horses[i].name+" has fallen. The jockey is well.")
			}
		}

		// move this to checkVictoryConditions()
		if horses[i].pos >= finishLine && !horses[i].fallen {

			if won == -1 {
				arrivalIdx++
				horses[i].winner = true
				horses[i].place = 1
				_, found := Find(comments, horses[i].name+" wins the race!")
				if !found {
					comments = append(comments, horses[i].name+" wins the race!")
					won = i
				}
			} else if !horses[i].winner {
				if horses[i].place != 0 {
					gap := 4
					if horses[i].place == 3 {
						gap = 3
					}
					if horses[i].place == 4 {
						gap = 2
					}
					if horses[i].place == 5 {
						gap = 1
					}

					h += " " + PadLeft(" ", " ", gap) + strconv.Itoa(horses[i].place) + " place"
				}
			}

			if !horses[i].finisher {
				horses[i].place = arrivalIdx
				arrivalIdx++
				horses[i].finisher = true
			}

		}

		if horses[i].winner {
			h += "      1st place, WINNER"
		}

		fmt.Fprintln(v, h)
	}

	return nil
}

func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func PadRight(str, pad string, lenght int) string {
	for {
		str += pad
		if len(str) > lenght {
			return str[0:lenght]
		}
	}
}

func PadLeft(s string, padStr string, overallLen int) string {
	var padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = strings.Repeat(padStr, padCountInt) + s
	return retStr[(len(retStr) - overallLen):]
}

func layout(g *gocui.Gui) error {

	if v, err := g.SetView("raceField", 0, 0, 79, 13); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Camptown Races"

		fmt.Fprintln(v, "\n\n Welcome to "+place.raceCourse+", in "+place.county+", "+place.country+".")
		fmt.Fprintln(v, " Today the weather is "+renderWeatherInfo()+".")
		raceLength := fmt.Sprintf("%.2f", race.lengthFurlong)
		fmt.Fprintln(v, " The next scheduled race is: "+race.name+", a "+race.category+" race.\n Its length is "+raceLength+" furlongs.")
		fmt.Fprintln(v, "\n Go to race (press r)")
	}

	if v, err := g.SetView("command", 0, 14, 22, 23); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, "(v)iew statistics")
		fmt.Fprintln(v, "(s)tart the race")
		fmt.Fprintln(v, "(n)ew race")
		v.Title = "Commands"
	}

	if v, err := g.SetView("comments", 23, 14, 79, 23); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, "Not available yet.")
		v.Title = "Race speaker"
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
	if err := g.SetKeybinding("", 'r', gocui.ModNone, openRace); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'n', gocui.ModNone, newRace); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'v', gocui.ModNone, showStats); err != nil {
		return err
	}
	return nil
}

func showStats(g *gocui.Gui, v *gocui.View) error {

	g.Update(func(g *gocui.Gui) error {
		v, err := g.View("comments")
		if err != nil {
			return err
		}
		v.Clear()
		v.Title = "Horses Overall Condition"
		fmt.Fprintln(v, " ")
		for i := 0; i < 5; i++ {
			fmt.Fprintln(v, " "+PadRight(horses[i].name, " ", 10)+"-> "+strconv.Itoa(horses[i].strenght)+"0%")
		}

		return nil
	})

	return nil
}

func newRace(g *gocui.Gui, v *gocui.View) error {
	initGame()
	horses = generateHorses()
	place = generatePlace()
	race = generateRace()

	g.Update(func(g *gocui.Gui) error {
		v, err := g.View("raceField")
		if err != nil {
			return err
		}
		v.Clear()

		fmt.Fprintln(v, "\n\n Welcome to "+place.raceCourse+", in "+place.county+", "+place.country+".")
		fmt.Fprintln(v, " Today the weather is "+renderWeatherInfo()+".")
		raceLength := fmt.Sprintf("%.2f", race.lengthFurlong)
		fmt.Fprintln(v, " The next scheduled race is: "+race.name+", a "+race.category+" race.\n Its length is "+raceLength+" furlongs.")
		fmt.Fprintln(v, "\n Go to race (press r)")

		return nil
	})

	return nil
}

func openRace(g *gocui.Gui, v *gocui.View) error {

	g.Update(func(g *gocui.Gui) error {
		v, err := g.View("raceField")
		if err != nil {
			return err
		}
		v.Clear()

		renderRaceTitle(v)
		renderHorses(v)
		return nil
	})

	return nil
}

func updateComments(g *gocui.Gui, v *gocui.View) error {

	g.Update(func(g *gocui.Gui) error {
		v, err := g.View("comments")
		if err != nil {
			return err
		}
		v.Clear()
		v.Title = "Race speaker"
		fmt.Fprintln(v, " ")
		for _, c := range comments {
			fmt.Fprintln(v, " "+c)
		}

		return nil
	})

	return nil
}

func renderRaceTitle(v *gocui.View) {
	fmt.Fprintln(v, "\n\n "+race.name+" at "+place.raceCourse)
	fmt.Fprintln(v, " Weather: "+renderWeatherInfo()+"\n")
}

func renderWeatherInfo() string {
	weatherInfo := ""
	if place.weather == 0 {
		weatherInfo = "good with sun"
	}
	if place.weather == 1 {
		weatherInfo = "covered, with chances of shower"
	}
	if place.weather == 2 {
		weatherInfo = "bad, expected heavy rain"
	}
	return weatherInfo
}

func start(g *gocui.Gui, v *gocui.View) error {
	go counter(g)
	return nil
}

func someonePassedTheFinishLine() int {
	for i := 0; i < 5; i++ {
		if (horses[i].pos) >= finishLine {
			return i
		}
	}
	return -1
}

func allEndedRace() bool {
	howManyPassed := 0
	for i := 0; i < 5; i++ {
		if (horses[i].pos) > finishLine || horses[i].fallen {
			howManyPassed++
		}
	}
	return howManyPassed == 5
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
		case <-time.After(650 * time.Millisecond):
			ctr++

			g.Update(func(g *gocui.Gui) error {
				v, err := g.View("raceField")
				if err != nil {
					return err
				}
				v.Clear()
				renderRaceTitle(v)
				moveHorses()
				renderHorses(v)
				updateComments(g, v)
				return nil
			})

			if allEndedRace() {
				ctr = 0
				<-done
			}
		}
	}
}
