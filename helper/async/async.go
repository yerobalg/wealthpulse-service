package async

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/yerobalg/wealthpulse-service/helper/logger"
)

type Async struct {
	log logger.Interface
}

type Interface interface {
	Run(ctx context.Context, fn func())
}

func Init(log logger.Interface) Interface {
	return &Async{log: log}
}

// Run executes the given function in a new goroutine with panic recovery.
// The goroutine runs in the background and does not block the caller.

func (a *Async) Run(ctx context.Context, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				a.log.Error(ctx, fmt.Sprintf("recovered from panic in goroutine: %v", r), string(debug.Stack()))
			}
		}()
		fn()
	}()
}
