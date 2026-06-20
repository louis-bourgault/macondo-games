package snake

//a template game to copy and build off.

import (
	"math/rand"
	"strconv"

	"github.com/louis-bourgault/macondo-games/software/internal/game"
	"github.com/louis-bourgault/macondo-games/software/internal/helpers"
	"github.com/louis-bourgault/macondo-games/software/internal/platform"
)

type RingBuffer struct {
	data  []BoardLocation
	head  int //newest  segment (snake head)
	tail  int //oldest segment
	size  int
	count int
}

func NewRingBuffer(capacity int) *RingBuffer {
	return &RingBuffer{data: make([]BoardLocation, 0, capacity), size: capacity}
}

type BoardLocation struct {
	x int8
	y int8
}

func (r *RingBuffer) Push(val BoardLocation) bool {
	if r.count == r.size {
		return false
	}
	if r.count > 0 {
		r.head = (r.head + 1) % r.size
	}
	if r.head >= len(r.data) {
		r.data = append(r.data, val)
	} else {
		r.data[r.head] = val
	}

	r.count++
	return true
}

func (r *RingBuffer) Pop() (BoardLocation, bool) {
	if r.count == 0 {
		return BoardLocation{}, false
	}
	val := r.data[r.tail]
	r.tail = (r.tail + 1) % r.size
	r.count--
	return val, true
}

func New() *Snake {
	p := NewRingBuffer(576)
	p.Push(BoardLocation{x: 5, y: 12})
	p.Push(BoardLocation{x: 6, y: 12})
	p.Push(BoardLocation{x: 7, y: 12})
	bl := make(map[BoardLocation]bool)
	bl[BoardLocation{x: 5, y: 12}] = true
	bl[BoardLocation{x: 6, y: 12}] = true
	bl[BoardLocation{x: 7, y: 12}] = true

	return &Snake{
		pieces:         *p,  //576 is equal to 24*24, the size of the screen
		timeSince:      0.5, //so that we get movement the moment that we hit a direction
		direction:      platform.None,
		boardLocations: bl,
		appleLocation:  BoardLocation{x: 0, y: 0},
		gameOver:       false,
	}
}

type Snake struct {
	pieces         RingBuffer
	timeSince      float64
	direction      platform.Button //we're gonna reuse this, and just ignore the buttons for A,B,X,Y
	boardLocations map[BoardLocation]bool
	appleLocation  BoardLocation
	gameOver       bool
}

func (s *Snake) GetPressedDirection(input platform.InputSystem) platform.Button {
	if input.WasKeyJustPressed(platform.Left) {
		return platform.Left
	}
	if input.WasKeyJustPressed(platform.Right) {
		return platform.Right
	}
	if input.WasKeyJustPressed(platform.Up) {
		return platform.Up
	}
	if input.WasKeyJustPressed(platform.Down) {
		return platform.Down
	}
	return 0
}

func IsOpposite(a, b platform.Button) bool { //there's probably a more elegant way to check this
	if a == platform.Left && b == platform.Right {
		return true
	} else if a == platform.Right && b == platform.Left {
		return true
	} else if a == platform.Up && b == platform.Down {
		return true
	} else if a == platform.Down && b == platform.Up {
		return true
	}
	if a == platform.None && b == platform.Left { //the starting setup has three segments, with the front one at the right. a will only ever be none in this condition.
		return true
	}
	return false

}

func (s *Snake) Update(dt float64, input platform.InputSystem, log platform.LogSystem) game.Game {
	if input.WasKeyJustPressed(platform.Exit) { //each game can define its own way to exit, but this is probably the simplest one
		return nil //to exit the game, just pass 'nil'. The main function will notice this and boot up the menu application
	}
	if s.gameOver == true {
		if input.WasKeyJustPressed(platform.A) {
			return New()
		} else {
			return s
		}
	}
	direction := s.GetPressedDirection(input)

	if direction != 0 && !IsOpposite(direction, s.direction) && direction != s.direction {
		s.direction = direction
		s.timeSince = 0.5
	}
	if s.direction == platform.None {
		s.timeSince = 0.5
		return s
	}
	s.timeSince += dt
	if s.timeSince < 0.2 {
		return s
	}
	//now we're past the checks, we can actaully do game stuff
	lastPos := s.pieces.data[s.pieces.head]
	var wantedPosition BoardLocation = lastPos
	switch s.direction {
	case platform.Left:
		wantedPosition.x--
	case platform.Right:
		wantedPosition.x++
	case platform.Up:
		wantedPosition.y--
	case platform.Down:
		wantedPosition.y++
	}
	//check for wall collision
	if wantedPosition.x < 0 || wantedPosition.x >= 24 || wantedPosition.y < 0 || wantedPosition.y >= 24 {
		s.gameOver = true
		return s
	}
	//check for self collision
	if s.boardLocations[wantedPosition] == true {
		s.gameOver = true
		return s
	}
	if !(s.appleLocation == wantedPosition) {
		bl, _ := s.pieces.Pop()
		s.boardLocations[bl] = false
	} else {
		s.SpawnApple()
	}
	s.boardLocations[wantedPosition] = true
	s.pieces.Push(wantedPosition)
	s.timeSince = 0
	return s
}

func (s *Snake) SpawnApple() { //this may not be the best way to do this, imagine a board with only one free space, could take so many iterations to find it.
	//we also should seed this randomness somewhere if we're using the device version, maybe do something with noise in the cmd module
	found := false
	var possibleLocation BoardLocation
	for found == false {
		possibleLocation = BoardLocation{x: int8(rand.Float32() * 24), y: int8(rand.Float32() * 24)}
		if s.boardLocations[possibleLocation] != true {
			found = true
		}
	}
	s.appleLocation = possibleLocation
}

func (s *Snake) Draw(screen platform.Screen) {
	screen.Fill(0x0000)
	if s.pieces.count == 0 {
		return
	}
	if s.gameOver == true {
		helpers.DrawText(screen, 10, 10, "game over", 0xf800)
		helpers.DrawText(screen, 10, 30, "your score: "+strconv.Itoa(s.pieces.count-3), 0x001f)
		helpers.DrawText(screen, 10, 70, "press A to restart", 0xf800)
		screen.Present()
		return

	}
	currentIndex := s.pieces.head
	for i := 0; i < s.pieces.count; i++ {
		colour := helpers.RGBto565(0, 255-(128/s.pieces.count*i), 0)
		piece := s.pieces.data[currentIndex]
		screen.FillRect(int(piece.x)*10, int(piece.y)*10, 10, 10, colour)
		currentIndex--
		if currentIndex < 0 {
			currentIndex = s.pieces.size - 1
		}
	}
	screen.FillRect(int(s.appleLocation.x)*10, int(s.appleLocation.y)*10, 10, 10, 0xf800)
	screen.Present() //present the screen
}
