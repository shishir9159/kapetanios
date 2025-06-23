package main

import (
	"context"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"net/http"
	"runtime"
	"sync"
)

func (t *testConnection) test(w http.ResponseWriter, r *http.Request) {

	t.log.Info("testAddClient called",
		zap.Int("number of go routines", runtime.NumGoroutine()))
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		t.log.Info("connection upgrade error:",
			zap.Error(err))
		return
	}

	client := Client{
		Conn: conn,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			t.log.Info("context done",
				zap.String("remote string", client.Conn.RemoteAddr().String()))
			break

		default:
			t.log.Info("default entry")
			//err := client.Conn.SetReadDeadline(time.Now().Add(20 * time.Second))
			//if err != nil {
			//	log.Println("error setting read deadline for client",
			//		zap.Error(err))
			//	return err
			//}

			msgType, msg, er := client.Conn.ReadMessage()
			if er != nil || msgType == websocket.CloseMessage {
				if websocket.IsCloseError(er) {
					t.log.Info("websocket connection close error",
						zap.String("connection remote address", client.Conn.RemoteAddr().String()))
					log.Println("websocket connection close error", client.Conn.RemoteAddr().String())
				}

				t.log.Info("error reading from %s for messagetype %d: %v - message: %s",
					zap.String("remote string", client.Conn.RemoteAddr().String()),
					zap.String("message type", string(msg)),
					zap.Error(er))
				log.Printf("error reading from %s for messagetype %d: %v - message: %s",
					client.Conn.RemoteAddr().String(), msgType, er, string(msg))
			}

			if msgType != websocket.TextMessage {
				t.log.Info("unexpected message type",
					zap.String("msgType", string(msgType)))
				log.Printf("unexpected message type: %v", msgType)
			}

			log.Printf("received from %s: %s",
				client.Conn.RemoteAddr().String(), string(msg))

			break
		}
	}

}

func main() {

	cfg := zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:   "message",
			LevelKey:     "level",
			EncodeLevel:  zapcore.CapitalLevelEncoder,
			TimeKey:      "time",
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}

	logger := zap.Must(cfg.Build())
	defer func(logger *zap.Logger) {
		er := logger.Sync()
		if er != nil {
			logger.Info("error syncing logger before application terminates",
				zap.Error(er))
		}
	}(logger)

	pool := NewPool()
	go pool.Run(logger)

	t := testConnection{
		mu:   &sync.Mutex{},
		pool: pool,
		log:  logger,
	}

	http.Handle("/test", http.HandlerFunc(t.test))
	t.log.Info("http server started")
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Println(err)
		return
	}
}
