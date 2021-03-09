package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/i1i1/rpc-go/pkg/events"
	"github.com/i1i1/rpc-go/pkg/game"

	"github.com/gdamore/tcell/v2"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/rivo/tview"
)

// GameUI is a Text User Interface (TUI) for a GameRoom.
// The Run method will draw the UI to the terminal in "fullscreen"
// mode. You can quit with Ctrl-C, or by typing "/quit" into the
// chat prompt.
type GameUI struct {
	gr        *game.GameRoom
	app       *tview.Application
	peersList *tview.TextView

	msgW    io.Writer
	inputCh chan string
	doneCh  chan struct{}
}

// NewGameUI returns a new GameUI struct that controls the text UI.
// It won't actually do anything until you call Run().
func NewGameUI(gr *game.GameRoom) *GameUI {
	app := tview.NewApplication()

	// make a text view to contain our chat messages
	msgBox := tview.NewTextView()
	msgBox.SetDynamicColors(true)
	msgBox.SetBorder(true)
	msgBox.SetTitle(fmt.Sprintf("Room: %s", gr.RoomName))

	// text views are io.Writers, but they don't automatically refresh.
	// this sets a change handler to force the app to redraw when we get
	// new messages to display.
	msgBox.SetChangedFunc(func() {
		app.Draw()
	})

	// an input field for typing messages into
	inputCh := make(chan string, 32)
	input := tview.NewInputField().
		SetLabel(gr.Self.Nick + " > ").
		SetFieldWidth(0).
		SetFieldBackgroundColor(tcell.ColorBlack)

	// the done func is called when the user hits enter, or tabs out of the field
	input.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			// we don't want to do anything if they just tabbed away
			return
		}
		line := input.GetText()
		if len(line) == 0 {
			// ignore blank lines
			return
		}

		// bail if requested
		if line == "/quit" {
			app.Stop()
			return
		}

		// send the line onto the input chan and reset the field text
		inputCh <- line
		input.SetText("")
	})

	// make a text view to hold the list of peers in the room, updated by ui.refreshPeers()
	peersList := tview.NewTextView()
	peersList.SetBorder(true)
	peersList.SetTitle("Peers")
	peersList.SetChangedFunc(func() { app.Draw() })

	// chatPanel is a horizontal box with messages on the left and peers on the right
	// the peers list takes 20 columns, and the messages take the remaining space
	chatPanel := tview.NewFlex().
		AddItem(msgBox, 0, 1, false).
		AddItem(peersList, 20, 1, false)

	// flex is a vertical box with the chatPanel on top and the input field at the bottom.

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(chatPanel, 0, 1, false).
		AddItem(input, 1, 1, true)

	app.SetRoot(flex, true)

	return &GameUI{
		gr:        gr,
		app:       app,
		peersList: peersList,
		msgW:      msgBox,
		inputCh:   inputCh,
		doneCh:    make(chan struct{}, 1),
	}
}

// Run starts the chat event loop in the background, then starts
// the event loop for the text UI.
func (ui *GameUI) Run() error {
	go ui.handleEvents()
	defer ui.end()

	return ui.app.Run()
}

// end signals the event loop to exit gracefully
func (ui *GameUI) end() {
	ui.doneCh <- struct{}{}
}

// ShortID returns the last 8 chars of a base58-encoded peer id.
func ShortID(p peer.ID) string {
	pretty := p.Pretty()
	return pretty[len(pretty)-8:]
}

// refreshPeers pulls the list of peers currently in the chat room and
// displays the last 8 chars of their peer id in the Peers panel in the ui.
func (ui *GameUI) refreshPeers() {
	peers := ui.gr.ListPeers()

	// clear is not threadsafe so we need to take the lock.
	ui.peersList.Lock()
	ui.peersList.Clear()
	ui.peersList.Unlock()

	for p := range peers {
		fmt.Fprintln(ui.peersList, ShortID(p))
	}

	ui.app.Draw()
}

// displayEvent writes a Event from the room to the message window,
// with the sender's nick highlighted in green.
func (ui *GameUI) displayEvent(ev events.Event) {
	prompt := withColor("green", fmt.Sprintf("<%s>:", ev.From().Nick))
	fmt.Fprintf(ui.msgW, "%s %s\n", prompt, ev.String())
}

// displaySelfMessage writes a message from ourself to the message window,
// with our nick highlighted in yellow.
func (ui *GameUI) displaySelfEvent(ev events.Event) {
	prompt := withColor("yellow", fmt.Sprintf("<%s>:", ev.From().Nick))
	fmt.Fprintf(ui.msgW, "%s %s\n", prompt, ev.String())
}

// handleEvents runs an event loop that sends user input to the chat room
// and displays messages received from the chat room. It also periodically
// refreshes the list of peers in the UI.
func (ui *GameUI) handleEvents() {
	peerRefreshTicker := time.NewTicker(time.Second)
	defer peerRefreshTicker.Stop()

	for {
		var cmd []string
		var input string

		select {
		case input = <-ui.inputCh:
			cmd = strings.Fields(input)
		case ev := <-ui.gr.Events:
			// when we receive a message from the chat room, print it to the message window
			ui.displayEvent(ev)
			continue
		case <-peerRefreshTicker.C:
			// refresh the list of peers in the chat room periodically
			ui.refreshPeers()
			continue
		case <-ui.gr.Ctx.Done():
			return
		case <-ui.doneCh:
			return
		}

		var event events.Event

		switch cmd[0] {
		case "/start_game_vote":
			event = events.NewStartGameVote(ui.gr.Self)
		case "/start_game":
			event = events.NewStartGame(ui.gr.Self)
		default:
			if cmd[0][0] == '/' {
				if cmd[0] == "/help" {
					fmt.Fprintf(os.Stderr, "TODO: list all commands")
				} else {
					fmt.Fprintf(os.Stderr, "Unknown command %v. Print /help to see all commands", cmd[0])
				}
				continue
			} else {
				event = events.NewMessage(ui.gr.Self, input)
			}
		}

		err := ui.gr.Publish(event)
		if err != nil {
			fmt.Fprintf(os.Stderr, "publish error: %s", err)
		}
		ui.displaySelfEvent(event)
	}
}

// withColor wraps a string with color tags for display in the messages text box.
func withColor(color, msg string) string {
	return fmt.Sprintf("[%s]%s[-]", color, msg)
}
