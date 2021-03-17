package events

import (
	"encoding/binary"
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
		ExtraContent() []byte
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
	Cancel struct {
		FromPlayer Player
		What       string
	}
	Move struct {
		EncryptedMove []byte
		FromPlayer    Player
	}
	Key struct {
		FromPlayer Player
		key        []byte
	}
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

	// Event with R-P-S-L-Spock
	EVENT_MOVE

	// Key to decrypt move
	EVENT_KEY
)

func init() {
	gob.Register(&StartGame{})
	gob.Register(&StartGameVote{})
	gob.Register(&StartKick{})
	gob.Register(&StartKickVote{})
	gob.Register(&Message{})
	gob.Register(&Cancel{})
	gob.Register(&Move{})
	gob.Register(&Key{})
}

func NewStartGameVote(from Player) *StartGameVote { return &StartGameVote{from} }
func (*StartGameVote) Type() EventType            { return EVENT_START_GAME_VOTE }
func (ev *StartGameVote) From() Player            { return ev.FromPlayer }
func (ev *StartGameVote) String() string {
	return fmt.Sprintf("Start of voting for next game!")
}
func (ev *StartGameVote) ExtraContent() []byte { return nil }

func NewStartGame(from Player) *StartGame { return &StartGame{from} }
func (*StartGame) Type() EventType        { return EVENT_START_GAME }
func (ev *StartGame) From() Player        { return ev.FromPlayer }
func (ev *StartGame) String() string {
	return fmt.Sprintf("%v voted for next game!", ev.FromPlayer.Nick)
}
func (ev *StartGame) ExtraContent() []byte { return nil }

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
func (ev *StartKick) ExtraContent() []byte { return nil }

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
func (ev *StartKickVote) ExtraContent() []byte { return nil }

func NewMessage(from Player, text string) *Message {
	return &Message{
		FromPlayer: from,
		Text:       text,
	}
}
func (*Message) Type() EventType         { return EVENT_MESSAGE }
func (ev *Message) From() Player         { return ev.FromPlayer }
func (ev *Message) String() string       { return ev.Text }
func (ev *Message) ExtraContent() []byte { return nil }

func NewCancel(from Player, what string) *Cancel {
	return &Cancel{FromPlayer: from, What: what}
}
func (*Cancel) Type() EventType { return EVENT_CANCEL }
func (ev *Cancel) From() Player { return ev.FromPlayer }
func (ev *Cancel) String() string {
	return fmt.Sprintf("%v cancelled %v", ev.FromPlayer.Nick, ev.What)
}
func (ev *Cancel) ExtraContent() []byte { return nil }

func NewMove(from Player, turn uint32) *Move {
	bs := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(bs, turn)
	// fmt.Println("events", bs)
	// TODO ENCRYPTION
	return &Move{FromPlayer: from, EncryptedMove: bs}
}
func (*Move) Type() EventType         { return EVENT_MOVE }
func (ev *Move) From() Player         { return ev.FromPlayer }
func (ev *Move) String() string       { return fmt.Sprintf("%v sent his move!", ev.FromPlayer.Nick) }
func (ev *Move) ExtraContent() []byte { return ev.EncryptedMove }

func NewKey(from Player, key []byte) *Key { return &Key{FromPlayer: from, key: key} }
func (*Key) Type() EventType              { return EVENT_KEY }
func (this *Key) From() Player            { return this.FromPlayer }
func (this *Key) String() string {
	return fmt.Sprintf("%v sent his decryption key!", this.FromPlayer.Nick)
}
func (this *Key) ExtraContent() []byte { return this.key }
