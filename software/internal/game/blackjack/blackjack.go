package blackjack

import (
	"fmt"

	"github.com/louis-bourgault/macondo-games/software/internal/game"
	"github.com/louis-bourgault/macondo-games/software/internal/helpers"
	"github.com/louis-bourgault/macondo-games/software/internal/platform"
)

func New() *BlackjackGame {
	return &BlackjackGame{
		dealerHand:      []int{},
		playerHand:      []int{},
		bust:            false,
		twentyOne:       false,
		showDealer:      false,
		win:             false,
		money:           1000,
		bet:             200,
		selectMoneyMode: true,
		deck: []uint8{
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13,
		},
	}
}

type BlackjackGame struct {
	playerHand      []int
	dealerHand      []int
	bust            bool
	twentyOne       bool
	showDealer      bool
	win             bool
	money           uint16
	bet             uint16
	selectMoneyMode bool
	tie             bool
	deck            []uint8
}

func (d *BlackjackGame) drawFromDeck() int {
	if len(d.deck) == 0 {
		d.deck = []uint8{ //single deck, dont care about card counting.
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13,
		}
	}
	intindex := helpers.RandomInt(0, len(d.deck))
	card := d.deck[intindex]

	d.deck = append(d.deck[:intindex], d.deck[intindex+1:]...)
	return int(card)
}

func (d *BlackjackGame) hit() {
	card := d.drawFromDeck()
	d.playerHand = append(d.playerHand, card)
	if d.calculateHandValue(d.playerHand) > 21 {
		d.bust = true
		d.money -= d.bet
	}
	if d.calculateHandValue(d.playerHand) == 21 {
		d.twentyOne = true
		d.showDealer = true
		d.win = true
		d.money += d.bet
	}
}

func (d *BlackjackGame) stand() {
	d.showDealer = true
	for d.calculateHandValue(d.dealerHand) < 17 {
		card := d.drawFromDeck()
		d.dealerHand = append(d.dealerHand, card)
	}
	if d.calculateHandValue(d.dealerHand) == d.calculateHandValue(d.playerHand) {
		d.tie = true
		d.money += 0
	}
	if d.calculateHandValue(d.dealerHand) > 21 {
		d.win = true
		d.money += d.bet
	} else if d.calculateHandValue(d.playerHand) > d.calculateHandValue(d.dealerHand) {
		d.win = true
		d.money += d.bet
	} else {
		d.win = false
		d.money -= d.bet
	}

}

func (d *BlackjackGame) calculateHandValue(hand []int) int {
	value := 0
	aces := 0
	for _, card := range hand {
		if card > 10 {
			value += 10
		} else if card == 1 {
			aces++
			value += 11
		} else {
			value += card
		}
	}
	for value > 21 && aces > 0 {
		value -= 10
		aces--
	}
	return value
}

func (d *BlackjackGame) Update(dt float64, input platform.InputSystem, log platform.LogSystem) game.Game {
	if input.WasKeyJustPressed(platform.Select) { //each game can define its own way to exit, but this is probably the simplest one
		return nil //to exit the game, just pass 'nil'. The main function will notice this and boot up the menu application
	}
	if d.bust || d.twentyOne || d.showDealer || d.tie || d.win {
		if input.WasKeyJustPressed(platform.A) {
			d.playerHand = []int{}
			d.dealerHand = []int{}
			d.bust = false
			d.twentyOne = false
			d.showDealer = false
			d.tie = false
			d.win = false
			d.selectMoneyMode = true
		}
		return d //if the game is over, we just return the same game, which will keep it on the screen until the player exits

	} else if d.selectMoneyMode {
		if input.WasKeyJustPressed(platform.Down) {
			if d.bet >= 10 {
				d.bet -= 10
			}
		}
		if input.WasKeyJustPressed(platform.Up) {
			if d.bet >= 10 && d.bet+10 <= d.money {
				d.bet += 10
			}
		}
		if input.WasKeyJustPressed(platform.A) {
			d.selectMoneyMode = false
			//start the game
			d.playerHand = append(d.playerHand, d.drawFromDeck(), d.drawFromDeck())
			d.dealerHand = append(d.dealerHand, d.drawFromDeck(), d.drawFromDeck())

		}
	} else if d.showDealer {
		if input.WasKeyJustPressed(platform.A) {
			d.playerHand = []int{}
			d.dealerHand = []int{}
			d.bust = false
			d.twentyOne = false
			d.showDealer = false
			if d.calculateHandValue(d.playerHand) > 21 {
				d.money -= d.bet
			} else if d.calculateHandValue(d.dealerHand) > 21 || d.calculateHandValue(d.playerHand) > d.calculateHandValue(d.dealerHand) {
				d.money += d.bet
			} else if d.calculateHandValue(d.playerHand) < d.calculateHandValue(d.dealerHand) {
				d.money -= d.bet
			}
			//after the game is over, we go back to the money selection mode to start a new game
			d.selectMoneyMode = true
		}
	} else {

		if input.WasKeyJustPressed(platform.A) {
			d.hit()
			log.Log(fmt.Sprintf("Player hand: %v (value: %d)", d.playerHand, d.calculateHandValue(d.playerHand)))
		}
		if input.WasKeyJustPressed(platform.B) {
			d.stand()
		}
		if input.WasKeyJustPressed(platform.Up) && len(d.playerHand) == 2 && d.bet*2 <= d.money {
			//double down
			d.bet = d.bet * 2
			d.hit()
			d.stand()
		}
	}

	return d
}

func renderHand(screen platform.Screen, hand []int, x int, y int) {
	w := 25 * len(hand) //cards are 20 wide, 5 pixels of spacing between.
	locationStartX := x - (w / 2)
	for i, card := range hand {
		cardX := locationStartX + (i * 25)
		screen.FillRect(cardX, y, 20, 30, 0xffff)
		helpers.DrawText(screen, cardX+5, y+10, fmtNum(card), 0x0000)
	}
}

func renderHiddenDealerHand(screen platform.Screen, hand []int, x int, y int) {
	w := 25 * len(hand) //cards are 20 wide, 5 pixels of spacing between.
	locationStartX := x - (w / 2)
	for i := range hand {
		cardX := locationStartX + (i * 25)
		screen.FillRect(cardX, y, 20, 30, 0xffff)
		if i == 0 {
			helpers.DrawText(screen, cardX+5, y+10, fmtNum(hand[i]), 0x0000)
		} else {
			screen.FillRect(cardX+5, y+5, 10, 20, 0x0000) //cover the card value with a black rectangle to hide it
		}
	}
}

func (d *BlackjackGame) Draw(screen platform.Screen) {
	screen.Fill(0x2ac7)
	helpers.DrawText(screen, 2, 2, "Money:"+fmt.Sprintf("%d", d.money), 0xf800)

	if d.selectMoneyMode {
		helpers.DrawText(screen, 2, 20, "Select bet: "+fmt.Sprintf("%d", d.bet), 0xf800)
		helpers.DrawText(screen, 2, 30, "Use up/down to adjust", 0xffff)
		helpers.DrawText(screen, 2, 40, "Press A to start", 0xffff)

		helpers.DrawText(screen, 2, 200, "Controls:", 0x0000)

		helpers.DrawText(screen, 2, 210, "A to hit", 0xffff)
		helpers.DrawText(screen, 2, 220, "B to stand", 0xffff)
		helpers.DrawText(screen, 2, 230, "Up to double down", 0xffff)

		screen.Present()
		return
	}
	renderHand(screen, d.playerHand, 120, 160)
	if d.showDealer { //we can change this to show the dealer's hand when the game is over
		renderHand(screen, d.dealerHand, 120, 80)
	} else {
		renderHiddenDealerHand(screen, d.dealerHand, 120, 80)
	}
	if d.bust {
		helpers.DrawText(screen, 68, 130, "BUST! Again:A", 0xf800)
	} else if d.twentyOne {
		helpers.DrawText(screen, 76, 130, "21! Again:A", 0xf800)
	} else if d.tie {
		helpers.DrawText(screen, 72, 130, "TIE! Again:A", 0xf800)
	} else if d.win {
		helpers.DrawText(screen, 56, 180, "YOU WIN! Again:A", 0xf800)
	} else if d.showDealer {
		helpers.DrawText(screen, 52, 130, "YOU LOSE! Again:A", 0xf800)
	}
	helpers.DrawText(screen, 2, 200, "Value: "+fmt.Sprintf("%d", d.calculateHandValue(d.playerHand)), 0xffff)
	helpers.DrawText(screen, 2, 210, "Bet: "+fmt.Sprintf("%d", d.bet), 0xffff)

	screen.Present() //present the screen
}

func fmtNum(num int) string {
	if num == 1 {
		return "A"
	} else if num == 11 {
		return "J"
	} else if num == 12 {
		return "Q"
	} else if num == 13 {
		return "K"
	} else {
		return fmt.Sprintf("%d", num)
	}
}
