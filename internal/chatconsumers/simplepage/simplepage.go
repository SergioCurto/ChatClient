package simplepage

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/SergioCurto/ChatClient/config"
	"github.com/SergioCurto/ChatClient/internal/chatmodels"
	"github.com/a-h/templ"
	"github.com/gorilla/websocket"
)

// SimplePageConsumer is a ChatConsumer that logs messages to an HTML page.
type SimplePageConsumer struct {
	Name         string
	messages     chan chatmodels.ChatMessage
	wsClients    map[*websocket.Conn]bool
	wsClientsMux sync.Mutex
	upgrader     websocket.Upgrader
	// history allows the page to show some of the most recent messages on page reload
	messageHistory []chatmodels.ChatMessage
	historyMutex   sync.Mutex
}

func NewSimplePageConsumer() *SimplePageConsumer {
	return &SimplePageConsumer{
		Name:      "SimplePage",
		messages:  make(chan chatmodels.ChatMessage),
		wsClients: make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		messageHistory: make([]chatmodels.ChatMessage, 0, 30),
	}
}

func (c *SimplePageConsumer) Consume(message chatmodels.ChatMessage) {
	c.messages <- message
	c.addToHistory(message)
}

func (c *SimplePageConsumer) GetName() string {
	return c.Name
}

// index is a templ.Component that renders the HTML page.
type index struct {
	messages []chatmodels.ChatMessage
}

// Render implements the templ.Component interface.
func (i index) Render(ctx context.Context, w io.Writer) error {
	// Creating a chat box with dark terminal like colors
	/// flex space used to manage the chat messages
	_, err := fmt.Fprint(w, `<!DOCTYPE html>
	<html lang="en">
		<head>
			<meta charset="UTF-8"/>
			<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
			<title>Chat Client</title>
			<style>
				body {
					font-family: sans-serif;
					overflow: hidden;
					background-color: #18181b !important;
					color: lightgray;
				}
				#chatbox {
					height: 95vh;
					overflow-y: auto;
					display: flex;
    				flex-direction: column;
					align-content: flex-end;
				}
				#fill {
					flex-grow: 1;
				}
				.message {
					margin-bottom: 5px;
					width: 100%;
				}
				.messagecontainer {
					display: flex;
					flex-direction: row;
				}
				.provider {
					min-width: 70px !important;
					overflow: hidden;
					text-overflow: ellipsis;
					white-space: nowrap;
				}
				.user {
					min-width: 100px !important;
					overflow: hidden;
					text-overflow: ellipsis;
					white-space: nowrap;
					text-align: right;
					padding-right: 10px;
				}
				.messagecontents {
					flex-grow: 1;
				}
				
			</style>
		</head>
		<body>
			<div id="chatbox">
				<div id="fill"></div>`)
	if err != nil {
		return err
	}

	for _, message := range i.messages {
		_, err = fmt.Fprintf(w, `<div class="message"><div class="messagecontainer"><div class="provider">%s</div><div class="user">%s:</div><div class="messagecontents">%s</div></div></div>`, message.Provider, message.AuthorName, message.Content)
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprint(w, `</div>
			<script>
				const chatbox = document.getElementById('chatbox');
				const ws = new WebSocket('ws://' + window.location.host + '/ws');

				ws.onmessage = (event) => {
					const message = JSON.parse(event.data);
					const messageElement = document.createElement('div');
					messageElement.classList.add('message');

					const container = document.createElement('div');
					container.classList.add('messagecontainer');
					messageElement.appendChild(container);

					const provider = document.createElement('div');
					provider.classList.add('provider');
					provider.textContent = message.Provider;

					const user = document.createElement('div');
					user.classList.add('user');
					user.textContent = message.AuthorName+":";

					const messageContents = document.createElement('div');
					messageContents.classList.add('messagecontents');
					messageContents.textContent = message.Content;
					
					container.appendChild(provider);
					container.appendChild(user);
					container.appendChild(messageContents);
					
					chatbox.appendChild(messageElement);
					chatbox.scrollTop = chatbox.scrollHeight;
				};
			</script>
		</body>
	</html>`)
	return err
}

func (c *SimplePageConsumer) Start(cfg *config.Config) error {
	// Use the index component directly
	http.Handle("/", templ.Handler(index{messages: c.getHistory()}))
	http.HandleFunc("/ws", c.handleConnections)

	go c.handleMessages()

	log.Println("HTTP server started on :", cfg.WebpageOutputPort)
	err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.WebpageOutputPort), nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *SimplePageConsumer) handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := c.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer ws.Close()

	c.wsClientsMux.Lock()
	c.wsClients[ws] = true
	c.wsClientsMux.Unlock()

	// Send the history to the new client
	c.sendHistoryToClient(ws)

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			c.wsClientsMux.Lock()
			delete(c.wsClients, ws)
			c.wsClientsMux.Unlock()
			break
		}
	}
}

func (c *SimplePageConsumer) handleMessages() {
	for msg := range c.messages {
		c.broadcastMessage(msg)
	}
}

func (c *SimplePageConsumer) broadcastMessage(message chatmodels.ChatMessage) {
	c.wsClientsMux.Lock()
	defer c.wsClientsMux.Unlock()
	for client := range c.wsClients {
		err := client.WriteJSON(message)
		if err != nil {
			log.Printf("error: %v", err)
			client.Close()
			delete(c.wsClients, client)
		}
	}
}

func (c *SimplePageConsumer) addToHistory(message chatmodels.ChatMessage) {
	c.historyMutex.Lock()
	defer c.historyMutex.Unlock()

	if len(c.messageHistory) >= 30 {
		c.messageHistory = c.messageHistory[1:]
	}
	c.messageHistory = append(c.messageHistory, message)
}

func (c *SimplePageConsumer) getHistory() []chatmodels.ChatMessage {
	c.historyMutex.Lock()
	defer c.historyMutex.Unlock()

	// Create a copy to avoid race conditions
	historyCopy := make([]chatmodels.ChatMessage, len(c.messageHistory))
	copy(historyCopy, c.messageHistory)
	return historyCopy
}

func (c *SimplePageConsumer) sendHistoryToClient(ws *websocket.Conn) {
	history := c.getHistory()
	for _, msg := range history {
		err := ws.WriteJSON(msg)
		if err != nil {
			log.Printf("error sending history: %v", err)
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}
