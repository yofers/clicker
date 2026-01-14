package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ActionType defines the type of recorded action
type ActionType string

const (
	ActionMouseMove  ActionType = "move"
	ActionMouseClick ActionType = "click" // Down + Up
	ActionMouseDown  ActionType = "down"
	ActionMouseUp    ActionType = "up"
	ActionScroll     ActionType = "scroll"
	ActionKeyDown    ActionType = "keydown"
	ActionKeyUp      ActionType = "keyup"
	ActionWait       ActionType = "wait"
)

// Action represents a single recorded event
type Action struct {
	Type      ActionType `json:"t"`
	X         int        `json:"x,omitempty"`
	Y         int        `json:"y,omitempty"`
	Button    string     `json:"b,omitempty"` // left, right, middle
	Key       string     `json:"k,omitempty"` // Key name or code
	ScrollX   int        `json:"sx,omitempty"`
	ScrollY   int        `json:"sy,omitempty"`
	Timestamp int64      `json:"ts"` // Relative timestamp in ms
}

// Recorder handles recording and playback
type Recorder struct {
	mu              sync.Mutex
	actions         []Action
	startTime       time.Time
	isRecording     bool
	isPlaying       bool
	stopPlayChan    chan struct{}
	ctx             context.Context // Add context for event emitting
	currentFilename string
}

var globalRecorder = &Recorder{}

// Track currently pressed keys/buttons for UI feedback
var activeKeys = make(map[string]time.Time)
var activeKeysMu sync.Mutex

func updateActiveState(ctx context.Context, action Action) string {
	activeKeysMu.Lock()
	defer activeKeysMu.Unlock()

	now := time.Now()

	// Key/Button name
	name := action.Key
	if action.Button != "" {
		switch action.Button {
		case "left":
			name = "鼠标左键"
		case "right":
			name = "鼠标右键"
		case "center":
			name = "鼠标中键"
		default:
			name = "鼠标" + action.Button
		}
	}

	if name == "" {
		return getActiveString(now) // Just return current state for Move/Scroll
	}

	switch action.Type {
	case ActionMouseDown, ActionKeyDown:
		activeKeys[name] = now
	case ActionMouseUp, ActionKeyUp:
		delete(activeKeys, name)
	}

	return getActiveString(now)
}

func getActiveString(now time.Time) string {
	// Build string "w(200ms) + 鼠标左键(5s)"
	var parts []string
	for k, start := range activeKeys {
		duration := now.Sub(start).Milliseconds()
		// Only show duration if > 0
		if duration > 0 {
			parts = append(parts, fmt.Sprintf("%s(%dms)", k, duration))
		} else {
			parts = append(parts, k)
		}
	}
	// Sort for stability
	sort.Strings(parts)
	if len(parts) == 0 {
		return ""
	}
	return "动作：" + strings.Join(parts, " + ")
}

// Crypto Key (In production, this should be managed better, but for this app hardcoded is acceptable as per requirements)
var cryptoKey = []byte("AutoClicker_YFC_Secure_Key_2026") // 32 bytes needed for AES-256

func init() {
	// Ensure key is 32 bytes
	if len(cryptoKey) < 32 {
		padding := make([]byte, 32-len(cryptoKey))
		cryptoKey = append(cryptoKey, padding...)
	} else if len(cryptoKey) > 32 {
		cryptoKey = cryptoKey[:32]
	}
}

// Encryption Helpers
func encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(cryptoKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, data, nil), nil
}

func decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(cryptoKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// App Methods for Recorder

func (a *App) StartRecording(filename string) error {
	fmt.Printf("DEBUG: StartRecording called with filename: '%s'\n", filename)
	globalRecorder.mu.Lock()
	defer globalRecorder.mu.Unlock()

	// Ensure filename ends with .yfc if provided and not empty
	if filename != "" && !strings.HasSuffix(filename, ".yfc") {
		filename = filename + ".yfc"
	}

	globalRecorder.currentFilename = filename
	globalRecorder.actions = make([]Action, 0)
	globalRecorder.startTime = time.Now()
	globalRecorder.isRecording = true
	globalRecorder.ctx = a.ctx // Store context

	// Reset active keys
	activeKeysMu.Lock()
	activeKeys = make(map[string]time.Time)
	activeKeysMu.Unlock()

	// Start feedback loop for recording
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			globalRecorder.mu.Lock()
			recording := globalRecorder.isRecording
			ctx := globalRecorder.ctx
			globalRecorder.mu.Unlock()

			if !recording {
				return
			}

			select {
			case <-ticker.C:
				activeKeysMu.Lock()
				if len(activeKeys) > 0 {
					msg := getActiveString(time.Now())
					if ctx != nil {
						runtime.EventsEmit(ctx, "recording-feedback", msg)
					}
				} else {
					// Also emit empty string to clear if needed, but usually RecordEvent handles changes.
					// But if we want to clear trailing durations, better emit.
					// However, RecordEvent emits when keys are released (msg becomes empty or shorter).
					// The ticker is mainly for updating durations of HELD keys.
				}
				activeKeysMu.Unlock()
			}
		}
	}()
	return nil
}

func (a *App) StopRecording() (string, error) {
	globalRecorder.mu.Lock()
	defer globalRecorder.mu.Unlock()

	if !globalRecorder.isRecording {
		return "", nil
	}
	globalRecorder.isRecording = false

	// Save to file
	data, err := json.Marshal(globalRecorder.actions)
	if err != nil {
		return "", err
	}

	encrypted, err := encrypt(data)
	if err != nil {
		return "", err
	}

	filename := globalRecorder.currentFilename
	if filename == "" {
		filename = fmt.Sprintf("recording_%s.yfc", time.Now().Format("2006_01_02_15_04_05"))
	}

	// Get User Config Directory or Home Directory to store recordings
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to home dir
		configDir, err = os.UserHomeDir()
		if err != nil {
			return "", err
		}
	}

	saveDir := filepath.Join(configDir, "AutoClicker", "recordings")
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}

	fullPath := filepath.Join(saveDir, filename)

	err = os.WriteFile(fullPath, encrypted, 0644)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func (a *App) GetRecordings() ([]string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir, _ = os.UserHomeDir()
	}

	saveDir := filepath.Join(configDir, "AutoClicker", "recordings")
	// Ensure it exists to avoid error on ReadDir
	os.MkdirAll(saveDir, 0755)

	files, err := os.ReadDir(saveDir)
	if err != nil {
		return nil, err
	}

	var yfcFiles []string
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".yfc" {
			yfcFiles = append(yfcFiles, f.Name())
		}
	}
	// Sort by newest first
	sort.Slice(yfcFiles, func(i, j int) bool {
		return yfcFiles[i] > yfcFiles[j] // Simple timestamp sort since name starts with recording_TIMESTAMP
	})
	return yfcFiles, nil
}

func (a *App) PlayRecording(filename string, loopInterval int, loopCount int) error {
	globalRecorder.mu.Lock()
	if globalRecorder.isPlaying {
		globalRecorder.mu.Unlock()
		return fmt.Errorf("already playing")
	}
	globalRecorder.stopPlayChan = make(chan struct{})
	globalRecorder.isPlaying = true
	globalRecorder.mu.Unlock()

	// Load file logic: Support absolute path or relative to default dir
	var filePath string
	if filepath.IsAbs(filename) {
		filePath = filename
	} else {
		configDir, err := os.UserConfigDir()
		if err != nil {
			configDir, _ = os.UserHomeDir()
		}
		saveDir := filepath.Join(configDir, "AutoClicker", "recordings")
		filePath = filepath.Join(saveDir, filename)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		a.StopPlayback()
		return err
	}

	decrypted, err := decrypt(content)
	if err != nil {
		a.StopPlayback()
		return fmt.Errorf("decrypt failed: %v", err)
	}

	var actions []Action
	err = json.Unmarshal(decrypted, &actions)
	if err != nil {
		a.StopPlayback()
		return err
	}

	if len(actions) == 0 {
		a.StopPlayback()
		return fmt.Errorf("recording is empty")
	}

	stopChan := globalRecorder.stopPlayChan // Capture locally

	go func() {
		defer a.StopPlayback()

		// Reset state
		activeKeysMu.Lock()
		activeKeys = make(map[string]time.Time)
		activeKeysMu.Unlock()

		currentLoop := 0
		for {
			// Check stop signal immediately at start of loop
			select {
			case <-stopChan:
				return
			default:
			}

			// Check loop count (0 = infinite)
			if loopCount > 0 && currentLoop >= loopCount {
				return
			}
			currentLoop++

			startTime := time.Now()

			for i, action := range actions {
				// Check stop signal before every action
				select {
				case <-stopChan:
					return
				default:
				}

				// Emit feedback event
				feedback := updateActiveState(a.ctx, action)
				if feedback != "" {
					runtime.EventsEmit(a.ctx, "playback-feedback", feedback)
				}

				// Calculate delay
				targetTime := startTime.Add(time.Duration(action.Timestamp) * time.Millisecond)

				// Wait loop with feedback updates
				for {
					now := time.Now()
					if now.After(targetTime) {
						break
					}

					select {
					case <-stopChan:
						return
					default:
					}

					// Update feedback for held keys
					activeKeysMu.Lock()
					if len(activeKeys) > 0 {
						msg := getActiveString(now)
						runtime.EventsEmit(a.ctx, "playback-feedback", msg)
					}
					activeKeysMu.Unlock()

					// Sleep small amount
					sleepDur := targetTime.Sub(now)
					if sleepDur > 10*time.Millisecond {
						sleepDur = 10 * time.Millisecond
					}
					time.Sleep(sleepDur)
				}

				// Execute Action
				executeAction(action)

				// Check stop signal immediately after execution
				select {
				case <-stopChan:
					return
				default:
				}

				// If last action or large gap, emit update
				if i < len(actions)-1 {
					// Check if next action is far away, maybe update UI?
					// For now, just update on every event
				}
			}

			// Clear state at end of loop
			activeKeysMu.Lock()
			activeKeys = make(map[string]time.Time)
			activeKeysMu.Unlock()
			runtime.EventsEmit(a.ctx, "playback-feedback", "")

			// Loop delay
			if loopInterval > 0 {
				// Sleep with check
				loopStart := time.Now()
				loopTarget := loopStart.Add(time.Duration(loopInterval) * time.Millisecond)
				for {
					if time.Now().After(loopTarget) {
						break
					}
					select {
					case <-stopChan:
						return
					default:
						time.Sleep(10 * time.Millisecond)
					}
				}
			} else {
				// If no loop interval, maybe just return if no infinite loop?
				// But loopCount=0 means infinite.
				// If loopInterval is 0, we just loop immediately.
			}
		}
	}()

	return nil
}

func (a *App) SelectRecordingFile() (string, error) {
	configDir, _ := os.UserConfigDir()
	defaultDir := filepath.Join(configDir, "AutoClicker", "recordings")
	// Ensure default directory exists
	os.MkdirAll(defaultDir, 0755)

	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "选择动作文件",
		DefaultDirectory: defaultDir,
		Filters: []runtime.FileFilter{
			{
				DisplayName: "YFC Recording Files (*.yfc)",
				Pattern:     "*.yfc",
			},
		},
	})

	if err != nil {
		return "", err
	}

	if file == "" {
		return "", nil
	}

	// If file is selected, copy it to default directory if it's not already there
	// This ensures it appears in the dropdown list
	// Check if file is already in defaultDir
	absFile, _ := filepath.Abs(file)
	absDefaultDir, _ := filepath.Abs(defaultDir)

	if filepath.Dir(absFile) != absDefaultDir {
		// Copy file
		destFile := filepath.Join(defaultDir, filepath.Base(file))

		input, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("failed to read source file: %v", err)
		}

		err = os.WriteFile(destFile, input, 0644)
		if err != nil {
			return "", fmt.Errorf("failed to copy file to default dir: %v", err)
		}

		// Return the filename only, so frontend can select it in dropdown
		return filepath.Base(file), nil
	}

	// If already in default dir, just return the filename
	return filepath.Base(file), nil
}

func (a *App) DeleteRecording(filename string) error {
	// Validate filename to prevent path traversal
	if filename == "" || strings.Contains(filename, "..") || strings.Contains(filename, string(filepath.Separator)) {
		return fmt.Errorf("invalid filename")
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir, _ = os.UserHomeDir()
	}

	saveDir := filepath.Join(configDir, "AutoClicker", "recordings")
	fullPath := filepath.Join(saveDir, filename)

	// Ensure file exists and is a file
	info, err := os.Stat(fullPath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("path is a directory")
	}

	return os.Remove(fullPath)
}

func (a *App) ConfirmDelete(filename string) (bool, error) {
	result, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         "确认删除",
		Message:       fmt.Sprintf("确定要删除文件 \"%s\" 吗？", filename),
		Buttons:       []string{"取消", "删除"},
		DefaultButton: "删除",
		CancelButton:  "取消",
	})
	if err != nil {
		return false, err
	}
	return result == "删除", nil
}

func (a *App) StopPlayback() {
	globalRecorder.mu.Lock()
	defer globalRecorder.mu.Unlock()

	if globalRecorder.isPlaying {
		if globalRecorder.stopPlayChan != nil {
			close(globalRecorder.stopPlayChan)
			globalRecorder.stopPlayChan = nil
		}
		globalRecorder.isPlaying = false
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "playback-stopped", true)
		}
	}
}

// RecordEvent is called by hooks
func RecordEvent(action Action) {
	globalRecorder.mu.Lock()
	defer globalRecorder.mu.Unlock()

	if globalRecorder.isRecording {
		// Set relative timestamp
		action.Timestamp = time.Since(globalRecorder.startTime).Milliseconds()
		globalRecorder.actions = append(globalRecorder.actions, action)
	}
}

// Platform specific execution need to be implemented or bridged
func executeAction(a Action) {
	switch a.Type {
	case ActionMouseMove:
		moveMouse(a.X, a.Y)
	case ActionMouseDown:
		mouseToggle(a.Button, true)
	case ActionMouseUp:
		mouseToggle(a.Button, false)
	case ActionKeyDown:
		keyToggle(a.Key, true)
	case ActionKeyUp:
		keyToggle(a.Key, false)
	}
}
