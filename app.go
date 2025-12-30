package main

import (
	"context"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var globalApp *App

// App struct
type App struct {
	ctx      context.Context
	running  bool
	mu       sync.Mutex
	stopChan chan struct{}
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	globalApp = a
	// 启动全局快捷键监听 (仅当有权限时)
	startGlobalListener()
}

func (a *App) triggerShortcut() {
	runtime.EventsEmit(a.ctx, "shortcut-pressed", "f8")
}

// StartClicking starts the auto clicker
func (a *App) StartClicking(interval int, mode string, keyOrButton string, clickType string, longPressDuration int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.running {
		return
	}

	a.running = true
	a.stopChan = make(chan struct{})

	// Mouse Button Logic
	var btnCode int
	if mode == "mouse" {
		switch keyOrButton {
		case "right":
			btnCode = 1
		case "center":
			btnCode = 2
		case "side1":
			btnCode = 3
		case "side2":
			btnCode = 4
		default:
			btnCode = 0
		}
	}

	go func() {
		// Global delay for both Clicker and Presser modes
		// This prevents shortcut interference (e.g. Ctrl+F8 being interpreted as Ctrl+Click)
		// and gives user time to release keys.
		time.Sleep(1000 * time.Millisecond)

		// Check if stopped during sleep
		select {
		case <-a.stopChan:
			return
		default:
		}

		if clickType == "hold" {

			// Hold mode: Press Down -> Wait Stop -> Press Up
			if mode == "mouse" {
				mouseHold(btnCode, true)
			} else {
				keyHold(keyOrButton, true)
			}

			<-a.stopChan

			if mode == "mouse" {
				mouseHold(btnCode, false)
			} else {
				keyHold(keyOrButton, false)
			}
			return
		}

		for {
			// Check stop before action
			select {
			case <-a.stopChan:
				return
			default:
			}

			// Perform Action
			if mode == "mouse" {
				click(btnCode, clickType, longPressDuration)
			} else if mode == "keyboard" {
				pressKey(keyOrButton, clickType, longPressDuration)
			}

			// Wait Interval
			select {
			case <-a.stopChan:
				return
			case <-time.After(time.Duration(interval) * time.Millisecond):
			}
		}
	}()
}

// StopClicking stops the auto clicker
func (a *App) StopClicking() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running {
		return
	}

	a.running = false
	if a.stopChan != nil {
		close(a.stopChan)
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return "Hello " + name
}

func (a *App) CheckPermission() bool {
	return CheckAccessibility()
}
