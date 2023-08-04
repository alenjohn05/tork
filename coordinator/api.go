package coordinator

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/tork/scheduler"
	"github.com/tork/task"
)

type api struct {
	server    *http.Server
	scheduler scheduler.Scheduler
}

func newAPI(cfg Config) *api {
	if cfg.Address == "" {
		cfg.Address = ":3000"
	}
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	s := &api{
		scheduler: cfg.Scheduler,
		server: &http.Server{
			Addr:    cfg.Address,
			Handler: r,
		},
	}
	r.GET("/status", s.status)
	r.POST("/task", s.submitTask)
	return s
}

func (s *api) status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "OK",
	})
}

func (s *api) submitTask(c *gin.Context) {
	t := task.Task{}
	if err := c.BindJSON(&t); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	t.State = task.Pending
	log.Info().Any("task", t).Msg("received task")
	if err := s.scheduler.Schedule(c, t); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
}

func (s *api) start() error {
	go func() {
		// service connections
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msgf("error starting up server")
		}
	}()
	return nil
}

func (s *api) shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
