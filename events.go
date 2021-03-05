package main

import (
	"fmt"

	"github.com/libp2p/go-libp2p-core/peer"
)

type (
	EventType int

	Event interface {
		From() peer.ID
		String() string
	}

	EventStartGameVote struct{ from peer.ID }
	EventStartGame     struct{ from peer.ID }
)

const (
	// Event which starts voting for game
	EVENT_START_GAME_VOTE EventType = iota
	// Event which happends during vote for game.
	// It means that peer is agreed to start game
	EVENT_START_GAME
	// Event which starts voting for kick of some peer
	EVENT_START_KICK_VOTE
	// Event which happends during vote of the peer.
	// Event which coresponds to agreement of peer to kick the another one
	EVENT_KICK_VOTE
)

func (ev *EventStartGameVote) From() peer.ID { return ev.from }
func (ev *EventStartGameVote) String() string {
	return fmt.Sprintf("Start of voting for next game!")
}

func (ev *EventStartGame) From() peer.ID { return ev.from }
func (ev *EventStartGame) String() string {
	return fmt.Sprintf("%v voted for next game!", ev.from)
}
