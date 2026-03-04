package routers

import (
	"backend/internal/delivery/handlers"
	"backend/internal/repository/chat"
	"backend/pkg/postgres"
	"github.com/gin-gonic/gin"
)

func RegisterChatRouter(r *gin.Engine, db *postgres.Pg) *gin.RouterGroup {
	chatRouter := r.Group("/ws")

	chatRepo := chat.InitChatRepository(db)
	chatHandler := handlers.InitChatHandler(chatRepo)
	chatRouter.POST("/chat/:userID", chatHandler.CreateChat)
	chatRouter.GET("/chat/user/:id", chatHandler.GetChats)
	chatRouter.GET("/chat/:id", chatHandler.WSEndpoint)
	chatRouter.GET("/chat/messages/:id", chatHandler.GetMessages)
	chatRouter.POST("/chat/add/:idChat/:idUser", chatHandler.AddUser)
	return chatRouter
}
