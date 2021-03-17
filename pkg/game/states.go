package game

import (
	"fmt"
	"time"

	"github.com/i1i1/rpc-go/pkg/events"
	"github.com/i1i1/rpc-go/pkg/my_crypto"
	"github.com/libp2p/go-libp2p-core/peer"
)

// States of game logic
const (
	STATE_WAIT = iota
	STATE_KICK_VOTE
	STATE_GAME_VOTE
	STATE_MOVES_EXCHANGE
	STATE_KEYS_EXCHANGE
	STATE_ENDGAME
)

// Possible Turns
type Turn uint32

const (
	TURN_ROCK = iota
	TURN_PAPER
	TURN_SCISSORS
	TURN_LIZARD
	TURN_SPOCK
)

const (
	// In seconds
	TIMEOUT = 10
)

type (
	GameStateType int
	GameState     struct {
		// gamestates can produce events that should be published
		pubChannel Publisher
		stateType  GameStateType
		// If vote is going on, it is not nil
		vote *Vote
		// cancellation timer
		timer *time.Timer
		// moves
		moves *Moves
	}

	Moves struct {
		encrypted map[peer.ID][]byte
		keys      map[peer.ID][]byte
	}

	// Info about votes
	Vote struct {
		// who started vote
		author peer.ID
		voted  map[peer.ID]struct{}
	}
)

func makeStartState(pubChannel Publisher) GameState {
	return GameState{pubChannel: pubChannel, stateType: STATE_WAIT, vote: nil}
}

func (this *GameState) processEvent(
	// incoming event
	event events.Event,
	// peers who participate
	peers map[peer.ID]struct{},
	// player
	self events.Player,
) {
	// fmt.Printf("%#v\n", this)
	// fmt.Printf("%#v\n", event)
	// fmt.Printf("%#v\n", event.Type())
	switch this.stateType {
	case STATE_WAIT:
		if event.Type() == events.EVENT_START_GAME_VOTE {

			voteList := make(map[peer.ID]struct{})
			vote := Vote{voted: voteList, author: event.From().ID}

			// next state - vote for game start
			this.stateType = STATE_GAME_VOTE
			this.vote = &vote

			// start timeout timer
			if this.vote.author == self.ID {
				this.timer = time.NewTimer(time.Second * TIMEOUT)
				this.pubChannel.Publish(events.NewStartGame(self))
				// fmt.Println("here1")
				go this.cancel(self, "voting for game start")
			}
		}
	case STATE_GAME_VOTE:
		if event.Type() == events.EVENT_START_GAME {
			// add players vote
			this.vote.voted[event.From().ID] = struct{}{}
			// check if this was last one
			if len(peers) == len(this.vote.voted) {
				if this.timer != nil {
					this.timer.Stop()
				}

				// update state. preparing to moves exchange
				this.stateType = STATE_MOVES_EXCHANGE
				this.vote.voted = nil
				this.moves = &Moves{}
				this.moves.encrypted = make(map[peer.ID][]byte)

				if this.vote.author == self.ID {
					this.timer = time.NewTimer(time.Second * TIMEOUT)
					// fmt.Println("here2")
					go this.cancel(self, "moves exchange")
				}
			}
		}
		if event.Type() == events.EVENT_CANCEL && this.vote.author == event.From().ID {
			this.stateType = STATE_WAIT
			this.vote = nil
			this.moves = nil
		}
	case STATE_MOVES_EXCHANGE:
		if event.Type() == events.EVENT_MOVE {
			this.moves.encrypted[event.From().ID] = event.ExtraContent()
			if len(peers) == len(this.moves.encrypted) {
				// fmt.Println("here")
				if this.timer != nil {
					this.timer.Stop()
				}
				this.stateType = STATE_KEYS_EXCHANGE
				this.moves.keys = make(map[peer.ID][]byte)
				// TODO SEND ACTUAL KEY (HOW???)
				this.pubChannel.Publish(events.NewKey(self, nil))
				this.timer = time.NewTimer(time.Second * TIMEOUT)
				// fmt.Println("here3")
				go this.cancel(self, "keys exchange")
			}
		}
		if event.Type() == events.EVENT_CANCEL && this.vote.author == event.From().ID {
			this.stateType = STATE_WAIT
			this.vote = nil
			this.moves = nil
		}
	case STATE_KEYS_EXCHANGE:
		if event.Type() == events.EVENT_KEY {
			this.moves.keys[event.From().ID] = event.ExtraContent()
			// fmt.Println("here1", len(peers), len(this.moves.keys))
			if len(peers) == len(this.moves.keys) {
				this.timer.Stop()
				this.resolveResult(self, peers)
				// reset game state
				this.stateType = STATE_WAIT
				this.moves = nil
				this.vote = nil
			}
		}
		if event.Type() == events.EVENT_CANCEL && this.vote.author == event.From().ID {
			this.stateType = STATE_WAIT
			this.vote = nil
			this.moves.encrypted = nil
			this.moves.keys = nil
		}
	}
}
func (this *GameState) cancel(self events.Player, what string) {
	<-this.timer.C
	this.pubChannel.Publish(events.NewCancel(self, what))
}
func (this *GameState) resolveResult(self events.Player, peers map[peer.ID]struct{}) error {
	moves := this.moves
	var actions = make(map[Turn][]peer.ID)
	// Decrypt all actions
	for player, _ := range peers {
		// fmt.Println("states", moves.encrypted)
		turn := Turn(my_crypto.Decrypt(moves.encrypted[player], moves.keys[player]))
		if _, ok := actions[turn]; ok {
			actions[turn] = append(actions[turn], player)
		} else {
			actions[turn] = []peer.ID{player}
		}
	}
	// fmt.Println("actions", actions)
	if len(actions) != 2 {
		this.pubChannel.Publish(events.NewMessage(self, "Tie!"))
	} else {
		keys := make([]Turn, 0, len(actions))
		for k := range actions {
			keys = append(keys, k)
		}

		var winners []peer.ID

		if keys[0].fights(keys[1]) {
			winners = actions[keys[0]]
		} else {
			winners = actions[keys[1]]
		}

		for _, player := range winners {
			if self.ID == player {
				messageStr := fmt.Sprintf("%v Won!", self.Nick)
				this.pubChannel.Publish(events.NewMessage(self, messageStr))
				return nil
			}
		}
		messageStr := fmt.Sprintf("%v lost. Maybe next time :(", self.Nick)
		this.pubChannel.Publish(events.NewMessage(self, messageStr))
		return nil
	}
	return nil
}

func (this Turn) fights(other Turn) bool {
	switch this {
	case TURN_ROCK:
		switch other {
		case TURN_PAPER:
			return false
		case TURN_SCISSORS:
			return true
		case TURN_LIZARD:
			return true
		case TURN_SPOCK:
			return false
		}
	case TURN_PAPER:
		switch other {
		case TURN_ROCK:
			return true
		case TURN_SCISSORS:
			return false
		case TURN_LIZARD:
			return false
		case TURN_SPOCK:
			return true
		}
	case TURN_SCISSORS:
		switch other {
		case TURN_ROCK:
			return false
		case TURN_PAPER:
			return true
		case TURN_LIZARD:
			return true
		case TURN_SPOCK:
			return false
		}
	case TURN_LIZARD:
		switch other {
		case TURN_ROCK:
			return false
		case TURN_PAPER:
			return true
		case TURN_SCISSORS:
			return false
		case TURN_SPOCK:
			return true
		}
	case TURN_SPOCK:
		switch other {
		case TURN_ROCK:
			return true
		case TURN_PAPER:
			return false
		case TURN_SCISSORS:
			return true
		case TURN_LIZARD:
			return false
		}
	}
	return false
}
