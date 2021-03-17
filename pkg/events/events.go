package events

import (
	"encoding/gob"
	"fmt"

	"github.com/libp2p/go-libp2p-core/peer"
)

type (
	EventType int

	Event interface {
		From() Player
		String() string
		Type() EventType
	}

	// Player is id of player
	Player struct {
		// ID is libp2p id (generally key pair)
		ID peer.ID
		// Nick is players nickname
		Nick string
	}

	StartGameVote struct{ FromPlayer Player }
	StartGame     struct{ FromPlayer Player }
	StartKick     struct {
		FromPlayer, Kick Player
		Reason           string
	}
	StartKickVote struct{ FromPlayer, Kick Player }
	Message       struct {
		FromPlayer Player
		Text       string
	}
	Cancel struct{ FromPlayer Player }
)

const (
	// Event which starts voting for game
	EVENT_START_GAME_VOTE EventType = iota
	// Event which happends during vote for game.
	// It means that peer is agreed to start game
	EVENT_START_GAME
	// Event which starts voting for kick of some peer
	EVENT_START_KICK
	// Event which happends during vote of the peer.
	// Event which coresponds to agreement of peer to kick the another one
	EVENT_START_KICK_VOTE
	// Event with some text message from peer
	EVENT_MESSAGE

	EVENT_CANCEL
)

func init() {
	gob.Register(&StartGame{})
	gob.Register(&StartGameVote{})
	gob.Register(&StartKick{})
	gob.Register(&StartKickVote{})
	gob.Register(&Message{})
}

func NewStartGameVote(from Player) *StartGameVote { return &StartGameVote{from} }
func (*StartGameVote) Type() EventType            { return EVENT_START_GAME_VOTE }
func (ev *StartGameVote) From() Player            { return ev.FromPlayer }
func (ev *StartGameVote) String() string {
	return fmt.Sprintf("Start of voting for next game!")
}

func NewStartGame(from Player) *StartGame { return &StartGame{from} }
func (*StartGame) Type() EventType        { return EVENT_START_GAME }
func (ev *StartGame) From() Player        { return ev.FromPlayer }
func (ev *StartGame) String() string {
	return fmt.Sprintf("%v voted for next game!", ev.FromPlayer.Nick)
}

func NewStartKick(from, kick Player, reason string) *StartKick {
	return &StartKick{
		FromPlayer: from,
		Kick:       kick,
		Reason:     reason,
	}
}
func (*StartKick) Type() EventType { return EVENT_START_KICK }
func (ev *StartKick) From() Player { return ev.FromPlayer }
func (ev *StartKick) String() string {
	return fmt.Sprintf(
		"%v started voting for kicking peer %v! Reason: \"%v\"",
		ev.FromPlayer.Nick, ev.Kick, ev.Reason,
	)
}

func NewStartKickVote(from, kick Player) *StartKickVote {
	return &StartKickVote{
		FromPlayer: from,
		Kick:       kick,
	}
}
func (*StartKickVote) Type() EventType { return EVENT_START_KICK_VOTE }
func (ev *StartKickVote) From() Player { return ev.FromPlayer }
func (ev *StartKickVote) String() string {
	return fmt.Sprintf("%v voted for kicking peer %v!", ev.FromPlayer.Nick, ev.Kick)
}

func NewMessage(from Player, text string) *Message {
	return &Message{
		FromPlayer: from,
		Text:       text,
	}
}
func (*Message) Type() EventType   { return EVENT_MESSAGE }
func (ev *Message) From() Player   { return ev.FromPlayer }
func (ev *Message) String() string { return ev.Text }

func NewCancel(from Player) *Cancel {
	return &Cancel{FromPlayer: from}
}
func (*Cancel) Type() EventType { return EVENT_CANCEL }
func (ev *Cancel) From() Player { return ev.FromPlayer }
func (ev *Cancel) String() string {
	return fmt.Sprintf("%v cancelled current vote", ev.FromPlayer.Nick)
}
