package events

import (
	"encoding/gob"
	"fmt"

	"github.com/libp2p/go-libp2p-core/peer"
)

type (
	EventType int

	Event interface {
		From() peer.ID
		String() string
		Type() EventType
	}

	StartGameVote struct{ from peer.ID }
	StartGame     struct{ from peer.ID }
	StartKickVote struct {
		from, Kick peer.ID
		Reason     string
	}
	KickVote struct{ from, Kick peer.ID }
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

func init() {
	event_types := []Event{
		&StartGame{}, &StartGameVote{}, &KickVote{}, &StartKickVote{},
	}
	for int := range event_types {
		gob.Register(int)
	}
}

func (*StartGameVote) Type() EventType  { return EVENT_START_GAME_VOTE }
func (ev *StartGameVote) From() peer.ID { return ev.from }
func (ev *StartGameVote) String() string {
	return fmt.Sprintf("Start of voting for next game!")
}

func (*StartGame) Type() EventType  { return EVENT_START_GAME }
func (ev *StartGame) From() peer.ID { return ev.from }
func (ev *StartGame) String() string {
	return fmt.Sprintf("%v voted for next game!", ev.from)
}

func (*StartKickVote) Type() EventType  { return EVENT_START_KICK_VOTE }
func (ev *StartKickVote) From() peer.ID { return ev.from }
func (ev *StartKickVote) String() string {
	return fmt.Sprintf(
		"%v started voting for kicking peer %v! Reason: \"%v\"",
		ev.from, ev.Kick, ev.Reason,
	)
}

func (*KickVote) Type() EventType  { return EVENT_KICK_VOTE }
func (ev *KickVote) From() peer.ID { return ev.from }
func (ev *KickVote) String() string {
	return fmt.Sprintf("%v voted for kicking peer %v!", ev.from, ev.Kick)
}
