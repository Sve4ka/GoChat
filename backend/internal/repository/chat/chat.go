package chat

import (
	"backend/internal/models"
	"backend/internal/repository"
	"backend/pkg/cerr"
	"backend/pkg/postgres"
	"context"
)

type RepoChat struct {
	db *postgres.Pg
}

func InitChatRepository(db *postgres.Pg) repository.ChatRepo {
	return RepoChat{db: db}
}

func (r RepoChat) Create(ctx context.Context, chatName string, userID int) (int, error) {
	var id int
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return 0, cerr.Transaction(err)
	}
	row := tx.QueryRow(ctx, "INSERT INTO chat (name) VALUES ($1) returning id", chatName)
	err = row.Scan(&id)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return 0, cerr.Rollback(err)
		}
		return 0, cerr.Scan(err)
	}
	row = tx.QueryRow(ctx, "INSERT INTO chat_user (id_user, id_chat) VALUES ($1, $2) returning id_chat", userID, id)
	err = row.Scan(&id)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return 0, cerr.Rollback(err)
		}
		return 0, cerr.Scan(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return 0, cerr.Commit(err)
	}
	return id, nil
}

func (r RepoChat) AddUser(ctx context.Context, chatID int, userID int) (int, error) {
	row := r.db.Pool.QueryRow(ctx, "SELECT count(*) from chat_user where id_user = $1 and id_chat = $2", userID, chatID)
	var cnt int
	err := row.Scan(&cnt)
	if err != nil {
		return 0, cerr.Scan(err)
	}
	if cnt != 0 {
		return chatID, nil
	}
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return 0, cerr.Transaction(err)
	}
	row = tx.QueryRow(ctx, "INSERT INTO  chat_user (id_user, id_chat) VALUES ($1, $2)  returning id_chat", userID, chatID)
	err = row.Scan(&chatID)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return 0, cerr.Rollback(err)
		}
		return 0, cerr.Scan(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return 0, cerr.Commit(err)
	}
	return chatID, nil
}

func (r RepoChat) Message(ctx context.Context, message models.MessageBase) (*models.Message, error) {
	var id int
	var newMessage models.Message
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return nil, cerr.Transaction(err)
	}

	row := r.db.Pool.QueryRow(ctx, "SELECT name FROM users WHERE id = $1", message.SenderID)
	err = row.Scan(&message.SenderName)
	if err != nil {
		return nil, cerr.Scan(err)
	}
	row = tx.QueryRow(ctx, "INSERT INTO messages (id_user, name_user, text, time) values ($1, $2, $3, $4) returning id", message.SenderID, message.SenderName, message.Text, message.Timestamp)
	err = row.Scan(&id)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return nil, cerr.Rollback(err)
		}
		return nil, cerr.Scan(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, cerr.Commit(err)
	}
	newMessage = models.Message{
		MessageBase: message,
		ID:          id,
	}
	return &newMessage, nil
}

func (r RepoChat) Chat(ctx context.Context, chatID int) (*models.AllChat, error) {
	var chat models.AllChat
	row := r.db.Pool.QueryRow(ctx, "SELECT name FROM chat WHERE id = $1", chatID)
	err := row.Scan(&chat.Name)
	if err != nil {
		return nil, cerr.Scan(err)
	}
	chat.ID = chatID
	rows, err := r.db.Pool.Query(ctx, "SELECT id, id_user, name_user, text, time, type from messages where id_chat = $1 order by id", chatID)
	if err != nil {
		return nil, cerr.ExecContext(err)
	}
	for rows.Next() {
		var message models.Message
		err = rows.Scan(&message.ID, &message.SenderID, &message.SenderName, &message.Text, &message.Timestamp, &message.Type)
		if err != nil {
			return nil, cerr.Scan(err)
		}
		chat.Messages = append(chat.Messages, message)
	}
	return &chat, nil
}

func (r RepoChat) ChatMessage(ctx context.Context, message models.MessageChatBase) (*models.MessageChat, error) {
	var id int
	var newMessage models.MessageChat
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return nil, cerr.Transaction(err)
	}

	row := r.db.Pool.QueryRow(ctx, "SELECT name FROM users WHERE id = $1", message.SenderID)
	err = row.Scan(&message.SenderName)
	if err != nil {
		return nil, cerr.Scan(err)
	}
	row = tx.QueryRow(ctx, "INSERT INTO messages (id_user, name_user, text, time, id_chat, type) values ($1, $2, $3, $4, $5, $6) returning id", message.SenderID, message.SenderName, message.Text, message.Timestamp, message.ChatID, message.Type)
	err = row.Scan(&id)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return nil, cerr.Rollback(err)
		}
		return nil, cerr.Scan(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, cerr.Commit(err)
	}
	newMessage = models.MessageChat{
		ChatID: message.ChatID,
		Message: models.Message{
			ID:          id,
			MessageBase: message.MessageBase,
		},
	}
	return &newMessage, nil

}

func (r RepoChat) GetAllMessage(ctx context.Context) ([]models.Message, error) {
	var messages []models.Message
	rows, err := r.db.Pool.Query(ctx, "SELECT id, id_user, name_user, text, time from messages")
	if err != nil {
		return nil, cerr.ExecContext(err)
	}
	for rows.Next() {
		var message models.Message
		err = rows.Scan(&message.ID, &message.SenderID, &message.SenderName, &message.Text, &message.Timestamp)
		if err != nil {
			return nil, cerr.Scan(err)
		}
		messages = append(messages, message)
	}
	return messages, nil

}

func (r RepoChat) GetAllChats(ctx context.Context) ([]models.Chat, error) {
	var chats []models.Chat
	rows, err := r.db.Pool.Query(ctx, "SELECT id, name from chat")
	if err != nil {
		return nil, cerr.ExecContext(err)
	}
	for rows.Next() {
		var chat models.Chat
		err = rows.Scan(&chat.ID, &chat.Name)
		if err != nil {
			return nil, cerr.Scan(err)
		}
		chats = append(chats, chat)
	}
	return chats, nil
}

func (r RepoChat) GetUsersChats(ctx context.Context, userID int) ([]models.Chat, error) {
	var chats []models.Chat
	rows, err := r.db.Pool.Query(ctx, "SELECT id_chat from chat_user where id_user = $1", userID)
	if err != nil {
		return nil, cerr.ExecContext(err)
	}
	for rows.Next() {
		var chat models.Chat
		err = rows.Scan(&chat.ID)
		if err != nil {
			return nil, cerr.Scan(err)
		}
		row := r.db.Pool.QueryRow(ctx, "SELECT name from chat where id = $1", chat.ID)
		err = row.Scan(&chat.Name)
		if err != nil {
			return nil, cerr.Scan(err)
		}
		chats = append(chats, chat)
	}
	return chats, nil

}
