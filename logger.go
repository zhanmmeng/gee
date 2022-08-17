package gee

import (
	"log"
	"time"
)

func Logger() HandlerFunc {
	return func(c *Context) {
		// Start timer
		t := time.Now()

		// Process request
		c.Next()

		// Calculate resolution time 计算解析时间
		log.Printf("[%d] %s in %v", c.StatusCode,c.Req.RequestURI,time.Since(t))
	}
}
