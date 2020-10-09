package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"text/template"
	"time"

	match "github.com/alexpantyukhin/go-pattern-match"
	petname "github.com/dustinkirkland/golang-petname"
	cam "github.com/iancoleman/strcase"
	"github.com/jroimartin/gocui"
)

type horse struct {
	Name     string
	Jockey   string
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

const horsesNumber = 5
const courseStep = 45

var (
	done       = make(chan struct{})
	won        = -1
	ctr        = 0
	finishLine = courseStep
	horses     []horse
	place      placeInfo
	placeData  = make([]placeInfo, 0)
	race       raceInfo
	comments   = make([]string, 0)
	arrivalIdx = 0
	words      = 1
	separator  = " "
)

func initGame() {
	done = make(chan struct{})
	won = -1
	ctr = 0
	finishLine = courseStep
	horses = nil
	comments = make([]string, 0)
	arrivalIdx = 0
}

func main() {

	initGame()
	loadPlaces()

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

func find(what interface{}, where []interface{}) (idx int) {
	for i, v := range where {
		if v == what {
			return i
		}
	}
	return -1
}

func loadPlaces() {
	placeData = append(placeData, placeInfo{city: "Salisbury", raceCourse: "Salisbury Racecourse", county: "Wiltshire", country: "England"})
	placeData = append(placeData, placeInfo{city: "Cheltenham", raceCourse: "Cheltenham Racecourse", county: "Gloucestershire", country: "England"})
	placeData = append(placeData, placeInfo{city: "Stratford-upon-Avon", raceCourse: "Stratford-on-Avon Racecourse", county: "Warwickshire", country: "England"})
	placeData = append(placeData, placeInfo{city: "Newbury", raceCourse: "Newbury Racecourse", county: "Berkshire", country: "England"})
	placeData = append(placeData, placeInfo{city: "Wolverhampton", raceCourse: "Wolverhampton Racecourse", county: "West Midlands", country: "England"})
}

func generateRace() raceInfo {
	return raceInfo{name: "Cathedral Stakes", category: "Flat", branch: "", lengthFurlong: 6}
}

func generatePlace() placeInfo {
	weather := rand.Intn(3)
	cityIdx := rand.Intn(5)
	place := placeData[cityIdx]
	place.weather = weather
	return place
}

func generateHorses() []horse {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < horsesNumber; i++ {

		temp := petname.Generate(words, separator)
		force := rand.Intn(9) + 1
		year := rand.Intn(7) + 1

		horses = append(horses,
			horse{
				Name:     cam.ToCamel(temp),
				age:      year,
				strenght: force,
				pos:      1,
				fallen:   false,
				winner:   false,
				finisher: false,
				place:    0,
				Jockey:   cam.ToCamel(petname.Generate(1, separator)[0:1]) + ". " + cam.ToCamel(petname.Generate(1, separator))})
	}

	return horses
}

func moveHorses() {
	for i := 0; i < horsesNumber; i++ {
		stride := rand.Intn(horses[i].strenght) + 1

		_, res := match.Match(stride).
			When(func(t int) bool { return t >= 9 }, 3).
			When(func(t int) bool { return t < 9 && t >= 5 }, 2).
			When(func(t int) bool { return t < 4 && t > 1 }, 1).
			When(match.ANY, 1).
			Result()

		if horses[i].age == 4 {
			chanceForYoung := rand.Intn(4)
			if chanceForYoung == 4 {
				stride++
			}
		}

		if horses[i].age > 5 {
			chanceForOlder := rand.Intn(10)
			if chanceForOlder == 10 {
				stride--
			}
		}

		stride = res.(int)

		if !(horses[i].fallen) && horses[i].pos <= finishLine {

			horses[i].pos += stride

			_, fallFactor := match.Match(place.weather).
				When(0, 200).
				When(1, 170).
				When(2, 120).
				When(3, 90).
				Result()

			fall := rand.Intn(fallFactor.(int))

			if fall <= 1 {
				horses[i].fallen = true
			}
		}

		if horses[i].pos >= finishLine && !horses[i].fallen {
			if won == -1 {
				arrivalIdx++
				horses[i].winner = true
				horses[i].place = arrivalIdx
				won = i
			}

			if !horses[i].finisher {
				horses[i].place = arrivalIdx
				arrivalIdx++
				horses[i].finisher = true
			}
		}
	}
}

func renderHorses(v *gocui.View) error {

	var fallenTemplates = make([]string, 2)
	fallenTemplates[0] = "{{.Name}} has fallen. {{ .Jockey}} is well."
	fallenTemplates[1] = "Look at {{.Name}}! Such a bad fall!"
	fallIndex := rand.Intn(1)

	t := template.New("fallInfo")
	t, _ = t.Parse(fallenTemplates[fallIndex])

	for i := 0; i < horsesNumber; i++ {
		h := " " + strconv.Itoa(i+1) + ". " + PadRight(horses[i].Name, " ", 9)

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
			buf := new(bytes.Buffer)
			t.Execute(buf, horses[i])
			_, found := Find(comments, buf.String())
			if !found {
				comments = append(comments, buf.String())
			}
		}

		if horses[i].pos >= finishLine && !horses[i].fallen {

			if horses[i].winner {
				_, found := Find(comments, horses[i].Name+" wins the race!")
				if !found {
					comments = append(comments, horses[i].Name+" wins the race!")
				}
			} else {
				if horses[i].place != 0 {
					h += getPlaceText(horses[i].place)
				}
			}

			if !horses[i].finisher {
				horses[i].place = arrivalIdx
				arrivalIdx++
				horses[i].finisher = true
			}
		}

		if horses[i].winner {
			h += "  1st WINS"
		}

		fmt.Fprintln(v, h)
	}

	return nil
}

func getPlaceText(place int) string {

	val := strconv.Itoa(place)
	_, res := match.Match(place).
		When(1, "st").
		When(2, "nd").
		When(3, "rd").
		When(func(ts int) bool { return ts > 3 }, "th").
		Result()
	val += res.(string)
	return val
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

	if v, err := g.SetView("raceField", 0, 0, 79, 10); err != nil {
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

	if v, err := g.SetView("command", 0, 11, 22, 24); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, "\n (v)iew statistics")
		fmt.Fprintln(v, " (s)tart the race")
		fmt.Fprintln(v, " (n)ew race")
		fmt.Fprintln(v, " (q)uit the game")
		v.Title = "Commands"
	}

	if v, err := g.SetView("comments", 23, 11, 79, 24); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, " Not available yet.")
		v.Title = "Race speaker"
	}

	return nil
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
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
		v.Title = "Horses Overall Condition and Stats"
		fmt.Fprintln(v, " ")
		fmt.Fprintln(v, " Name      Age   Condition   Jockey")
		for i := 0; i < horsesNumber; i++ {
			fmt.Fprintln(v, " "+
				PadRight(horses[i].Name, " ", 10)+
				PadRight(strconv.Itoa(horses[i].age), " ", 6)+
				PadRight(strconv.Itoa(horses[i].strenght)+"0%", " ", 12)+
				PadRight(horses[i].Jockey, " ", 20))
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

	_, weatherInfo := match.Match(place.weather).
		When(0, "Sunny, hot temperature.").
		When(1, "Partly cloudy, chances of showers.").
		When(2, "Chilly, with heavy rain.").
		When(3, "Chilly, slightly snowing.").
		Result()

	return weatherInfo.(string)
}

func start(g *gocui.Gui, v *gocui.View) error {
	go counter(g)
	return nil
}

func someonePassedTheFinishLine() int {
	for i := 0; i < horsesNumber; i++ {
		if (horses[i].pos) >= finishLine {
			return i
		}
	}
	return -1
}

func allEndedRace() bool {
	howManyPassed := 0
	for i := 0; i < horsesNumber; i++ {
		if (horses[i].pos) > finishLine || horses[i].fallen {
			howManyPassed++
		}
	}
	return howManyPassed == horsesNumber
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
