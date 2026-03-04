package handlers

import (
	"backend/internal/models"
	"backend/internal/repository"
	"backend/pkg/log"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type ChatHandler struct {
	repo  repository.ChatRepo
	chats map[int]map[string]*websocket.Conn
	mutex sync.Mutex
}

func InitChatHandler(repo repository.ChatRepo) ChatHandler {
	chats := make(map[int]map[string]*websocket.Conn)

	return ChatHandler{repo: repo, chats: chats}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	HandshakeTimeout: 60 * time.Second,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *ChatHandler) WSEndpoint(c *gin.Context) {
	id := c.Param("id")
	chatID, err := strconv.Atoi(id)
	if err != nil {
		log.Log.Error(err)
		c.JSON(418, gin.H{"error": err.Error()})
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Log.Error(err)
		return
	}
	clientID := uuid.New().String()

	conn.SetReadLimit(1024 * 1024)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	h.mutex.Lock()
	if h.chats[chatID] == nil {
		h.chats[chatID] = make(map[string]*websocket.Conn)
	}
	h.chats[chatID][clientID] = conn
	h.mutex.Unlock()
	done := make(chan struct{})
	h.reader(conn, clientID, done, chatID)

	go h.writePump(conn, clientID, done)
}

func (h *ChatHandler) reader(conn *websocket.Conn, clientID string, done chan<- struct{}, chatID int) {
	defer func() {
		h.mutex.Lock()
		delete(h.chats[chatID], clientID)
		h.mutex.Unlock()
		conn.Close()
		close(done)
	}()

	for {
		var getMSG models.MessageGet
		var msg models.MessageBase
		err := conn.ReadJSON(&getMSG)
		if err != nil {
			log.Log.Error(err)
			break
		}

		parsedTime, err := time.Parse("Mon Jan 02 2006 15:04:05 GMT-0700 (Москва, стандартное время)", getMSG.Timestamp)
		parsedTime1, err1 := time.Parse("Mon Jan 02 2006 15:04:05 GMT-0700  (Moscow Standard Time)", getMSG.Timestamp)
		if err != nil && err1 != nil {
			log.Log.Error(err)
			conn.WriteJSON(map[string]string{"error": "Invalid timestamp"})
			continue
		}
		if err != nil {
			parsedTime = parsedTime1
		}

		msg = models.MessageBase{
			Type:       getMSG.Type,
			Text:       getMSG.Text,
			SenderID:   getMSG.SenderID,
			SenderName: getMSG.SenderName,
			Timestamp:  parsedTime,
		}
		if msg.Text == "" {
			conn.WriteJSON(map[string]string{"error": "Empty message"})
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		chatMsg := models.MessageChatBase{
			ChatID:      chatID,
			MessageBase: msg,
		}
		newMsg, err := h.repo.ChatMessage(ctx, chatMsg)
		if err != nil {
			conn.WriteJSON(map[string]string{"error": "Cant save message"})
			continue
		}

		h.writer(newMsg, clientID)
	}
}

func (h *ChatHandler) writer(newMsg *models.MessageChat, senderID string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	var wg sync.WaitGroup
	for _, client := range h.chats[newMsg.ChatID] {
		wg.Add(1)
		go h.send(client, newMsg, &wg)
	}
	wg.Wait()
}

func (h *ChatHandler) send(conn *websocket.Conn, message *models.MessageChat, wg *sync.WaitGroup) {
	conn.WriteJSON(message)
	wg.Done()
}

func (h *ChatHandler) writePump(conn *websocket.Conn, clientID string, done <-chan struct{}) {
	ticker := time.NewTicker(50 * time.Second)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second))
			if err != nil {
				log.Log.Error(err)
				return
			}
		case <-done:
			return
		}
	}
}

func (h *ChatHandler) GetOldMessages(g *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messages, err := h.repo.GetAllMessage(ctx)

	if err != nil {
		log.Log.Error(err)
		g.JSON(418, gin.H{"error": err.Error()})
		return
	}

	g.JSON(http.StatusOK, gin.H{"messages": messages})
}

func (h *ChatHandler) GetChats(g *gin.Context) {
	id := g.Param("id")
	aid, err := strconv.Atoi(id)
	if err != nil {
		log.Log.Error(err)
		g.JSON(418, gin.H{"error": err.Error()})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	chats, err := h.repo.GetUsersChats(ctx, aid)

	if err != nil {
		log.Log.Error(err)
		g.JSON(418, gin.H{"error": err.Error()})
		return
	}
	if len(chats) == 0 {
		var newChat models.Chat
		newChat.ID, err = h.repo.Create(ctx, "Избранное", aid)
		if err != nil {
			log.Log.Error(err)
			g.JSON(418, gin.H{"error": err.Error()})
			return
		}
		newChat.Name = "Избранное"
		chats = append(chats, newChat)
	}
	g.JSON(http.StatusOK, gin.H{"chats": chats})
}

func (h *ChatHandler) GetMessages(g *gin.Context) {
	id := g.Param("id")
	aid, err := strconv.Atoi(id)
	if err != nil {
		log.Log.Error(err)
		g.JSON(418, gin.H{"error": err.Error()})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messages, err := h.repo.Chat(ctx, aid)
	if err != nil {
		log.Log.Error(err)
		g.JSON(418, gin.H{"error": err.Error()})
	}
	var out []models.MessageBase
	for _, elem := range messages.Messages {
		out = append(out, elem.MessageBase)
	}
	g.JSON(http.StatusOK, gin.H{"messages": out})
}

func (h *ChatHandler) CreateChat(g *gin.Context) {
	id := g.Param("userID")
	aid, err := strconv.Atoi(id)
	if err != nil {
		log.Log.Error(err)
		g.JSON(418, gin.H{"error": err.Error()})
	}
	var newChat models.NewChat

	if err := g.ShouldBindJSON(&newChat); err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	chatID, err := h.repo.Create(ctx, newChat.Name, aid)
	if err != nil {
		log.Log.Error(err)
		g.JSON(418, gin.H{"error": err.Error()})
	}
	var chat models.Chat
	chat.ID = chatID
	chat.Name = newChat.Name
	g.JSON(http.StatusOK, chat)
}

func (h *ChatHandler) AddUser(g *gin.Context) {
	idChat := g.Param("idChat")
	chatID, err := strconv.Atoi(idChat)
	if err != nil {
		log.Log.Error(err)
		g.JSON(418, gin.H{"error": err.Error()})
	}
	idUser := g.Param("idUser")
	userID, err := strconv.Atoi(idUser)
	if err != nil {
		log.Log.Error(err)
		g.JSON(418, gin.H{"error": err.Error()})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	id, err := h.repo.AddUser(ctx, chatID, userID)
	if err != nil {
		log.Log.Error(err)
		g.JSON(418, gin.H{"error": err.Error()})
	}
	g.JSON(http.StatusOK, gin.H{"chat": id})

}
