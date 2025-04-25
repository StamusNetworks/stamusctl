package auth

import (
	"context"
	"encoding/base64"
	"strings"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/logging"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/spf13/afero"
)

var token string

func WatchForToken(pathToWatch string) {
	ctx, span := logging.Tracer.Start(context.Background(), "watchForToken")
	defer span.End()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logging.LoggerWithContextToSpanContext(ctx).Sugar().Fatal(err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					logging.LoggerWithContextToSpanContext(ctx).Info("token file updated. Updating token")
					tokenFromFile, err := afero.ReadFile(app.FS, pathToWatch)
					if err != nil {
						logging.LoggerWithContextToSpanContext(ctx).Sugar().Error("failed to read token file", err)
					}
					token = string(tokenFromFile)
					logging.LoggerWithContextToSpanContext(ctx).Info("updated token")
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logging.LoggerWithContextToSpanContext(ctx).Sugar().Error("err:", err)
			}
		}
	}()
	err = watcher.Add(pathToWatch)
	if err != nil {
		logging.LoggerWithContextToSpanContext(ctx).Sugar().Fatal(err)
	}

	<-make(chan struct{})
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if token == "" {
			c.Next()
			return
		}
		if c.Request.Header.Get("authorization") == "" {
			c.AbortWithStatusJSON(401, gin.H{"message": "no token provided. token needed"})
			return
		}
		header := c.Request.Header.Get("authorization")
		split := strings.Split(header, " ")
		if len(split) != 2 {
			c.AbortWithStatusJSON(401, gin.H{"message": "bad token formating"})
			return
		}
		if split[0] != "Basic" {
			c.AbortWithStatusJSON(401, gin.H{"message": "header is not in Basic auth format"})
			return
		}

		data, err := base64.StdEncoding.DecodeString(split[1])
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"message": "decoding of the token failed"})
			return
		}

		split = strings.Split(string(data), ":")
		if len(split) != 2 {
			c.AbortWithStatusJSON(401, gin.H{"message": "bad token formating after decoding"})
			return
		}

		requestToken := split[1]
		if requestToken != token {
			c.AbortWithStatusJSON(403, gin.H{"message": "bad token"})
			return
		}
	}
}
