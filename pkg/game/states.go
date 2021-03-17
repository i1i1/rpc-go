package game

import (
	"github.com/i1i1/rpc-go/pkg/events"

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

type (
	GameStateType int
	GameState     struct {
		stateType GameStateType
		// If vote is going on, it is not nil
		vote *Vote
	}

	// Info about votes
	Vote struct {
		// who started vote
		author peer.ID
		voted  map[peer.ID]struct{}
	}
)

func makeStartState() GameState {
	return GameState{stateType: STATE_WAIT, vote: nil}
}

func (this GameState) processEvent(
	// incoming event
	event events.Event,
	// peers who participate
	peers map[peer.ID]struct{},
	// player
	self events.Player,
) GameState {
	switch this.stateType {
	case STATE_WAIT:
		if event.Type() == events.EVENT_START_GAME_VOTE {

			// who started vote is voted
			voteList := map[peer.ID]struct{}{self.ID: struct{}{}}
			vote := Vote{voted: voteList, author: self.ID}
			// next state - vote for game start
			return GameState{stateType: STATE_GAME_VOTE, vote: &vote}

			// TODO TIMEOUT

		}
	// kick
	// else if event.Type() == events.EVENT_START_GAME {

	// 	this.vote.voted[self.ID] = struct{}{}

	// 	if len(peers) == len(this.vote.voted) {
	// 		// All voted

	// 		// CHOOSE MOVE

	// 		return GameState{stateType: STATE_MOVES_EXCHANGE, vote: nil}
	// 	}
	// 	return this
	// }

	case STATE_GAME_VOTE:

	}
}
func timeout() {

}
