package http

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	stdhttp "net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/geometry-software/gin-products/apps/providers/shared/config"
	"github.com/gin-gonic/gin"
)

// Error represents an HTTP error with a safe public message.
type Error struct {
	Status  int    `json:"-"`
	Message string `json:"error"`
}

// Error returns the safe public message for a typed HTTP error.
func (e *Error) Error() string { return e.Message }

// Fail constructs a typed HTTP error.
func Fail(status int, message string) error { return &Error{status, message} }

// Paging validates pagination parameters.
func Paging(c *gin.Context) (int64, int64, error) {
	page, e1 := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, e2 := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)
	if e1 != nil || e2 != nil || page < 1 || page > 1000000 || limit < 1 || limit > 100 {
		return 0, 0, Fail(400, "page must be 1..1000000 and limit 1..100")
	}
	return page, limit, nil
}

// New constructs a configured Gin router.
func New(service string) *gin.Engine {
	gin.SetMode(config.Get("GIN_MODE", "release"))
	r := gin.New()
	_ = r.SetTrustedProxies(nil)
	r.Use(gin.CustomRecovery(func(c *gin.Context, _ any) { c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"}) }))
	if service != "adapters" {
		r.Use(func(c *gin.Context) {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
			if c.Request.Method == stdhttp.MethodOptions {
				c.AbortWithStatus(stdhttp.StatusNoContent)
				return
			}
			c.Next()
		})
	}
	r.Use(func(c *gin.Context) {
		c.Request.Body = stdhttp.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
		start := time.Now()
		c.Next()
		slog.Info("http", "service", service, "method", c.Request.Method, "route", c.FullPath(), "status", c.Writer.Status(), "duration", time.Since(start))
	})
	return r
}

// Internal requires a scoped internal service credential.
func Internal(allowed ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, name := range allowed {
			if subtle.ConstantTimeCompare([]byte(c.GetHeader("X-Service-Token")), []byte(config.ServiceToken(name))) == 1 {
				c.Set("service", name)
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(401, gin.H{"error": "Invalid service credential"})
	}
}

var Client = &stdhttp.Client{Timeout: 15 * time.Second}

// Call sends a typed HTTP request with an optional scoped service credential.
func Call(ctx context.Context, method, url, caller string, input, output any) error {
	var body io.Reader
	if input != nil {
		data, err := json.Marshal(input)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}
	req, err := stdhttp.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if caller != "" {
		req.Header.Set("X-Service-Token", config.ServiceToken(caller))
	}
	resp, err := Client.Do(req)
	if err != nil {
		return Fail(503, "Upstream service unavailable")
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		var remote Error
		if json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&remote) != nil || remote.Message == "" {
			remote.Message = "Upstream request failed"
		}
		remote.Status = resp.StatusCode
		return &remote
	}
	if output != nil && resp.StatusCode != 204 {
		return json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(output)
	}
	return nil
}

// Run serves HTTP until shutdown and calls the cleanup function.
func Run(service string, r *gin.Engine, cleanup func(context.Context)) {
	server := &stdhttp.Server{Addr: config.Get("BIND_HOST", "127.0.0.1") + ":" + config.Port(service), Handler: r, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 20 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
	}()
	slog.Info("listening", "service", service, "address", server.Addr)
	err := server.ListenAndServe()
	if cleanup != nil {
		end, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cleanup(end)
	}
	if err != nil && !errors.Is(err, stdhttp.ErrServerClosed) {
		slog.Error("server stopped", "service", service, "reason", err)
		os.Exit(1)
	}
}
