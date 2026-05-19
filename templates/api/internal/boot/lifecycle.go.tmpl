package boot

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// Run 启动 HTTP 服务并在收到上游取消信号后优雅关闭。
func (a *Application) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		if err := a.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		timeout := a.Config.Server.ShutdownTimeout
		if timeout <= 0 {
			timeout = 10 * time.Second
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		return a.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

// Shutdown 优雅关闭 HTTP Server 与底层资源。
func (a *Application) Shutdown(ctx context.Context) error {
	var firstErr error

	if err := a.Server.Shutdown(ctx); err != nil && firstErr == nil {
		firstErr = err
	}
	if err := a.DB.Close(); err != nil && firstErr == nil {
		firstErr = err
	}
	if a.Publisher != nil {
		if err := a.Publisher.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if a.Redis != nil {
		if err := a.Redis.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if err := a.Logger.Sync(); err != nil && firstErr == nil {
		firstErr = err
	}

	return firstErr
}
