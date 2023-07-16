package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
	"zenquote/internal/quoteapi"

	"zenquote/internal/config"
	"zenquote/internal/logger"
	storage "zenquote/internal/redisdb"
	"zenquote/internal/server/tcp"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

var options = []fx.Option{
	fx.Provide(
		config.New,
		tcp.NewServer,
		tcp.NewHandler,
		logger.New,
		func() *http.Client {
			return http.DefaultClient
		},
		func(cfg config.Config) tcp.HashcashRepo {
			return storage.NewRedisStorage(cfg)
		},
		func(client *http.Client) tcp.ZenquoteRepo {
			return quoteapi.NewQuoteAPI(client)
		},
	),
	fx.Invoke(func(
		lc fx.Lifecycle,
		stop fx.Shutdowner,
		config config.Config,
		logger *zap.Logger,
		server *tcp.Server,
	) {
		lc.Append(fx.Hook{
			OnStart: func(startCtx context.Context) error {
				go server.Start(startCtx, stop)

				return nil
			},
			OnStop: func(stopCtx context.Context) error {
				go server.Shutdown()
				_ = logger.Sync()

				return nil
			},
		})
	}),
	fx.StartTimeout(10 * time.Minute),
	fx.StopTimeout(30 * time.Second),
	fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
		return &fxevent.ZapLogger{Logger: logger}
	}),
}

func main() {
	app := fx.New(options...)

	if app.Err() != nil {
		vis, err := fx.VisualizeError(app.Err())
		if err == nil {
			_, _ = fmt.Fprintln(os.Stderr, vis)
		}

		panic(app.Err())
	}

	app.Run()
}
