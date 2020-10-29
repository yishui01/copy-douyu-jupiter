package xlog

import (
	"bufio"
	"context"
	"go.uber.org/zap/zapcore"
	"sync"
	"time"
)

type bufferWriterSyncer struct {
	sync.Mutex
	bufferWriter *bufio.Writer
	ticker       *time.Ticker
}

const (
	defaultBufferSize    = 256 * 1024
	defaultFlushInterval = 30 * time.Second
)

type CloseFunc func() error

func Buffer(ws zapcore.WriteSyncer, bufferSize int, flushInterval time.Duration) (zapcore.WriteSyncer, CloseFunc) {
	if _, ok := ws.(*bufferWriterSyncer); ok {
		return ws, func() error {
			return nil
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	if bufferSize == 0 {
		bufferSize = defaultBufferSize
	}

	if flushInterval == 0 {
		flushInterval = defaultFlushInterval
	}

	ticker := time.NewTicker(flushInterval)

	ws = &bufferWriterSyncer{
		bufferWriter: bufio.NewWriterSize(ws, bufferSize),
		ticker:       ticker,
	}

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = ws.Sync()
			case <-ctx.Done():
				return
			}
		}
	}()

	closefunc := func() error {
		cancel()
		return ws.Sync()
	}
	return ws, closefunc
}

func (s *bufferWriterSyncer) Write(bs []byte) (int, error) {
	s.Lock()
	defer s.Unlock()

	if len(bs) > s.bufferWriter.Available() && s.bufferWriter.Buffered() > 0 {
		if err := s.bufferWriter.Flush(); err != nil {
			return 0, err
		}
	}
	return s.bufferWriter.Write(bs)
}

func (s *bufferWriterSyncer) Sync() error {
	s.Lock()
	defer s.Unlock()

	return s.bufferWriter.Flush()
}
