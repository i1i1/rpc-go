package main

import (
	"context"
	"encoding/json"

	"github.com/libp2p/go-libp2p-core/peer"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// GameRoomBufSize is the number of incoming messages to buffer for each topic.
const GameRoomBufSize = 128

type (
	// GameRoom represents a subscription to a single PubSub topic. Events
	// can be published to the topic with GameRoom.Publish, and received
	// events are pushed to the Events channel.
	GameRoom struct {
		// Events is a channel of events received from other peers in the game room
		Events chan Event

		ctx   context.Context
		ps    *pubsub.PubSub
		topic *pubsub.Topic
		sub   *pubsub.Subscription

		roomName RoomName
		self     peer.ID
		nick     string
	}

	// ChatMessage gets converted to/from JSON and sent in the body of pubsub messages.
	ChatMessage struct {
		Message    string
		SenderID   string
		SenderNick string
	}

	RoomName string
)

func (name *RoomName) topicName() string {
	return "chat-room:" + string(*name)
}

// JoinGameRoom tries to subscribe to the PubSub topic for the room name, returning
// a GameRoom on success.
func JoinGameRoom(
	ctx context.Context,
	ps *pubsub.PubSub,
	selfID peer.ID,
	nickname string,
	roomName RoomName,
) (*GameRoom, error) {
	// join the pubsub topic
	topic, err := ps.Join(roomName.topicName())
	if err != nil {
		return nil, err
	}

	// and subscribe to it
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	cr := &GameRoom{
		ctx:      ctx,
		ps:       ps,
		topic:    topic,
		sub:      sub,
		self:     selfID,
		nick:     nickname,
		roomName: roomName,
		Events:   make(chan Event, GameRoomBufSize),
	}

	// start reading messages from the subscription in a loop
	go cr.readLoop()
	return cr, nil
}

// Publish sends a message to the pubsub topic.
func (cr *GameRoom) Publish(message string) error {
	m := ChatMessage{
		Message:    message,
		SenderID:   cr.self.Pretty(),
		SenderNick: cr.nick,
	}
	msgBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return cr.topic.Publish(cr.ctx, msgBytes)
}

func (cr *GameRoom) ListPeers() []peer.ID {
	return cr.ps.ListPeers(cr.roomName.topicName())
}

// readLoop pulls messages from the pubsub topic and pushes them onto the Messages channel.
func (cr *GameRoom) readLoop() {
	for {
		msg, err := cr.sub.Next(cr.ctx)
		if err != nil {
			close(cr.Events)
			return
		}
		// only forward events delivered by others
		if msg.ReceivedFrom == cr.self {
			continue
		}

		type intermidiateEvent struct {
			Type EventType
			Data json.RawMessage
		}
		int_event := intermidiateEvent{}
		err = json.Unmarshal(msg.Data, &int_event)
		if err != nil {
			continue
		}

		var event Event

		switch int_event.Type {
		case EVENT_START_GAME_VOTE:
			event = &EventStartGameVote{}
		case EVENT_START_GAME:
			event = &EventStartGame{}
		default:
			continue
		}

		// send valid messages onto the Messages channel
		cr.Events <- event
	}
}
