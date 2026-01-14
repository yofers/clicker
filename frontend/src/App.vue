<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { StartClicking, StopClicking, CheckPermission, StartRecording, StopRecording, GetRecordings, PlayRecording, StopPlayback, SelectRecordingFile, DeleteRecording, ConfirmDelete } from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime/runtime'

import bgImage from './assets/backgrounds/background.jpg'

const isRunning = ref(false) // Deprecated global state
const mouseRunMode = ref(null) // 'clicker', 'presser' or null
const keyboardRunMode = ref(null) // 'clicker', 'presser' or null
const activeTasks = ref([]) // Array of task info strings

const appMode = ref('clicker') // 'clicker' or 'presser'
const interval = ref(1000)
const mode = ref('mouse') // 'mouse' or 'keyboard'
const buttons = ref(['left'])
const selectedKeys = ref(['Space'])
const clickType = ref('single')
const longPressDuration = ref(1000)
const message = ref('')

// Recorder State
const recordings = ref([])
const isRecording = ref(false)
const isPlaying = ref(false)
const playLoopInterval = ref(1000)
const playLoopCount = ref(0) // 0 = infinite
const currentPlayingFile = ref('')
const recorderMode = ref('record') // 'record' or 'execute'
const selectedFile = ref('')
const recordingFilename = ref('')
const recordingFeedback = ref('')
const playbackFeedback = ref('')

const formatFileName = (name) => {
    return name ? name.replace(/\.yfc$/, '') : ''
}

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

const isDropdownOpen = ref(false)

// Computed states for UI locking
const isMouseBusy = computed(() => !!mouseRunMode.value)
const isKeyboardBusy = computed(() => !!keyboardRunMode.value)

const shouldDisableMouseBtn = computed(() => {
    // Disable Mouse Mode button if Mouse is running in a DIFFERENT appMode
    return mouseRunMode.value && mouseRunMode.value !== appMode.value
})

const shouldDisableKeyboardBtn = computed(() => {
    // Disable Keyboard Mode button if Keyboard is running in a DIFFERENT appMode
    return keyboardRunMode.value && keyboardRunMode.value !== appMode.value
})

const shouldDisableInputs = computed(() => {
    // Disable inputs if the current device is busy anywhere
    // If I am on Mouse, and Mouse is busy (in any mode), lock inputs
    if (mode.value === 'mouse') return isMouseBusy.value
    if (mode.value === 'keyboard') return isKeyboardBusy.value
    return false
})

const addKey = () => selectedKeys.value.push('Space')
const removeKey = (index) => selectedKeys.value.splice(index, 1)
const addButton = () => buttons.value.push('left')
const removeButton = (index) => buttons.value.splice(index, 1)

// Helper to update active tasks list
const updateActiveTasks = () => {
    const tasks = []
    if (mouseRunMode.value) tasks.push(`鼠标${mouseRunMode.value === 'clicker' ? '连点' : '长按'}运行中`)
    if (keyboardRunMode.value) tasks.push(`键盘${keyboardRunMode.value === 'clicker' ? '连点' : '长按'}运行中`)
    if (isPlaying.value) tasks.push(`动作回放运行中`)
    if (isRecording.value) tasks.push(`正在录制动作`)
    activeTasks.value = tasks
}

const switchAppMode = () => {
  if (isRecording.value) {
      message.value = "录制中无法切换模式"
      return
  }
  // Allow switching even if running, just update UI defaults
  if (appMode.value === 'clicker') {
    appMode.value = 'presser'
    clickType.value = 'hold'
  } else if (appMode.value === 'presser') {
    appMode.value = 'recorder'
    loadRecordings()
  } else {
    appMode.value = 'clicker'
    clickType.value = 'single'
  }

  // Auto-switch device mode if current device is busy in the other appMode
  // Only relevant for Clicker/Presser modes
  if (appMode.value !== 'recorder') {
      if (mode.value === 'mouse' && shouldDisableMouseBtn.value) {
          mode.value = 'keyboard'
      } else if (mode.value === 'keyboard' && shouldDisableKeyboardBtn.value) {
          mode.value = 'mouse'
      }
  }
}

const toggleMode = () => {
  // Allow switching even if running
  mode.value = mode.value === 'mouse' ? 'keyboard' : 'mouse'
}

const isCurrentModeRunning = () => {
    return mode.value === 'mouse' ? !!mouseRunMode.value : !!keyboardRunMode.value
}

const toggle = async () => {
  const currentRunning = isCurrentModeRunning()

  // Check permission before starting new task
  if (!currentRunning) {
      // Check if we can start (is button disabled?)
      // Actually toggle is called by button click or shortcut.
      // If via shortcut, we must check if we are allowed to start.
      if (mode.value === 'mouse' && shouldDisableMouseBtn.value) return
      if (mode.value === 'keyboard' && shouldDisableKeyboardBtn.value) return
      
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

  if (currentRunning) {
    await StopClicking(mode.value)
    if (mode.value === 'mouse') mouseRunMode.value = null
    else keyboardRunMode.value = null
    message.value = `${mode.value === 'mouse' ? '鼠标' : '键盘'}任务已停止`
  } else {
    if (interval.value < 1) interval.value = 1
    if (longPressDuration.value < 10) longPressDuration.value = 10
    
    // Ensure clickType matches appMode
    if (appMode.value === 'presser') {
        clickType.value = 'hold'
    }

    // Pass mode and key/button to backend
    // Signature: StartClicking(interval int, mode string, keysOrButtons []string, clickType string, longPressDuration int)
    let keysToPass = []
    if (mode.value === 'mouse') {
        keysToPass = buttons.value
    } else {
        keysToPass = selectedKeys.value
    }
    
    await StartClicking(parseInt(interval.value), mode.value, keysToPass, clickType.value, parseInt(longPressDuration.value))
    
    if (mode.value === 'mouse') mouseRunMode.value = appMode.value
    else keyboardRunMode.value = appMode.value
    
    message.value = `${mode.value === 'mouse' ? '鼠标' : '键盘'}任务运行中...`
  }
  updateActiveTasks()
}

// Recorder Functions
const loadRecordings = async () => {
    try {
        const list = await GetRecordings()
        recordings.value = list || []
        // Select newest by default if not selected
        if (recordings.value.length > 0) {
             selectedFile.value = recordings.value[0]
        }
    } catch (e) {
        console.error("Failed to load recordings", e)
    }
}

const toggleRecording = async () => {
    if (isRecording.value) {
        try {
            const filename = await StopRecording()
            isRecording.value = false
            message.value = `录制完成: ${filename}`
            // Auto switch to execute mode
            await loadRecordings()
            recorderMode.value = 'execute'
            selectedFile.value = filename
            updateActiveTasks()
        } catch (e) {
            message.value = `停止录制失败: ${e}`
        }
    } else {
        // Check permissions first
        try {
            const allowed = await CheckPermission()
            if (!allowed) {
                 alert("请在系统设置 -> 隐私与安全性 -> 辅助功能中授予本应用权限，然后重试。")
                 return
            }
            console.log("DEBUG: Calling StartRecording with:", recordingFilename.value)
            await StartRecording(recordingFilename.value)
            isRecording.value = true
            message.value = "正在录制..."
            updateActiveTasks()
        } catch (e) {
             message.value = `启动录制失败: ${e}`
        }
    }
}

const playRec = async () => {
    if (isPlaying.value) {
        // Stop playback
        await stopPlay()
        return
    }

    if (!selectedFile.value) {
        message.value = "请先选择录制文件"
        return
    }

    try {
        currentPlayingFile.value = selectedFile.value
        isPlaying.value = true
        message.value = `正在回放: ${selectedFile.value}`
        updateActiveTasks()
        await PlayRecording(selectedFile.value, parseInt(playLoopInterval.value), parseInt(playLoopCount.value))
    } catch (e) {
        isPlaying.value = false
        currentPlayingFile.value = ''
        message.value = `回放失败: ${e}`
        updateActiveTasks()
    }
}

const stopPlay = async () => {
    try {
        await StopPlayback()
        isPlaying.value = false
        currentPlayingFile.value = ''
        message.value = "回放已停止"
        playbackFeedback.value = '' // Clear feedback immediately
        updateActiveTasks()
    } catch (e) {
        message.value = `停止回放失败: ${e}`
    }
}

const selectFile = async () => {
    if (isPlaying.value) return
    try {
        const file = await SelectRecordingFile()
        if (file) {
            // Reload list to include the newly imported file
            await loadRecordings()
            selectedFile.value = file
        }
    } catch (e) {
        console.error("Failed to select file", e)
    }
}

const setRecorderMode = (mode) => {
    if (isRecording.value) {
        message.value = "录制中无法切换页面"
        return
    }
    recorderMode.value = mode
    if (mode === 'execute') {
        loadRecordings()
    }
}

const stopAll = async () => {
    if (mouseRunMode.value) {
        await StopClicking('mouse')
        mouseRunMode.value = null
    }
    if (keyboardRunMode.value) {
        await StopClicking('keyboard')
        keyboardRunMode.value = null
    }
    
    // Stop Playback if running
    if (isPlaying.value) {
        await stopPlay()
    }
    // Stop Recording if running
    if (isRecording.value) {
        await toggleRecording()
    }

    updateActiveTasks()
    message.value = "所有任务已停止"
}

const toggleDropdown = () => {
  if (isPlaying.value) return
  isDropdownOpen.value = !isDropdownOpen.value
}

const selectRecording = (file) => {
  selectedFile.value = file
  isDropdownOpen.value = false
}

const removeRecording = async (file) => {
  message.value = `正在处理删除请求: ${file}`
  try {
    const confirmed = await ConfirmDelete(file)
    if (!confirmed) {
        message.value = "已取消删除"
        return
    }

    await DeleteRecording(file)
    await loadRecordings()
    // If we deleted the currently selected file, clear selection or select first
    if (selectedFile.value === file) {
      selectedFile.value = recordings.value.length > 0 ? recordings.value[0] : ''
    }
    message.value = `已删除文件: ${file}`
  } catch (e) {
    message.value = `删除失败: ${e}`
  }
}

const closeDropdown = (e) => {
    if (!e.target.closest('.custom-select-container')) {
        isDropdownOpen.value = false
    }
}

onMounted(async () => {
    document.addEventListener('click', closeDropdown)
  try {
    EventsOn("shortcut-pressed", (key) => {
      if (key === "f6") {
          // Switch App Mode (Clicker <-> Presser)
          switchAppMode()
      } else if (key === "f7") {
          // Ctrl+F7: Toggle Mouse Action (Clicker/Presser) or Recording (Recorder)
          if (appMode.value === 'recorder') {
              if (recorderMode.value !== 'record') {
                  setRecorderMode('record')
              }
              toggleRecording()
          } else {
              // Clicker or Presser mode - Toggle Mouse
              if (mode.value !== 'mouse') {
                  // Only switch mode if we want to visualize it, 
                  // but requirements say "Ctrl+F7 starts mouse clicker", implies specific device toggle
                  // Let's switch UI to mouse for feedback, then toggle
                  mode.value = 'mouse'
              }
              // If already running keyboard, we might need to stop it or allow parallel?
              // Current logic: toggle() starts whatever is in mode.value
              // But we want specific F7 -> Mouse, F8 -> Keyboard
              
              // We need to ensure mode is set to mouse, then call toggle if not running, or stop if running mouse
              if (mouseRunMode.value) {
                  // Mouse is running, stop it
                  // We need to set mode to mouse to call toggle() correctly or call StopClicking directly
                  mode.value = 'mouse' // Sync UI
                  toggle() 
              } else {
                  // Mouse not running. 
                  // If Keyboard is running, can we run both? 
                  // The backend supports separate StartClicking calls for mouse/keyboard if we implemented it right.
                  // StartClicking takes "mode" arg.
                  // Let's assume we can run parallel.
                  mode.value = 'mouse'
                  toggle()
              }
          }
      } else if (key === "f8") {
          // Ctrl+F8: Toggle Keyboard Action (Clicker/Presser) or Playback (Recorder)
          if (appMode.value === 'recorder') {
              if (isRecording.value) {
                  message.value = "录制中无法执行动作"
                  return
              }
              if (recorderMode.value !== 'execute') {
                  setRecorderMode('execute')
              }
              playRec()
          } else {
              // Clicker or Presser mode - Toggle Keyboard
              if (mode.value !== 'keyboard') {
                  mode.value = 'keyboard'
              }
              
              if (keyboardRunMode.value) {
                  mode.value = 'keyboard'
                  toggle()
              } else {
                  mode.value = 'keyboard'
                  toggle()
              }
          }
      } else if (key === "f9") {
          // Ctrl+F9: Stop All
          stopAll()
      }
    })
    
    EventsOn("recording-feedback", (msg) => {
        recordingFeedback.value = msg
    })

    EventsOn("playback-feedback", (msg) => {
        playbackFeedback.value = msg
    })
  } catch (e) {
    console.error("Wails runtime not ready", e)
  }
})

onUnmounted(() => {
    document.removeEventListener('click', closeDropdown)
})

</script>

<template>
  <div class="background" :style="{ backgroundImage: `url(${bgImage})` }"></div>
  <div class="container">
    <div class="header">
        <button class="switch-btn" @click="switchAppMode" title="快捷键: Ctrl+F6" :disabled="isRecording">
            切换到{{ appMode === 'clicker' ? '长按器' : (appMode === 'presser' ? '录制器' : '连点器') }}
        </button>
    </div>
    <div class="status-panel" v-if="activeTasks.length > 0">
        <div v-for="task in activeTasks" :key="task" class="status-item">{{ task }}</div>
    </div>
    <h1>{{ appMode === 'clicker' ? '连点器' : (appMode === 'presser' ? '长按器' : '录制器') }}</h1>
    
    <!-- Recorder UI -->
    <div v-if="appMode === 'recorder'" class="recorder-container">
        
        <!-- Mode Switcher -->
        <div class="mode-switch">
            <button class="mode-btn" :class="{ active: recorderMode === 'record' }" @click="setRecorderMode('record')" :disabled="isRecording || isPlaying">动作录制</button>
            <button class="mode-btn" :class="{ active: recorderMode === 'execute' }" @click="setRecorderMode('execute')" :disabled="isRecording || isPlaying">执行动作</button>
        </div>

        <!-- Record Mode Content -->
        <div v-if="recorderMode === 'record'" class="input-group">
            <div class="input-group" style="margin-bottom: 0.8rem;">
                <label>保存文件名 (可选)</label>
                <input v-model="recordingFilename" type="text" placeholder="留空则自动生成 (recording_日期.yfc)" :disabled="isRecording" style="width: 100%; padding: 0.7rem 0.8rem; border: 1px solid rgba(0, 0, 0, 0.1); border-radius: 12px; font-size: 0.9rem; box-sizing: border-box; background-color: rgba(255, 255, 255, 0.8);" />
            </div>
            <button class="toggle-btn" :class="{ running: isRecording }" @click="toggleRecording" :disabled="isPlaying">
                {{ isRecording ? '结束录制' : '开始录制' }}
            </button>
            <div style="margin-top: 10px; font-size: 0.85rem; color: #666; text-align: center;">
                点击开始录制后，将记录所有的鼠标和键盘操作。
            </div>
            <div v-if="isRecording" class="feedback-panel" style="margin-top: 15px; text-align: center; color: #2196F3; font-weight: bold; min-height: 20px;">
                {{ recordingFeedback }}
            </div>
        </div>

        <!-- Execute Mode Content -->
        <div v-if="recorderMode === 'execute'" class="execute-panel">
             <div class="input-group">
                <label>选择动作文件</label>
                <div class="joined-input-row">
                    <div class="custom-select-container" :class="{ disabled: isPlaying }">
                        <div class="custom-select-trigger" @click="toggleDropdown">
                            <span>{{ selectedFile ? formatFileName(selectedFile) : (recordings.length === 0 ? '无录制文件' : '请选择文件') }}</span>
                            <span class="arrow">▼</span>
                        </div>
                        <div class="custom-options" v-if="isDropdownOpen">
                            <div v-for="file in recordings" :key="file" class="custom-option" :class="{ selected: file === selectedFile }" @click="selectRecording(file)">
                                <span class="option-text">{{ formatFileName(file) }}</span>
                                <span class="delete-icon" @click.stop="removeRecording(file)" title="删除文件">✕</span>
                            </div>
                            <div v-if="recordings.length === 0" class="custom-option disabled">无录制文件</div>
                        </div>
                    </div>
                    <button @click="selectFile" class="browse-btn" :disabled="isPlaying" title="导入文件">
                        导入文件
                    </button>
                </div>
             </div>

             <!-- Playback Settings -->
             <div style="margin-bottom: 15px;">
                 <div class="input-group">
                      <label>回放间隔(毫秒)</label>
                      <input v-model.number="playLoopInterval" type="number" min="0" :disabled="isPlaying || isRecording" />
                 </div>
                 <div class="input-group">
                      <label>循环次数 (0 = 无限)</label>
                      <input v-model.number="playLoopCount" type="number" min="0" :disabled="isPlaying || isRecording" />
                 </div>
             </div>

             <button class="toggle-btn" :class="{ running: isPlaying }" @click="playRec" :disabled="!selectedFile">
                {{ isPlaying ? '停止回放' : '启动(Ctrl+F8)' }}
             </button>
             
             <div v-if="isPlaying" style="margin-top: 10px; text-align: center; color: #2196F3; font-weight: bold;">
                 <div>正在运行: {{ currentPlayingFile }}</div>
                 <div style="margin-top: 5px;">{{ playbackFeedback }}</div>
             </div>
        </div>
    </div>

    <!-- Clicker/Presser UI -->
    <template v-else>
    <div class="input-group" v-if="appMode === 'clicker'">
      <label>间隔 (毫秒)</label>
      <input v-model="interval" type="number" min="1" :disabled="shouldDisableInputs" />
    </div>

    <div class="mode-switch">
        <button class="mode-btn" :class="{ active: mode === 'mouse' }" @click="toggleMode" :disabled="shouldDisableMouseBtn">鼠标{{ appMode === 'clicker' ? '连点' : '长按' }}</button>
        <button class="mode-btn" :class="{ active: mode === 'keyboard' }" @click="toggleMode" :disabled="shouldDisableKeyboardBtn">键盘{{ appMode === 'clicker' ? '连点' : '长按' }}</button>
    </div>

    <div v-if="mode === 'mouse'" class="input-group">
      <div class="label-row">
          <label>鼠标按键</label>
          <button @click="addButton" class="add-btn" title="添加按键" :disabled="shouldDisableInputs">+</button>
      </div>
      
      <div>
          <div v-for="(btn, index) in buttons" :key="index" class="select-row">
            <select v-model="buttons[index]" :disabled="shouldDisableInputs">
                <option value="left">左键</option>
                <option value="right">右键</option>
                <option value="center">中键</option>
                <option value="side1">侧键 1 (后退)</option>
                <option value="side2">侧键 2 (前进)</option>
            </select>
            <button v-if="buttons.length > 1" @click="removeButton(index)" class="remove-btn" :disabled="shouldDisableInputs" title="移除按键">-</button>
          </div>
      </div>
    </div>

    <div v-else class="input-group">
      <div class="label-row">
          <label>键盘按键</label>
          <button @click="addKey" class="add-btn" title="添加按键" :disabled="shouldDisableInputs">+</button>
      </div>
      
      <div>
          <div v-for="(k, index) in selectedKeys" :key="index" class="select-row">
            <select v-model="selectedKeys[index]" :disabled="shouldDisableInputs">
                <option v-for="opt in keys" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
            </select>
            <button v-if="selectedKeys.length > 1" @click="removeKey(index)" class="remove-btn" :disabled="shouldDisableInputs" title="移除按键">-</button>
          </div>
      </div>
    </div>

    <div class="input-group" v-if="appMode === 'clicker'">
      <label>点击模式</label>
      <select v-model="clickType" :disabled="shouldDisableInputs">
        <option value="single">单击</option>
        <option value="double">双击</option>
        <option value="long">长按</option>
      </select>
    </div>

    <div class="input-group" v-if="appMode === 'clicker' && clickType === 'long'">
      <label>长按时长 (毫秒)</label>
      <input v-model.number="longPressDuration" type="number" min="10" :disabled="shouldDisableInputs" />
    </div>

    <button class="toggle-btn" :class="{ running: isCurrentModeRunning() }" @click="toggle" :disabled="mode === 'mouse' ? shouldDisableMouseBtn : shouldDisableKeyboardBtn" :title="mode === 'mouse' ? '快捷键: Ctrl+F7' : '快捷键: Ctrl+F8'">
      {{ isCurrentModeRunning() ? '停止' : '启动' }}
    </button>
    </template>

    <p class="status">{{ message }}</p>

    <div class="help-container">
      <div class="help-btn">?</div>
      <div class="tooltip">
        <div class="tooltip-item"><strong>Ctrl+F6</strong>: 切换模式</div>
        <div class="tooltip-item" v-if="appMode === 'recorder'"><strong>Ctrl+F7</strong>: 录制动作开关</div>
        <div class="tooltip-item" v-else><strong>Ctrl+F7</strong>: 鼠标任务开关</div>
        <div class="tooltip-item" v-if="appMode === 'recorder'"><strong>Ctrl+F8</strong>: 执行动作开关</div>
        <div class="tooltip-item" v-else><strong>Ctrl+F8</strong>: 键盘任务开关</div>
        <div class="tooltip-item"><strong>Ctrl+F9</strong>: 停止所有任务</div>
      </div>
    </div>
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
  background-color: rgba(255, 255, 255, 0.8);
  transition: all 0.2s;
  outline: none;
  color: #333;
}

select {
  appearance: none;
  -webkit-appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%23666' d='M6 8.825L1.175 4 2.238 2.938 6 6.7l3.763-3.762L10.825 4z'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 1rem center;
  background-size: 12px;
  cursor: pointer;
  padding-right: 2.5rem;
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

.label-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.label-row label {
  margin-bottom: 0;
}

.add-btn {
  background: #4CAF50;
  color: white;
  border: none;
  border-radius: 50%;
  width: 24px;
  height: 24px;
  cursor: pointer;
  font-size: 18px;
  line-height: 1;
  display: flex;
  justify-content: center;
  align-items: center;
  box-shadow: 0 2px 4px rgba(0,0,0,0.2);
  transition: all 0.2s;
}

.add-btn:hover {
  transform: scale(1.1);
  background: #45a049;
}

.add-btn:disabled {
  background: #ccc;
  cursor: not-allowed;
  transform: none;
}

.select-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: center;
}

.select-row select {
  flex: 1;
}

.remove-btn {
  background: #f44336;
  color: white;
  border: none;
  border-radius: 50%;
  width: 24px;
  height: 24px;
  cursor: pointer;
  font-size: 18px;
  line-height: 1;
  display: flex;
  justify-content: center;
  align-items: center;
  box-shadow: 0 2px 4px rgba(0,0,0,0.2);
  transition: all 0.2s;
  flex-shrink: 0;
}

.remove-btn:hover {
  transform: scale(1.1);
  background: #d32f2f;
}

.remove-btn:disabled {
  background: #ccc;
  cursor: not-allowed;
  transform: none;
}
.status-panel {
  position: absolute;
  top: 12px;
  right: 12px;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
  z-index: 10;
}

.status-item {
  background: rgba(33, 150, 243, 0.9);
  color: white;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 600;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  backdrop-filter: blur(4px);
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(5px); }
  to { opacity: 1; transform: translateY(0); }
}

.help-container {
  position: fixed;
  bottom: 16px;
  left: 16px;
  z-index: 1000;
}

.help-btn {
  width: 24px;
  height: 24px;
  background: rgba(0, 0, 0, 0.4);
  color: white;
  border-radius: 50%;
  display: flex;
  justify-content: center;
  align-items: center;
  font-weight: bold;
  font-size: 14px;
  cursor: help;
  transition: all 0.2s;
  backdrop-filter: blur(4px);
}

.help-btn:hover {
  background: rgba(33, 150, 243, 0.9);
  transform: scale(1.1);
}

.tooltip {
  position: absolute;
  bottom: 32px;
  left: 0;
  background: rgba(0, 0, 0, 0.85);
  color: white;
  padding: 8px 12px;
  border-radius: 8px;
  width: max-content;
  font-size: 12px;
  line-height: 1.6;
  opacity: 0;
  visibility: hidden;
  transform: translateY(10px);
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 4px 12px rgba(0,0,0,0.2);
  backdrop-filter: blur(8px);
  pointer-events: none;
}

.help-container:hover .tooltip {
  opacity: 1;
  visibility: visible;
  transform: translateY(0);
}

.tooltip-item strong {
  color: #64B5F6;
}

/* Recorder Styles */
.recorder-container {
  text-align: left;
}

.browse-btn {
  background: #2196F3;
  color: white;
  border: none;
  border-radius: 12px;
  padding: 0 10px;
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 600;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 80px;
}

.browse-btn:hover {
  background: #1976D2;
  transform: scale(1.05);
}

.browse-btn:disabled {
  background: #ccc;
  cursor: not-allowed;
  transform: none;
}

.no-data {
  text-align: center;
  color: #999;
  padding: 20px;
  font-style: italic;
}

/* Merged Input Row */
.joined-input-row {
  display: flex;
  align-items: stretch;
  margin-bottom: 8px; /* Match select-row margin */
  width: 100%;
}

.joined-input-row select {
  flex: 1;
  border-top-right-radius: 0;
  border-bottom-right-radius: 0;
  border-right: none;
  z-index: 1;
  margin: 0; /* Ensure no margins */
}

.joined-input-row .browse-btn {
  border-top-left-radius: 0;
  border-bottom-left-radius: 0;
  margin-left: 0;
  border-left: 1px solid rgba(255, 255, 255, 0.2); /* Add subtle separator */
}

/* Custom Dropdown Styles */
.custom-select-container {
  position: relative;
  flex: 1;
  min-width: 0; /* Critical: allows flex item to shrink below content size */
}

.custom-select-container.disabled {
  opacity: 0.6;
  pointer-events: none;
}

.custom-select-trigger {
  padding: 0.7rem 0.8rem;
  border: 1px solid rgba(0, 0, 0, 0.1);
  border-radius: 12px;
  background-color: rgba(255, 255, 255, 0.8);
  cursor: pointer;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 1rem;
  color: #333;
  /* Match input/select styling in joined-row */
  border-top-right-radius: 0;
  border-bottom-right-radius: 0;
  border-right: none;
  width: 100%;
  box-sizing: border-box;
}

.custom-select-trigger span:first-child {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-right: 8px;
  flex: 1;
  min-width: 0; /* Critical for nested flex overflow */
}

.custom-select-trigger .arrow {
  font-size: 0.8rem;
  color: #666;
}

.custom-options {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: white;
  border: 1px solid rgba(0, 0, 0, 0.1);
  border-radius: 12px;
  margin-top: 4px;
  z-index: 100;
  max-height: 200px;
  overflow-y: auto;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.custom-option {
  padding: 0.7rem 0.8rem;
  cursor: pointer;
  display: flex;
  justify-content: space-between;
  align-items: center;
  transition: background 0.2s;
}

.custom-option:hover {
  background-color: #f5f5f5;
}

.custom-option.selected {
  background-color: #e3f2fd;
  color: #1976D2;
  font-weight: 500;
}

.custom-option.disabled {
  color: #999;
  cursor: default;
  justify-content: center;
}

.custom-option .option-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  margin-right: 8px;
}

.delete-icon {
  color: #999;
  font-weight: bold;
  padding: 4px 8px;
  border-radius: 4px;
  transition: all 0.2s;
  position: relative;
  z-index: 10;
}

.delete-icon:hover {
  color: #f44336;
  background-color: rgba(244, 67, 54, 0.1);
}
</style>