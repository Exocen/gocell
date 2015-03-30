package main

import (
	"strconv"
	"math/rand"
	"time"
	"github.com/artsimboldo/golibs"
	"github.com/nsf/termbox-go"
	"github.com/davecheney/profile"
)

const (
	worldSize = 50
	agentNum = 26
	epoch = 100
	sleepTime = 30
)

var (
	directions = [8]direction{
		direction{1,0},
		direction{1,1},
		direction{0,1},
		direction{-1,-1},
		direction{-1,0},
		direction{-1,1},
		direction{0,-1},
		direction{1,-1},
	}
)

type cell struct {
	rune
	*agent
}

type position struct {
	x int
	y int
}

type direction struct {
	dx int
	dy int
}

// Duck agent list to shuffle's Interface
type AgentList []agent
func (a AgentList) Len() int { return len(a) }
func (a AgentList) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

type agent struct {
	id int
	world [][]cell
	position
	direction
}

func (self *agent) turn() {
	self.direction = directions[rand.Intn(7)]
}

func (self *agent) move() bool {
	width := len(self.world)
	height := len(self.world[0])
	newpos := self.position
	newpos.x += self.direction.dx
	newpos.y += self.direction.dy
	if newpos.x >= 0 && newpos.x < width && newpos.y >= 0 && newpos.y < height && self.world[newpos.x][newpos.y].agent == nil {
		self.world[self.position.x][self.position.y].agent = nil
		self.position = newpos
		self.world[self.position.x][self.position.y].agent = self
		return true
	} else {
		return false
	}
}

func (self *agent) toString() string {
    return strconv.Itoa(self.id) + " at (" + strconv.Itoa(self.position.x) + ", " + strconv.Itoa(self.position.y) + ")\n"
}

func drawWorld(world [][]cell) {
	width := len(world)
	height := len(world[0])
	for j := 0; j < height; j++ {
		for i := 0; i < width; i++ {
			if world[i][j].agent == nil {
				termbox.SetCell(i, j, world[i][j].rune, termbox.ColorGreen, termbox.ColorDefault)
			} else {
				termbox.SetCell(i, j, rune(world[i][j].agent.id + 'a'), termbox.ColorWhite | termbox.AttrBold, termbox.ColorDefault)
			}
		}
	}
	termbox.Flush()
}

func update(world [][]cell, agents []agent) {
	// Shuffle agents for order independance
	golibs.Shuffle(AgentList(agents))

	// Simple agent behaviour: turn if no move is possible
	for i := range agents {
		oldpos := agents[i].position
		if agents[i].move() {
			termbox.SetCell(oldpos.x, oldpos.y, world[oldpos.x][oldpos.y].rune, termbox.ColorGreen, termbox.ColorDefault)
			termbox.SetCell(agents[i].position.x, agents[i].position.y, rune(agents[i].id + 'a'), termbox.ColorWhite | termbox.AttrBold, termbox.ColorDefault)
		} else {
			agents[i].turn()
		}
	}
}

func main() {
	if agentNum > worldSize * worldSize {
		panic("Too many agents for the world created!")
	}

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	// Initiate simple profiler
 	defer profile.Start(profile.CPUProfile).Stop()

	// Populate a world slice with cells
	world := make([][]cell, worldSize)
	for x := 0; x < worldSize; x++ {
		world[x] = make([]cell, worldSize)
		for y := 0; y < worldSize; y++ {
			world[x][y] = cell{'.', nil}
		}
	}

	// Populate an agents slice with agents
	agents := make([]agent, agentNum)
	rand.Seed(time.Now().UTC().UnixNano())
	x := rand.Intn(worldSize - 1) ; y := rand.Intn(worldSize - 1)
	for i := range agents {
		for world[x][y].agent != nil {
			x = rand.Intn(worldSize - 1) ; y = rand.Intn(worldSize - 1)
		}
		agents[i] = agent{i, world, position{x, y}, direction{0, 0}}
		agents[i].turn()
		world[x][y].agent = &agents[i]
	}

	// Install main loop
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	// goroutine for non-blocking PollEvent 
	eventQueue := make(chan termbox.Event)
	go func() {
		for {
			eventQueue <- termbox.PollEvent()
		}
	}()

	drawWorld(world)
loop:
	for i := 0 ; i < epoch; i++ {
		select {
		case ev := <-eventQueue:
			if ev.Key == termbox.KeyEsc {
				break loop
			}
		default:
			update(world, agents)
			itertext := strconv.Itoa(i)
			for index, rune := range itertext {
				termbox.SetCell(index, 50, rune, termbox.ColorWhite | termbox.AttrBold, termbox.ColorDefault)
			}
			termbox.Flush()
			time.Sleep(sleepTime * time.Millisecond)
		}
	}
}