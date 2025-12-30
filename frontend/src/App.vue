<script setup>
import { ref, onMounted } from 'vue'
import { StartClicking, StopClicking, CheckPermission } from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime/runtime'

import bgImage from './assets/backgrounds/background.jpg'

const isRunning = ref(false)
const appMode = ref('clicker') // 'clicker' or 'presser'
const interval = ref(1000)
const mode = ref('mouse') // 'mouse' or 'keyboard'
const button = ref('left')
const key = ref('Space')
const clickType = ref('single')
const longPressDuration = ref(1000)
const message = ref('按下 Ctrl+F8 (或 Ctrl+Fn+F8) 启动/停止')

const keys = [
  { value: 'Space', label: 'Space (空格)' },
  { value: 'Enter', label: 'Enter (回车)' },
  { value: 'Tab', label: 'Tab (制表)' },
  { value: 'Esc', label: 'Esc (退出)' },
  { value: 'Backspace', label: 'Backspace (退格)' },
  { value: 'Delete', label: 'Delete (删除)' },
  { value: 'Up', label: 'Up (上)' },
  { value: 'Down', label: 'Down (下)' },
  { value: 'Left', label: 'Left (左)' },
  { value: 'Right', label: 'Right (右)' },
  { value: 'F1', label: 'F1' }, { value: 'F2', label: 'F2' }, { value: 'F3', label: 'F3' },
  { value: 'F4', label: 'F4' }, { value: 'F5', label: 'F5' }, { value: 'F6', label: 'F6' },
  { value: 'F7', label: 'F7' }, { value: 'F8', label: 'F8' }, { value: 'F9', label: 'F9' },
  { value: 'F10', label: 'F10' }, { value: 'F11', label: 'F11' }, { value: 'F12', label: 'F12' },
  { value: 'A', label: 'A' }, { value: 'B', label: 'B' }, { value: 'C', label: 'C' },
  { value: 'D', label: 'D' }, { value: 'E', label: 'E' }, { value: 'F', label: 'F' },
  { value: 'G', label: 'G' }, { value: 'H', label: 'H' }, { value: 'I', label: 'I' },
  { value: 'J', label: 'J' }, { value: 'K', label: 'K' }, { value: 'L', label: 'L' },
  { value: 'M', label: 'M' }, { value: 'N', label: 'N' }, { value: 'O', label: 'O' },
  { value: 'P', label: 'P' }, { value: 'Q', label: 'Q' }, { value: 'R', label: 'R' },
  { value: 'S', label: 'S' }, { value: 'T', label: 'T' }, { value: 'U', label: 'U' },
  { value: 'V', label: 'V' }, { value: 'W', label: 'W' }, { value: 'X', label: 'X' },
  { value: 'Y', label: 'Y' }, { value: 'Z', label: 'Z' },
  { value: '0', label: '0' }, { value: '1', label: '1' }, { value: '2', label: '2' },
  { value: '3', label: '3' }, { value: '4', label: '4' }, { value: '5', label: '5' },
  { value: '6', label: '6' }, { value: '7', label: '7' }, { value: '8', label: '8' },
  { value: '9', label: '9' }
]

const switchAppMode = () => {
  if (isRunning.value) return
  if (appMode.value === 'clicker') {
    appMode.value = 'presser'
    clickType.value = 'hold'
  } else {
    appMode.value = 'clicker'
    clickType.value = 'single'
  }
}

const toggleMode = () => {
  if (isRunning.value) return
  mode.value = mode.value === 'mouse' ? 'keyboard' : 'mouse'
}

const toggle = async () => {
  // Check permission before starting
  if (!isRunning.value) {
      try {
          const allowed = await CheckPermission()
          if (!allowed) {
              message.value = "请在系统弹窗中授予辅助功能权限"
              // Alert user to check permissions
              alert("请在系统设置 -> 隐私与安全性 -> 辅助功能中授予本应用权限，然后重试。")
              return
          }
      } catch (e) {
          console.error("Failed to check permission", e)
      }
  }

  if (isRunning.value) {
    await StopClicking()
    isRunning.value = false
    message.value = '已停止'
  } else {
    if (interval.value < 1) interval.value = 1
    if (longPressDuration.value < 10) longPressDuration.value = 10
    
    // Ensure clickType matches appMode
    if (appMode.value === 'presser') {
        clickType.value = 'hold'
    } else {
        // In clicker mode, it could be single or double, but definitely not long if we want strict separation
        // However, if user selected double in UI, keep it.
        // If user somehow has 'long' selected while in clicker mode (e.g. state persistence), force single?
        // But UI hides 'long' option in clicker mode, so it should be fine.
    }

    // Pass mode and key/button to backend
    // Signature: StartClicking(interval int, mode string, keyBtn string, clickType string, longPressDuration int)
    const keyBtn = mode.value === 'mouse' ? button.value : key.value
    await StartClicking(parseInt(interval.value), mode.value, keyBtn, clickType.value, parseInt(longPressDuration.value))
    isRunning.value = true
    message.value = '运行中...'
  }
}

onMounted(async () => {
  try {
    EventsOn("shortcut-pressed", (key) => {
      if (key === "f8") {
        toggle()
      }
    })
  } catch (e) {
    console.error("Wails runtime not ready", e)
  }
})

</script>

<template>
  <div class="background" :style="{ backgroundImage: `url(${bgImage})` }"></div>
  <div class="container">
    <div class="header">
        <button class="switch-btn" @click="switchAppMode" :disabled="isRunning">
            切换到{{ appMode === 'clicker' ? '长按器' : '连点器' }}
        </button>
    </div>
    <h1>{{ appMode === 'clicker' ? '连点器' : '长按器' }}</h1>
    
    <div class="input-group" v-if="appMode === 'clicker'">
      <label>间隔 (毫秒)</label>
      <input v-model="interval" type="number" min="1" :disabled="isRunning" />
    </div>

    <div class="mode-switch">
        <button class="mode-btn" :class="{ active: mode === 'mouse' }" @click="toggleMode" :disabled="isRunning">鼠标{{ appMode === 'clicker' ? '连点' : '长按' }}</button>
        <button class="mode-btn" :class="{ active: mode === 'keyboard' }" @click="toggleMode" :disabled="isRunning">键盘{{ appMode === 'clicker' ? '连点' : '长按' }}</button>
    </div>

    <div v-if="mode === 'mouse'" class="input-group">
      <label>鼠标按键</label>
      <select v-model="button" :disabled="isRunning">
        <option value="left">左键</option>
        <option value="right">右键</option>
        <option value="center">中键</option>
        <option value="side1">侧键 1 (后退)</option>
        <option value="side2">侧键 2 (前进)</option>
      </select>
    </div>

    <div v-else class="input-group">
      <label>键盘按键</label>
      <select v-model="key" :disabled="isRunning">
        <option v-for="k in keys" :key="k.value" :value="k.value">{{ k.label }}</option>
      </select>
    </div>

    <div class="input-group" v-if="appMode === 'clicker'">
      <label>点击模式</label>
      <select v-model="clickType" :disabled="isRunning">
        <option value="single">单击</option>
        <option value="double">双击</option>
        <option value="long">长按</option>
      </select>
    </div>

    <div class="input-group" v-if="appMode === 'clicker' && clickType === 'long'">
      <label>长按时长 (毫秒)</label>
      <input v-model.number="longPressDuration" type="number" min="10" :disabled="isRunning" />
    </div>

    <button class="toggle-btn" :class="{ running: isRunning }" @click="toggle">
      {{ isRunning ? '停止 (Ctrl+F8)' : '启动 (Ctrl+F8)' }}
    </button>

    <p class="status">{{ message }}</p>
  </div>
</template>

<style>
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  margin: 0;
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  user-select: none;
  overflow: hidden;
}

#app {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
}

.background {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-size: cover;
  background-position: center;
  z-index: -1;
  filter: blur(8px); /* Add blur for depth */
  transform: scale(1.1); /* Prevent blur edges */
}

.container {
  background: rgba(255, 255, 255, 0.75);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  padding: 2.5rem;
  width: 320px;
  text-align: center;
  position: relative;
  border-radius: 20px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.4);
}

.header {
  position: fixed;
  top: 16px;
  left: 16px;
  z-index: 1000;
}

.switch-btn {
  background: rgba(33, 150, 243, 0.9);
  color: white;
  border: none;
  padding: 0.5rem 1rem;
  border-radius: 20px;
  cursor: pointer;
  font-size: 0.85rem;
  font-weight: 600;
  transition: all 0.2s ease;
  box-shadow: 0 2px 8px rgba(33, 150, 243, 0.3);
  backdrop-filter: blur(4px);
}

.switch-btn:hover {
  background: #1976D2;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(33, 150, 243, 0.4);
}

.switch-btn:active {
  transform: translateY(0);
}

.switch-btn:disabled {
  background: #ccc;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

h1 {
  margin-top: 0.5rem;
  margin-bottom: 1.5rem;
  color: #1a1a1a;
  font-size: 1.8rem;
  font-weight: 700;
  letter-spacing: -0.5px;
}

.input-group {
  margin-bottom: 1.2rem;
  text-align: left;
}

.input-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 600;
  color: #444;
  font-size: 0.9rem;
  margin-left: 4px;
}

input[type="number"], select {
  width: 100%;
  padding: 0.7rem 0.8rem;
  border: 1px solid rgba(0, 0, 0, 0.1);
  border-radius: 12px;
  font-size: 1rem;
  box-sizing: border-box;
  background: rgba(255, 255, 255, 0.8);
  transition: all 0.2s;
  outline: none;
  color: #333;
}

input[type="number"]:focus, select:focus {
  border-color: #2196F3;
  background: white;
  box-shadow: 0 0 0 3px rgba(33, 150, 243, 0.15);
}

.mode-switch {
  display: flex;
  margin-bottom: 1.5rem;
  background: rgba(0, 0, 0, 0.05);
  border-radius: 12px;
  padding: 4px;
}

.mode-btn {
  flex: 1;
  padding: 0.6rem;
  border: none;
  background: transparent;
  cursor: pointer;
  font-size: 0.95rem;
  font-weight: 600;
  color: #666;
  border-radius: 10px;
  transition: all 0.2s ease;
}

.mode-btn.active {
  background: white;
  color: #2196F3;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.mode-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.toggle-btn {
  width: 100%;
  padding: 1rem;
  border: none;
  border-radius: 14px;
  background: linear-gradient(135deg, #4CAF50, #45a049);
  color: white;
  font-size: 1.1rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s ease;
  margin-top: 1rem;
  box-shadow: 0 4px 12px rgba(76, 175, 80, 0.3);
  text-shadow: 0 1px 2px rgba(0,0,0,0.1);
}

.toggle-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(76, 175, 80, 0.4);
}

.toggle-btn:active {
  transform: translateY(0);
}

.toggle-btn.running {
  background: linear-gradient(135deg, #f44336, #d32f2f);
  box-shadow: 0 4px 12px rgba(244, 67, 54, 0.3);
}

.toggle-btn.running:hover {
  box-shadow: 0 6px 16px rgba(244, 67, 54, 0.4);
}

.status {
  margin-top: 1.2rem;
  color: #666;
  font-size: 0.85rem;
  font-weight: 500;
}
</style>
