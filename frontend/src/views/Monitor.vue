<template>
  <div class="monitor-view">
    <div class="monitor-header">
      <h1>📊 实时监控</h1>
      <div class="monitor-toggle">
        <span class="status-indicator" :class="{ active: status.enabled }"></span>
        <span>{{ status.enabled ? '监控中' : '已停止' }}</span>
        <button @click="toggleMonitor" class="btn-toggle" :class="{ active: status.enabled }">
          {{ status.enabled ? '停止' : '启动' }}
        </button>
      </div>
    </div>

    <div v-if="status.enabled" class="stats-grid">
      <div class="stat-card">
        <div class="stat-value">{{ status.total_operations }}</div>
        <div class="stat-label">总操作数</div>
      </div>
      <div class="stat-card success">
        <div class="stat-value">{{ status.allowed_count }}</div>
        <div class="stat-label">已放行</div>
      </div>
      <div class="stat-card danger">
        <div class="stat-value">{{ status.blocked_count }}</div>
        <div class="stat-label">已拦截</div>
      </div>
    </div>

    <div class="logs-section">
      <div class="section-header">
        <h3>操作日志</h3>
        <div class="header-buttons">
          <button @click="clearLogs" class="btn-clear">清空</button>
          <button @click="refreshLogs" class="btn-refresh">刷新</button>
        </div>
      </div>

      <div v-if="logs.length === 0" class="empty-state">
        <p>暂无操作日志</p>
      </div>

      <div v-else class="logs-list">
        <div v-for="log in logs" :key="log.id" class="log-item" :class="{ blocked: !log.allowed }">
          <div class="log-header">
            <span class="log-status">{{ log.allowed ? '✓ 放行' : '✕ 拦截' }}</span>
            <span class="log-time">{{ formatTime(log.timestamp) }}</span>
          </div>
          <div class="log-content">
            <span class="log-type">{{ getTypeLabel(log.type) }}</span>
            <span v-if="log.path" class="log-path">{{ log.path }}</span>
            <span v-if="log.command" class="log-command">{{ log.command }}</span>
          </div>
          <div v-if="log.reason" class="log-reason">
            原因: {{ log.reason }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import api from '../services/api'

const status = ref({
  enabled: false,
  running: false,
  total_operations: 0,
  blocked_count: 0,
  allowed_count: 0
})

const logs = ref([])

const loadStatus = async () => {
  try {
    const data = await api.getMonitorStatus()
    const s = data.status || data
    status.value.enabled = s.enabled !== false
    status.value.running = status.value.enabled
    status.value.total_operations = s.total_operations || 0
    status.value.blocked_count = s.blocked_count || 0
    status.value.allowed_count = s.allowed_count || 0
  } catch (error) {
    console.error('加载状态失败:', error)
  }
}

const loadLogs = async () => {
  try {
    const data = await api.getLogs()
    logs.value = data.logs || []
  } catch (error) {
    console.error('加载日志失败:', error)
  }
}

const refreshLogs = () => {
  loadStatus()
  loadLogs()
}

const clearLogs = async () => {
  try {
    await api.clearLogs()
    logs.value = []
    status.value.total_operations = 0
    status.value.blocked_count = 0
    status.value.allowed_count = 0
  } catch (error) {
    console.error('清空日志失败:', error)
  }
}

const toggleMonitor = async () => {
  const newEnabled = !status.value.enabled
  try {
    await api.setMonitorEnabled(newEnabled)
    status.value.enabled = newEnabled
    status.value.running = newEnabled
    await loadStatus()
    await loadLogs()
  } catch (error) {
    console.error('切换监控状态失败:', error)
  }
}

const formatTime = (timestamp) => {
  // timestamp 是 Unix 秒，需要乘以 1000 转换为毫秒
  const date = new Date(timestamp * 1000)
  return date.toLocaleTimeString('zh-CN')
}

const getTypeLabel = (type) => {
  const labels = {
    read: '📖 读取',
    write: '✏️ 写入',
    delete: '🗑️ 删除',
    execute: '⚡ 执行'
  }
  return labels[type] || type
}

const refreshInterval = ref(null)

onMounted(() => {
  loadStatus()
  loadLogs()

  // 定时刷新
  refreshInterval.value = setInterval(() => {
    if (status.value.running) {
      loadStatus()
      loadLogs()
    }
  }, 5000)
})

onUnmounted(() => {
  if (refreshInterval.value) {
    clearInterval(refreshInterval.value)
  }
})
</script>

<style scoped>
.monitor-view {
  max-width: 1000px;
  margin: 0 auto;
}

.monitor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}

.monitor-header h1 {
  font-size: 28px;
}

.monitor-toggle {
  display: flex;
  align-items: center;
  gap: 12px;
}

.status-indicator {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #666;
}

.status-indicator.active {
  background: #4CAF50;
  box-shadow: 0 0 10px #4CAF50;
}

.btn-toggle {
  padding: 8px 20px;
  border: 1px solid #4CAF50;
  background: transparent;
  color: #4CAF50;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-toggle.active {
  background: rgba(244, 63, 54, 0.2);
  border-color: #f44336;
  color: #f44336;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
  margin-bottom: 30px;
}

.stat-card {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 12px;
  padding: 24px;
  text-align: center;
}

.stat-card.success {
  border: 1px solid rgba(76, 175, 80, 0.3);
}

.stat-card.danger {
  border: 1px solid rgba(244, 63, 54, 0.3);
}

.stat-value {
  font-size: 36px;
  font-weight: bold;
  margin-bottom: 8px;
}

.stat-label {
  color: #888;
  font-size: 14px;
}

.logs-section {
  background: rgba(255, 255, 255, 0.03);
  border-radius: 12px;
  padding: 20px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.section-header h3 {
  font-size: 18px;
}

.header-buttons {
  display: flex;
  gap: 8px;
}

.btn-clear {
  padding: 6px 16px;
  background: rgba(244, 63, 54, 0.2);
  border: 1px solid rgba(244, 63, 54, 0.5);
  border-radius: 6px;
  color: #f44336;
  cursor: pointer;
}

.btn-clear:hover {
  background: rgba(244, 63, 54, 0.3);
}

.btn-refresh {
  padding: 6px 16px;
  background: rgba(255, 255, 255, 0.1);
  border: none;
  border-radius: 6px;
  color: white;
  cursor: pointer;
}

.btn-refresh:hover {
  background: rgba(255, 255, 255, 0.2);
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #666;
}

.logs-list {
  max-height: 500px;
  overflow-y: auto;
}

.log-item {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
  padding: 12px 16px;
  margin-bottom: 8px;
  border-left: 3px solid #4CAF50;
}

.log-item.blocked {
  border-left-color: #f44336;
  background: rgba(244, 63, 54, 0.1);
}

.log-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.log-status {
  font-weight: 600;
}

.log-time {
  color: #888;
  font-size: 14px;
}

.log-content {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  color: #ccc;
}

.log-type {
  background: rgba(255, 255, 255, 0.1);
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 13px;
}

.log-path,
.log-command {
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 13px;
  color: #aaa;
}

.log-reason {
  margin-top: 8px;
  color: #f44336;
  font-size: 13px;
}
</style>
