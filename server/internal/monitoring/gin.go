package monitoring

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type GinLoggerConfig struct {
	SkipPaths map[string]bool
	Source    string
}

func GinRequestLogger(hub *Hub, cfg GinLoggerConfig) gin.HandlerFunc {
	source := cfg.Source
	if source == "" {
		source = "gin"
	}
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		path := c.Request.URL.Path
		if cfg.SkipPaths[path] {
			return
		}
		if fullPath := c.FullPath(); fullPath != "" {
			path = fullPath
		}
		message := fmt.Sprintf("method=%s path=%s status=%d latency=%s", c.Request.Method, path, c.Writer.Status(), time.Since(start).Truncate(time.Microsecond))
		hub.Append("LOG", source, message)
	}
}
