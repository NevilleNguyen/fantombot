package core

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type ErrResponse struct {
	Error   string `json:"error"`
	Success bool   `json:"success"`
}

type HttpHandler struct {
	l          *zap.SugaredLogger
	router     *gin.Engine
	httpServer *http.Server
	core       *Core
}

func NewHttpHandler(core *Core) (*HttpHandler, error) {
	l := zap.S()
	router := gin.New()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.MaxAge = 5 * time.Minute

	router.Use(cors.New(corsConfig))

	port := viper.GetInt("http.port")
	if err := validation.Validate(port, validation.Required); err != nil {
		l.Errorw("error get http port", "error", err)
		return nil, err
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	return &HttpHandler{
		l:          zap.S(),
		router:     router,
		httpServer: server,
		core:       core,
	}, nil
}

func (h *HttpHandler) Run() {
	api := h.router.Group("/api")
	api.POST("/addChatGroup", h.AddChatGroupApi)
	api.GET("/removeChatGroup", h.DeleteChatGroupApi)

	if err := h.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		h.l.Panicw("run http server error", "error", err)
	}
}

func (h *HttpHandler) AddChatGroupApi(c *gin.Context) {
	var req struct {
		Token  string `json:"token"`
		ChatId int64  `json:"chat_id"`
	}

	if err := c.BindJSON(req); err != nil {
		c.JSON(
			http.StatusBadRequest,
			ErrResponse{
				Error: err.Error(),
			},
		)
	}

	if err := h.core.AddChatGroup(req.Token, req.ChatId); err != nil {
		c.JSON(
			http.StatusUnprocessableEntity,
			ErrResponse{
				Error: err.Error(),
			},
		)
	}
	c.JSON(
		http.StatusOK,
		ErrResponse{
			Success: true,
		},
	)
}

func (h *HttpHandler) DeleteChatGroupApi(c *gin.Context) {
	chatIdStr := c.Query("chat_id")
	chatId, err := strconv.ParseInt(chatIdStr, 10, 64)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			ErrResponse{
				Error: err.Error(),
			},
		)
	}
	if err := h.core.DeleteChatGroup(chatId); err != nil {
		c.JSON(
			http.StatusUnprocessableEntity,
			ErrResponse{
				Error: err.Error(),
			},
		)
	}
	c.JSON(
		http.StatusOK,
		ErrResponse{
			Success: true,
		},
	)
}
