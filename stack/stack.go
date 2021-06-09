package stack

import (
	"fmt"
	"math/rand"
	"os"
	"runtime/debug"
	"time"

	"context"
	"sync"
)

func RecordDump() {
	if err := recover(); err != nil {
		rand := rand.New(rand.NewSource(time.Now().UnixNano()))
		rvalue := rand.Intn(100000000)
		now := time.Now()
		filename := fmt.Sprintf("core_%d%02d%02d%02d%02d%02d%d_%d.dump", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(),
			now.Second(), now.Nanosecond(), rvalue)

		var f *os.File

		if _, err := os.Stat(filename); os.IsNotExist(err) {
			f, _ = os.Create(filename)
		} else {
			f, _ = os.OpenFile(filename, os.O_APPEND, 0666)
		}

		defer f.Close()
		f.Write([]byte(err.(error).Error()))
		f.Write(debug.Stack())
	}
}

func SafeCall(callback func(args ...interface{}) error, iargs ...interface{}) error {
	defer RecordDump()
	return callback(iargs...)
}

func SafeCallNull(callback func()) {
	defer RecordDump()
	callback()
}

func SafeGo(f func() error) {
	go func() {
		defer RecordDump()
		f()
	}()
}

type Group struct {
	cancel func()

	wg sync.WaitGroup

	errOnce sync.Once
	err     error
}

func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{cancel: cancel}, ctx
}

func (g *Group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}

	return g.err
}

func (g *Group) Go(f func() error) {
	g.wg.Add(1)
	go func() {
		defer RecordDump()
		if err := f(); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
		}
	}()
}
