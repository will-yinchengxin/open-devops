package route

import (
	"github.com/gin-gonic/gin"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"net/http"
	"openDevops/src/modules/server/web/middleware"
	"time"
)

// Todo 前端交互, 方便 dashboard 操作
func StartGin(httpAddr string, logger log.Logger) error {
	r := gin.New()
	r.Use(gin.Logger())

	// 使用中间件
	m := make(map[string]interface{})
	m["logger"] = logger
	r.Use(middleware.ConfigMiddleware(m))

	// 设置路由
	configRoutes(r)
	s := &http.Server{
		Addr:           httpAddr,
		Handler:        r,
		ReadTimeout:    time.Second * 5,
		WriteTimeout:   time.Second * 5,
		MaxHeaderBytes: 1 << 20,
	}
	level.Info(logger).Log("msg", "web_server_available_at", "httpAddr", httpAddr)

	err := s.ListenAndServe()
	return err
}
