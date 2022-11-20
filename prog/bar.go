package prog

import (
	"fmt"
	"math"
	"strings"
	"sync"
)

type Bar struct {
	Msg  string
	Head string
	Pre  func(bar *Bar) string

	Max  int
	Curr int
}

type Bars struct {
	Bars []*Bar

	Row int
	Col int
	sync.Mutex
}

func (bar *Bar) perc() float64 {
	perc := float64(bar.Curr) / float64(bar.Max)
	return math.Min(1, math.Max(0, perc))
}

func (bar *Bar) print(col int) {
	if bar.Pre != nil {
		bar.Head = bar.Pre(bar)
	}
	clear()

	perc := fmt.Sprintf("%.2f%%", bar.perc()*100)
	wid := col / 2
	size := wid - len([]rune(perc)) - 3
	fill := int(math.Floor(float64(size) * bar.perc()))

	print := fmt.Sprintf(
		"%s%s%s%c%s%s%c %s",
		bar.Msg,
		strings.Repeat(" ", wid-len([]rune(bar.Head))-len([]rune(bar.Msg))),
		bar.Head,
		'[',
		strings.Repeat("=", fill),
		strings.Repeat("-", size-fill),
		']',
		perc,
	)
	if len([]rune(print)) <= col {
		fmt.Print(print)
	}
}

func New() *Bars {
	row, col := size()
	return &Bars{Row: row, Col: col}
}

func (bars *Bars) Add(bar *Bar) {
	bars.Bars = append(bars.Bars, bar)
	if len(bars.Bars) > 1 {
		fmt.Println()
	}
	bars.Print()
}

func (bars *Bars) Print() {
	bars.Lock()
	defer bars.Unlock()

	for index := range bars.Bars {
		bars.Bars[len(bars.Bars)-index-1].print(bars.Col)
		up(1)
	}
	down(len(bars.Bars))
}
