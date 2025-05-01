<template>
  <div class="p-4">
    <div class="flex border-b mb-4">
      <button
        class="px-4 py-2"
        :class="tab === 0 ? 'border-b-2 border-blue-500 font-bold' : ''"
        @click="tab = 0"
      >MQTT Alert</button>
      <!-- 可扩展其他Tab -->
    </div>
    <div v-if="tab === 0">
      <details class="mb-4 border rounded">
        <summary class="cursor-pointer font-semibold p-3">MQTT Topic 配置</summary>
        <div class="p-3 pt-0">
          <div class="flex gap-2">
            <input
              v-model="mqttTopic"
              placeholder="输入要订阅的 Topic (例如: sensors/temperature)"
              class="border px-2 py-1 rounded flex-grow"
            />
            <button 
              @click="updateTopic" 
              :disabled="subscribing"
              class="bg-blue-500 text-white px-4 py-1 rounded hover:bg-blue-600 active:bg-blue-700 transition-colors"
              :class="{'opacity-75 cursor-not-allowed': subscribing}"
            >
              <span v-if="subscribing">订阅中...</span>
              <span v-else>确认订阅</span>
            </button>
          </div>
          <div class="mt-2 flex items-center">
            <span class="text-sm text-gray-600">当前订阅: {{ currentTopic || "无" }}</span>
            <!-- 成功提示 -->
            <div v-if="showSuccess" class="ml-3 text-sm text-green-600 flex items-center animate-fade-in">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
              订阅成功!
            </div>
          </div>
        </div>
      </details>
      <details class="mb-4" open>
        <summary class="cursor-pointer font-semibold">Add Alert Rule</summary>
        <form @submit.prevent="addAlert" class="space-y-2 mt-2">
          <div>
            <input
              v-model="form.keyword"
              placeholder="Keyword (separate multiple with commas)"
              required
              class="border px-2 py-1 rounded w-full"
            />
            <div class="text-xs text-gray-500 mt-1">Multiple keywords can be separated by commas (e.g. "error,warning,alert")</div>
          </div>
          <div>
            <label>
              <input type="radio" value="builtin" v-model="form.musicType" />
              Built-in
            </label>
            <label class="ml-4">
              <input type="radio" value="upload" v-model="form.musicType" />
              Upload
            </label>
          </div>
          <div v-if="form.musicType === 'builtin'">
            <select v-model="form.music" class="border px-2 py-1 rounded w-full">
              <option v-for="m in builtinMusic" :key="m.value" :value="m.value">{{ m.label }}</option>
            </select>
          </div>
          <div v-else>
            <input type="file" @change="onFileChange" accept="audio/*" />
            <span v-if="form.musicFile">{{ form.musicFile.name }}</span>
          </div>
          <div>
            <label>Volume: {{ Math.round(form.volume * 100) }}%</label>
            <input type="range" min="0" max="1" step="0.01" v-model.number="form.volume" />
          </div>
          <button type="submit" class="bg-blue-500 text-white px-4 py-1 rounded">Add</button>
        </form>
      </details>
      <div v-if="alerts.length" class="space-y-2">
        <div
          v-for="(alert, idx) in alerts"
          :key="idx"
          class="border rounded p-2 flex items-center justify-between"
        >
          <div v-if="editingIndex !== idx">
            <div><b>Keyword:</b> {{ alert.keyword }}</div>
            <div><b>Music:</b> {{ alert.musicLabel }}</div>
            <div><b>Volume:</b> {{ Math.round(alert.volume * 100) }}%</div>
          </div>
          <div v-else class="flex flex-col gap-1">
            <input v-model="editingForm.keyword" class="border px-2 py-1 rounded" />
            <div>
              <label>
                <input type="radio" value="builtin" v-model="editingForm.musicType" />
                Built-in
              </label>
              <label class="ml-4">
                <input type="radio" value="upload" v-model="editingForm.musicType" />
                Upload
              </label>
            </div>
            <div v-if="editingForm.musicType === 'builtin'">
              <select v-model="editingForm.music" class="border px-2 py-1 rounded w-full">
                <option v-for="m in builtinMusic" :key="m.value" :value="m.value">{{ m.label }}</option>
              </select>
            </div>
            <div v-else>
              <input type="file" @change="onEditFileChange" accept="audio/*" />
              <span v-if="editingForm.musicFile && typeof editingForm.musicFile !== 'string'">{{ editingForm.musicFile.name }}</span>
              <span v-else-if="editingForm.musicFile">{{ editingForm.musicFile }}</span>
            </div>
            <div>
              <label>Volume: {{ Math.round(editingForm.volume * 100) }}%</label>
              <input type="range" min="0" max="1" step="0.01" v-model.number="editingForm.volume" />
            </div>
            <div class="flex gap-2 mt-1">
              <button class="bg-green-500 text-white px-2 py-1 rounded" @click="saveEdit(idx)">Save</button>
              <button class="bg-gray-300 px-2 py-1 rounded" @click="cancelEdit">Cancel</button>
            </div>
          </div>
          <div class="flex gap-2 items-center">
            <button @click="startEdit(idx)" v-if="editingIndex !== idx" title="Edit">
              <Edit class="w-5 h-5 text-blue-500" />
            </button>
            <button @click="removeAlert(idx)" title="Delete">
              <Trash2 class="w-5 h-5 text-red-500" />
            </button>
            <button @click="stopMusic" title="Stop Music">
              <StopCircle class="w-5 h-5 text-gray-700" />
            </button>
          </div>
        </div>
      </div>
      <div v-else class="text-gray-500">No alert rules yet.</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import mqtt from 'mqtt'
import { saveFileToIDB, getFileFromIDB } from '../utils/idb'
import { Trash2, Edit, StopCircle } from 'lucide-vue-next'

const tab = ref(0)

const builtinMusic = [
 { value: 'beep', label: 'Beep Sound', src: 'beep' },

  // { value: 'alarm1', label: 'didi', src: '/alarm1.mp3' },
]

const form = ref({
  keyword: '',
  musicType: 'builtin',
  music: 'alarm1',
  musicFile: null as File | null,
  volume: 0.5,
})

const alerts = ref<any[]>([])

const editingIndex = ref<number | null>(null)
const editingForm = ref<any>(null)
const currentAudio = ref<HTMLAudioElement | null>(null)

// MQTT Topic 相关
const mqttTopic = ref('icomefrom') // 默认值
const currentTopic = ref('')
const subscribing = ref(false)
const showSuccess = ref(false)

// 用于 Beep 音效的对象
let audioCtx: AudioContext | null = null
let oscillator: OscillatorNode | null = null
let gainNode: GainNode | null = null

// 增加一个标记当前是否在播放 beep
const isBeepPlaying = ref(false)

function saveAlerts() {
  localStorage.setItem('mqtt_alerts', JSON.stringify(alerts.value))
}
function loadAlerts() {
  const data = localStorage.getItem('mqtt_alerts')
  if (data) alerts.value = JSON.parse(data)
}
async function addAlert() {
  let musicSrc = ''
  let musicLabel = ''
  let musicFileName = null
  
  if (form.value.musicType === 'builtin') {
    const m = builtinMusic.find(m => m.value === form.value.music)
    if (m) {
      musicSrc = m.src
      musicLabel = m.label
    }
  } else if (form.value.musicFile) {
    musicFileName = form.value.musicFile.name
    await saveFileToIDB(musicFileName, form.value.musicFile)
    musicSrc = musicFileName
    musicLabel = form.value.musicFile.name
  }
  
  alerts.value.push({
    keyword: form.value.keyword.trim(),
    musicSrc,
    musicLabel,
    volume: form.value.volume,
    musicType: form.value.musicType,
    music: form.value.music,
    musicFile: musicFileName,
  })
  saveAlerts()
  // 重置表单
  form.value.keyword = ''
  form.value.musicType = 'builtin'
  form.value.music = 'alarm1'
  form.value.musicFile = null
  form.value.volume = 0.5
}
function removeAlert(idx: number) {
  alerts.value.splice(idx, 1)
  saveAlerts()
}
function onFileChange(e: Event) {
  const files = (e.target as HTMLInputElement).files
  if (files && files[0]) {
    form.value.musicFile = files[0]
  }
}

let client: mqtt.MqttClient | null = null

function preloadBuiltinAudio() {
  builtinMusic.forEach(music => {
    if (music.src !== 'beep') {
      const audio = new Audio(music.src);
      audio.preload = 'auto';
      audio.load();
      console.log(`Preloading audio: ${music.label} from ${music.src}`);
    }
  });
}

onMounted(() => {
  loadAlerts()
  connectMqtt()
  preloadBuiltinAudio()
})

onUnmounted(() => {
  if (client) {
    client.end()
  }
  
  // 确保关闭任何正在播放的 beep
  stopBeep()
  
  // 关闭 AudioContext
  if (audioCtx) {
    audioCtx.close()
    audioCtx = null
  }
})

function updateTopic() {
  if (!mqttTopic.value) {
    alert('Topic 不能为空')
    return
  }
  
  if (client) {
    subscribing.value = true
    
    // 先取消之前的订阅
    if (currentTopic.value) {
      client.unsubscribe(currentTopic.value)
    }
    
    // 订阅新的 Topic
    client.subscribe(mqttTopic.value, (err) => {
      subscribing.value = false
      
      if (err) {
        alert(`订阅失败: ${err.message}`)
        return
      }
      
      currentTopic.value = mqttTopic.value
      // 保存到 localStorage
      localStorage.setItem('mqtt_topic', mqttTopic.value)
      
      // 显示成功提示
      showSuccess.value = true
      setTimeout(() => {
        showSuccess.value = false
      }, 3000) // 3秒后自动消失
    })
  } else {
    alert('MQTT 客户端未连接，请稍后再试')
  }
}

function connectMqtt() {
  // 使用 MQTT 库连接到 broker
  client = mqtt.connect('wss://broker.emqx.io:8084/mqtt')
  
  client.on('connect', () => {
    console.log('Connected to MQTT broker')
    // 加载保存的 Topic
    const savedTopic = localStorage.getItem('mqtt_topic')
    if (savedTopic) {
      mqttTopic.value = savedTopic
      currentTopic.value = savedTopic
      client?.subscribe(savedTopic)
    } else {
      // 默认订阅
      client?.subscribe('icomefrom')
      currentTopic.value = 'icomefrom'
    }
  })
  
  client.on('message', (_topic:any, message) => {
    const msg = message.toString()
    // 检查消息是否包含关键词
    for (const alert of alerts.value) {
      if (alert.keyword) {
        // 分割关键词并逐个检查
        // @ts-ignore
        const keywords = alert.keyword.split(',').map(kw => kw.trim()).filter(kw => kw)
        for (const keyword of keywords) {
          if (msg.includes(keyword)) {
            playMusic(alert, alert.volume)
            break // 一旦匹配到一个关键词，就停止检查并播放音乐
          }
        }
      }
    }
  })
  
  client.on('error', (err) => {
    console.error('MQTT connection error:', err)
  })
}

function initAudio() {
  // @ts-ignore
  if (!audioCtx) audioCtx = new (window.AudioContext || window.webkitAudioContext)();
}

function playBeep(volume = 0.2) {
  stopBeep();
  initAudio();
  if (!audioCtx) return;
  
  // 创建主振荡器
  oscillator = audioCtx.createOscillator();
  gainNode = audioCtx.createGain();
  
  // 创建调制振荡器
  const modulator = audioCtx.createOscillator();
  modulator.type = 'sine';
  modulator.frequency.value = 2; // 每秒变化2次
  
  // 创建调制增益节点
  const modulationGain = audioCtx.createGain();
  modulationGain.gain.value = 50; // 调制深度
  
  // 连接调制链
  modulator.connect(modulationGain);
  modulationGain.connect(oscillator.frequency);
  
  // 设置主振荡器
  oscillator.type = 'sine';
  oscillator.frequency.value = 440; // 基础频率
  gainNode.gain.value = volume;
  
  // 连接输出并开始播放
  oscillator.connect(gainNode).connect(audioCtx.destination);
  oscillator.start();
  modulator.start();
}

function stopBeep() {
  if (oscillator) {
    oscillator.stop();
    oscillator.disconnect();
    if (gainNode) gainNode.disconnect();
    oscillator = null;
    gainNode = null;
  }
}

async function playMusic(alert: any, volume: number) {
  // 处理 beep 特殊情况
  if (alert.musicType === 'builtin' && (alert.musicSrc === 'beep' || alert.music === 'beep')) {
    playBeep(volume);
    isBeepPlaying.value = true;
    return;
  }
  
  // 处理其他音频文件
  let src = ''
  
  if (alert.musicType === 'builtin') {
    // 对于内置音乐，确保使用正确的src
    if (alert.music) {
      // 如果存储了音乐类型，根据它找src
      const music = builtinMusic.find(m => m.value === alert.music);
      src = music?.src || alert.musicSrc;
    } else {
      // 向后兼容，使用musicSrc
      src = alert.musicSrc;
    }
  } else if (alert.musicType === 'upload' && alert.musicFile) {
    const file = await getFileFromIDB(alert.musicFile);
    if (file) {
      src = URL.createObjectURL(file);
    } else {
      alert(`上传的音频文件 "${alert.musicFile}" 丢失，请编辑此规则并重新上传音频。`);
      return;
    }
  }
  
  if (!src) {
    alert(`无法确定音频源，请编辑此规则并选择有效的音频。`);
    return;
  }
  
  if (currentAudio.value) {
    currentAudio.value.pause();
    currentAudio.value.currentTime = 0;
  }
  
  // 停止任何正在播放的 beep
  stopBeep();
  isBeepPlaying.value = false;
  
  try {
    const audio = new Audio(src);
    
    // 添加错误处理
    audio.onerror = (e) => {
      console.error('Audio playback error:', e);
      if (alert.musicType === 'builtin') {
        alert(`内置音频 "${alert.musicLabel}" 无法播放，文件可能不存在。请编辑此规则并选择其他音频。`);
      } else {
        alert(`音频 "${alert.musicLabel}" 无法播放。请编辑此规则并选择其他音频。`);
      }
    };
    
    audio.volume = volume;
    await audio.play();
    currentAudio.value = audio;
  } catch (err) {
    console.error('播放音频失败:', err);
    if (alert.musicType === 'builtin') {
      alert(`内置音频 "${alert.musicLabel}" 无法播放，文件可能不存在。请编辑此规则并选择其他音频。`);
    } else {
      alert(`音频 "${alert.musicLabel}" 无法播放。请编辑此规则并选择其他音频。`);
    }
  }
}

function stopMusic() {
  if (currentAudio.value) {
    currentAudio.value.pause();
    currentAudio.value.currentTime = 0;
    currentAudio.value = null;
  }
  
  // 停止任何 beep 音效
  stopBeep();
  isBeepPlaying.value = false;
}

function startEdit(idx: number) {
  editingIndex.value = idx
  editingForm.value = { ...alerts.value[idx] }
}

async function saveEdit(idx: number) {
  if (editingForm.value.musicType === 'upload' && editingForm.value.musicFile instanceof File) {
    await saveFileToIDB(editingForm.value.musicFile.name, editingForm.value.musicFile)
    editingForm.value.musicSrc = editingForm.value.musicFile.name
    editingForm.value.musicLabel = editingForm.value.musicFile.name
    editingForm.value.musicFile = editingForm.value.musicFile.name
  }
  // 保存内置音乐的选择值
  if (editingForm.value.musicType === 'builtin') {
    editingForm.value.music = editingForm.value.music || editingForm.value.musicSrc
  }
  
  alerts.value[idx] = { ...editingForm.value }
  saveAlerts()
  editingIndex.value = null
  editingForm.value = null
}

function cancelEdit() {
  editingIndex.value = null
  editingForm.value = null
}

function onEditFileChange(e: Event) {
  const files = (e.target as HTMLInputElement).files
  if (files && files[0]) {
    editingForm.value.musicFile = files[0]
  }
}
</script>

<style scoped>
.animate-fade-in {
  animation: fadeIn 0.5s ease-in-out;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
</style> 