package core

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type HttpHandler struct {
	l          *zap.SugaredLogger
	httpServer *http.Server
}

func NewHttpHandler() (*HttpHandler, error) {
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
		httpServer: server,
	}, nil
}
