<template>
  <div class="overview-view">
    <!-- Hero Section -->
    <div class="hero-section">
      <div class="hero-content">
        <h1>安全防护中心</h1>
        <p class="hero-subtitle">为 AI 编程助手提供实时安全防护</p>
      </div>
      <div class="hero-glow"></div>
    </div>

    <!-- Status Badge -->
    <div class="status-section">
      <div class="status-badge" :class="{ active: statsData.monitorEnabled }">
        <span class="status-dot"></span>
        <span>{{ statsData.monitorEnabled ? '防护已启用' : '防护已停止' }}</span>
      </div>
    </div>

    <!-- Stats Cards -->
    <div class="stats-grid">
      <div class="stat-card" v-for="(stat, index) in stats" :key="index" :style="{ '--delay': index * 0.1 + 's' }">
        <div class="stat-header">
          <div class="stat-icon" :style="{ background: stat.gradient }">
            <component :is="stat.iconComponent" />
          </div>
          <div class="stat-trend" :class="stat.trend > 0 ? 'up' : 'down'" v-if="stat.trend">
            <component :is="stat.trend > 0 ? TrendUpIcon : TrendDownIcon" />
            {{ Math.abs(stat.trend) }}%
          </div>
        </div>
        <div class="stat-content">
          <span class="stat-value">{{ formatNumber(stat.value) }}</span>
          <span class="stat-label">{{ stat.label }}</span>
        </div>
        <!-- Mini Sparkline -->
        <div class="sparkline">
          <svg viewBox="0 0 100 30" preserveAspectRatio="none">
            <polyline
              fill="none"
              :stroke="stat.color"
              stroke-width="2"
              :points="generateSparkline(stat.history)"
            />
          </svg>
        </div>
      </div>
    </div>

    <!-- Charts Section -->
    <div class="charts-section">
      <!-- System Monitor -->
      <div class="chart-card large">
        <div class="chart-header">
          <h3>系统资源监控</h3>
          <span class="live-dot"><span class="dot-anim"></span>实时</span>
        </div>
        <div class="sys-monitor">
          <!-- CPU -->
          <div class="metric-row">
            <div class="metric-label">
              <span>🖥️ CPU 使用率</span>
              <span class="metric-val" :class="getCpuClass()">{{ sysMetrics.cpu }}%</span>
            </div>
            <div class="metric-bar">
              <div class="metric-fill" :class="getCpuClass()" :style="{ width: sysMetrics.cpu + '%' }"></div>
            </div>
            <div class="metric-sub">{{ sysMetrics.cpuCores }} 核心 &nbsp;·&nbsp; {{ sysMetrics.cpuModel }}</div>
          </div>
          <!-- Memory -->
          <div class="metric-row">
            <div class="metric-label">
              <span>🧠 内存使用率</span>
              <span class="metric-val" :class="getMemClass()">{{ sysMetrics.memPercent }}%</span>
            </div>
            <div class="metric-bar">
              <div class="metric-fill" :class="getMemClass()" :style="{ width: sysMetrics.memPercent + '%' }"></div>
            </div>
            <div class="metric-sub">{{ sysMetrics.memUsed }} MB / {{ sysMetrics.memTotal }} MB</div>
          </div>
          <!-- Uptime -->
          <div class="metric-row">
            <div class="metric-label">
              <span>⏱️ 系统运行时长</span>
              <span class="metric-val ok">{{ formatUptime(sysMetrics.uptime) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Security Score -->
      <div class="chart-card">
        <h3>安全评分</h3>
        <div class="score-container">
          <div class="score-ring">
            <svg viewBox="0 0 100 100">
              <defs>
                <linearGradient id="scoreGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                  <stop offset="0%" :style="{ stopColor: getScoreColor(securityScore) }" />
                  <stop offset="100%" :style="{ stopColor: getScoreColor(securityScore, 0.6) }" />
                </linearGradient>
              </defs>
              <circle class="bg" cx="50" cy="50" r="45" />
              <circle
                class="progress"
                cx="50"
                cy="50"
                r="45"
                stroke="url(#scoreGradient)"
                :stroke-dasharray="283"
                :stroke-dashoffset="283 - (283 * securityScore) / 100"
              />
            </svg>
            <div class="score-text">
              <span class="score-value">{{ securityScore }}</span>
              <span class="score-label">/ 100</span>
            </div>
          </div>
          <div class="score-details">
            <div class="score-item">
              <span class="dot" style="background: #4ade80"></span>
              <span>防火墙</span>
              <span class="status">已启用</span>
            </div>
            <div class="score-item">
              <span class="dot" style="background: #4ade80"></span>
              <span>实时扫描</span>
              <span class="status">已启用</span>
            </div>
            <div class="score-item">
              <span class="dot" style="background: #fbbf24"></span>
              <span>自动更新</span>
              <span class="status">待处理</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Alerts Section -->
    <div class="alerts-section">
      <div class="section-header">
        <h2>最近告警</h2>
        <button class="view-all">查看全部</button>
      </div>
      <div class="alerts-list">
        <div v-if="alerts.length === 0" class="empty-state">
          <div class="empty-icon">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
              <path d="M9 12l2 2 4-4"/>
            </svg>
          </div>
          <p>一切正常，暂无告警</p>
        </div>
        <div
          v-else
          v-for="(alert, index) in alerts"
          :key="index"
          class="alert-item"
          :class="'severity-' + alert.severity"
        >
          <div class="alert-icon" :class="alert.severity">
            <svg v-if="alert.severity === 'high'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"/>
              <line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/>
            </svg>
            <svg v-else-if="alert.severity === 'medium'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
            </svg>
            <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/><path d="M9 12l2 2 4-4"/>
            </svg>
          </div>
          <div class="alert-content">
            <div class="alert-header">
              <span class="alert-type">{{ alert.type }}</span>
              <span class="alert-time">{{ formatTime(alert.timestamp) }}</span>
            </div>
            <p class="alert-message">{{ alert.message }}</p>
          </div>
          <div class="alert-severity" :class="alert.severity">
            {{ alert.severity === 'high' ? '高危' : alert.severity === 'medium' ? '中危' : '低危' }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, h } from 'vue'
import api from '../services/api'

const emit = defineEmits(['navigate'])

// Icons as render functions
const ShieldIcon = () => h('svg', { width: 24, height: 24, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': 2 }, [
  h('path', { d: 'M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z' })
])

const EyeIcon = () => h('svg', { width: 24, height: 24, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': 2 }, [
  h('path', { d: 'M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z' }),
  h('circle', { cx: 12, cy: 12, r: 3 })
])

const AlertIcon = () => h('svg', { width: 24, height: 24, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': 2 }, [
  h('path', { d: 'M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z' }),
  h('line', { x1: 12, y1: 9, x2: 12, y2: 13 }),
  h('line', { x1: 12, y1: 17, x2: 12.01, y2: 17 })
])

const ChartIcon = () => h('svg', { width: 24, height: 24, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': 2 }, [
  h('line', { x1: 18, y1: 20, x2: 18, y2: 10 }),
  h('line', { x1: 12, y1: 20, x2: 12, y2: 4 }),
  h('line', { x1: 6, y1: 20, x2: 6, y2: 14 })
])

const ScanIcon = () => h('svg', { width: 24, height: 24, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': 2 }, [
  h('circle', { cx: 11, cy: 11, r: 8 }),
  h('path', { d: 'M21 21l-4.35-4.35' })
])

const SettingsIcon = () => h('svg', { width: 24, height: 24, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': 2 }, [
  h('circle', { cx: 12, cy: 12, r: 3 }),
  h('path', { d: 'M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-2 2 2 2 0 01-2-2v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83 0 2 2 0 010-2.83l.06-.06a1.65 1.65 0 00.33-1.82 1.65 1.65 0 00-1.51-1H3a2 2 0 01-2-2 2 2 0 012-2h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 010-2.83 2 2 0 012.83 0l.06.06a1.65 1.65 0 001.82.33H9a1.65 1.65 0 001-1.51V3a2 2 0 012-2 2 2 0 012 2v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 0 2 2 0 010 2.83l-.06.06a1.65 1.65 0 00-.33 1.82V9a1.65 1.65 0 001.51 1H21a2 2 0 012 2 2 2 0 01-2 2h-.09a1.65 1.65 0 00-1.51 1z' })
])

const TrendUpIcon = () => h('svg', { width: 14, height: 14, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': 2 }, [
  h('polyline', { points: '23 6 13.5 15.5 8.5 10.5 1 18' }),
  h('polyline', { points: '17 6 23 6 23 12' })
])

const TrendDownIcon = () => h('svg', { width: 14, height: 14, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': 2 }, [
  h('polyline', { points: '23 18 13.5 8.5 8.5 13.5 1 6' }),
  h('polyline', { points: '17 18 23 18 23 12' })
])

const statsData = ref({ monitorEnabled: true, uptime: 0 })
const securityScore = ref(85)
const timeRange = ref('24h')
const activityInterval = ref(null)
const lastScanResult = ref(null)

// Stats with icon components
const stats = ref([
  {
    iconComponent: ScanIcon,
    value: 0,
    label: '累计扫描',
    gradient: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
    color: '#8b5cf6',
    trend: 0,
    history: [0, 0, 0, 0, 0, 0, 0]
  },
  {
    iconComponent: EyeIcon,
    value: 0,
    label: '已监控文件',
    gradient: 'linear-gradient(135deg, #ec4899 0%, #f43f5e 100%)',
    color: '#f43f5e',
    trend: 0,
    history: [0, 0, 0, 0, 0, 0, 0]
  },
  {
    iconComponent: AlertIcon,
    value: 0,
    label: '已拦截威胁',
    gradient: 'linear-gradient(135deg, #f59e0b 0%, #ef4444 100%)',
    color: '#ef4444',
    trend: 0,
    history: [0, 0, 0, 0, 0, 0, 0]
  },
  {
    iconComponent: ChartIcon,
    value: 0,
    label: '操作日志',
    gradient: 'linear-gradient(135deg, #10b981 0%, #06b6d4 100%)',
    color: '#06b6d4',
    trend: 0,
    history: [0, 0, 0, 0, 0, 0, 0]
  }
])

// System metrics
const sysMetrics = ref({ cpu: 0, memUsed: 0, memTotal: 0, memPercent: 0, uptime: 0, cpuCores: 0, cpuModel: '' })

const fetchMetrics = async () => {
  try {
    const m = await api.getSystemMetrics()
    if (m) Object.assign(sysMetrics.value, m)
  } catch (e) {}
}

const getCpuClass = () => {
  const v = sysMetrics.value.cpu
  if (v >= 80) return 'danger'
  if (v >= 50) return 'warn'
  return 'ok'
}
const getMemClass = () => {
  const v = sysMetrics.value.memPercent
  if (v >= 85) return 'danger'
  if (v >= 60) return 'warn'
  return 'ok'
}
const formatUptime = (sec) => {
  if (!sec) return '--'
  const d = Math.floor(sec / 86400)
  const h = Math.floor((sec % 86400) / 3600)
  const m = Math.floor((sec % 3600) / 60)
  return d > 0 ? `${d} 天 ${h} 小时` : `${h} 小时 ${m} 分钟`
}


// Actions removed

// Alerts
const alerts = ref([
  { type: '文件访问', severity: 'high', message: '检测到未授权访问尝试 ~/.ssh', timestamp: Date.now() / 1000 - 300 },
  { type: '命令拦截', severity: 'medium', message: '危险命令 "rm -rf /" 已被拦截', timestamp: Date.now() / 1000 - 1800 },
])

// Methods
const formatNumber = (num) => {
  if (num >= 1000) return (num / 1000).toFixed(1) + 'k'
  return num
}

const generateSparkline = (data) => {
  const max = Math.max(...data)
  return data.map((v, i) => `${(i / (data.length - 1)) * 100},${30 - (v / max) * 28}`).join(' ')
}

const getBarColor = (value) => {
  if (value > 70) return '#ef4444'
  if (value > 40) return '#f59e0b'
  return '#10b981'
}

const getBarLabel = (i) => {
  const hours = ['2时', '4时', '6时', '8时', '10时', '12时', '14时', '16时', '18时', '20时', '22时', '0时']
  return hours[i % hours.length] || ''
}

const getScoreColor = (score, opacity = 1) => {
  if (score >= 80) return `rgba(16, 185, 129, ${opacity})`
  if (score >= 60) return `rgba(245, 158, 11, ${opacity})`
  return `rgba(239, 68, 68, ${opacity})`
}

const formatTime = (timestamp) => {
  if (!timestamp) return ''
  const date = new Date(timestamp * 1000)
  return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

const handleAction = (id) => {
  emit('navigate', id)
}

// Generate random activity
const generateActivity = () => {
  activityData.value = activityData.value.map(() => Math.random() * 80 + 20)
}

onMounted(async () => {
  try {
    // 获取监控状态
    const statusData = await api.getMonitorStatus()
    const status = statusData.status || {}
    statsData.value = status
    statsData.value.monitorEnabled = status.enabled !== false

    // 获取扫描结果 (真实数据)
    try {
      const scanResult = await api.scan()
      lastScanResult.value = scanResult.result || scanResult

      // 更新统计数据
      if (lastScanResult.value) {
        const summary = lastScanResult.value.summary || {}
        stats.value[0].value = 1  // Total Scans
        stats.value[2].value = summary.high_risks || 0  // Threats
        securityScore.value = Math.max(0, 100 - (summary.high_risks || 0) * 20 - (summary.medium_risks || 0) * 10)
      }
    } catch (scanErr) {
      console.warn('Scan failed:', scanErr)
    }

    // 获取日志
    try {
      const logsData = await api.getLogs()
      const logs = logsData.logs || []
      stats.value[3].value = logs.length  // 操作日志总数
      stats.value[1].value = 4  // 监控目录数（固定：.claude/.openclaw/Projects/code）

      // 用真实日志重建时间线
      activityData.value = buildActivityFromLogs(logs)

      // 设置告警（来自被拦截的操作）
      const blocked = logs.filter(l => !l.allowed)
      stats.value[2].value = blocked.length  // 已拦截威胁

      if (blocked.length > 0) {
        alerts.value = blocked.slice(0, 5).map(log => ({
          type: log.type === 'write' ? '文件写入' : log.type === 'read' ? '文件读取' : '文件访问',
          severity: 'high',
          message: (log.reason || '黑名单路径') + ': ' + (log.path || ''),
          timestamp: log.timestamp
        }))
      } else {
        alerts.value = []
      }
    } catch (logErr) {
      console.warn('Logs fetch failed:', logErr)
    }

  } catch (e) {
    console.error('Failed to get status:', e)
  }

  // 每 30 秒用真实数据刷新一次时间线（不再随机）
  activityInterval.value = setInterval(async () => {
    try {
      const logsData = await api.getLogs()
      const logs = logsData.logs || []
      stats.value[3].value = logs.length
    } catch (e) {}
  }, 30000)

  // 系统资源监控：每 3 秒刷新一次
  await fetchMetrics()
  const metricsTimer = setInterval(fetchMetrics, 3000)
  // 把 metrics timer 也存起来以便卸载
  activityInterval._metricsTimer = metricsTimer
})

onUnmounted(() => {
  if (activityInterval.value) clearInterval(activityInterval.value)
  if (activityInterval._metricsTimer) clearInterval(activityInterval._metricsTimer)
})
</script>

<style scoped>
.overview-view {
  max-width: 1100px;
  margin: 0 auto;
  padding-bottom: 40px;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

/* Hero Section */
.hero-section {
  position: relative;
  text-align: center;
  padding: 40px 20px;
  margin-bottom: 20px;
}

.hero-content h1 {
  font-size: 36px;
  font-weight: 700;
  margin: 0 0 8px;
  color: #fff;
  letter-spacing: -0.5px;
}

.hero-subtitle {
  color: #6b7280;
  font-size: 16px;
  margin: 0;
}

.hero-glow {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 600px;
  height: 300px;
  background: radial-gradient(ellipse, rgba(99, 102, 241, 0.12) 0%, transparent 70%);
  pointer-events: none;
}

/* Status Badge */
.status-section {
  display: flex;
  justify-content: center;
  margin-bottom: 24px;
}

.status-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 20px;
  background: rgba(31, 41, 55, 0.5);
  border: 1px solid rgba(55, 65, 81, 0.5);
  border-radius: 50px;
  font-size: 13px;
  color: #9ca3af;
  transition: all 0.3s;
}

.status-badge.active {
  color: #10b981;
  border-color: rgba(16, 185, 129, 0.3);
  background: rgba(16, 185, 129, 0.1);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #6b7280;
}

.status-badge.active .status-dot {
  background: #10b981;
  box-shadow: 0 0 12px #10b981;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

/* Stats Grid */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  background: rgba(31, 41, 55, 0.5);
  border: 1px solid rgba(55, 65, 81, 0.5);
  border-radius: 16px;
  padding: 20px;
  transition: all 0.3s;
  animation: fadeInUp 0.5s ease forwards;
  animation-delay: var(--delay);
  opacity: 0;
}

.stat-card:hover {
  transform: translateY(-4px);
  border-color: rgba(99, 102, 241, 0.3);
  background: rgba(31, 41, 55, 0.8);
}

@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

.stat-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;
}

.stat-icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}

.stat-trend {
  display: flex;
  align-items: center;
  gap: 2px;
  font-size: 12px;
  padding: 4px 8px;
  border-radius: 20px;
  background: rgba(31, 41, 55, 0.8);
}

.stat-trend.up { color: #10b981; }
.stat-trend.down { color: #ef4444; }

.stat-content {
  position: relative;
  z-index: 1;
}

.stat-value {
  display: block;
  font-size: 32px;
  font-weight: 700;
  color: #fff;
  line-height: 1.1;
}

.stat-label {
  font-size: 13px;
  color: #6b7280;
}

.sparkline {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 30px;
  opacity: 0.5;
}

.sparkline svg {
  width: 100%;
  height: 100%;
}

/* Charts Section */
.charts-section {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 16px;
  margin-bottom: 24px;
}

.chart-card {
  background: rgba(31, 41, 55, 0.5);
  border: 1px solid rgba(55, 65, 81, 0.5);
  border-radius: 16px;
  padding: 24px;
}

.chart-card.large {
  grid-column: 1;
}

/* Live dot */
.live-dot {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #10b981;
}
.dot-anim {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #10b981;
  animation: pulse-dot 1.5s ease-in-out infinite;
}
@keyframes pulse-dot {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.4; transform: scale(0.7); }
}

/* System monitor */
.sys-monitor {
  display: flex;
  flex-direction: column;
  gap: 22px;
}
.metric-row {}
.metric-label {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-size: 13px;
  color: #9ca3af;
}
.metric-val {
  font-weight: 700;
  font-size: 15px;
}
.metric-val.ok     { color: #10b981; }
.metric-val.warn   { color: #f59e0b; }
.metric-val.danger { color: #ef4444; }
.metric-bar {
  height: 8px;
  background: rgba(255,255,255,0.07);
  border-radius: 4px;
  overflow: hidden;
}
.metric-fill {
  height: 100%;
  border-radius: 4px;
  transition: width 0.6s ease;
}
.metric-fill.ok     { background: linear-gradient(90deg, #059669, #10b981); }
.metric-fill.warn   { background: linear-gradient(90deg, #d97706, #f59e0b); }
.metric-fill.danger { background: linear-gradient(90deg, #dc2626, #ef4444); }
.metric-sub {
  margin-top: 6px;
  font-size: 11px;
  color: #6b7280;
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.chart-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: #d1d5db;
  margin: 0;
}

.chart-tabs {
  display: flex;
  gap: 4px;
  background: rgba(17, 24, 39, 0.5);
  padding: 4px;
  border-radius: 8px;
}

.chart-tabs button {
  padding: 6px 12px;
  background: transparent;
  border: none;
  color: #6b7280;
  font-size: 12px;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.2s;
}

.chart-tabs button.active {
  background: rgba(99, 102, 241, 0.2);
  color: #a5b4fc;
}

.chart-body {
  height: 180px;
}

.timeline-chart {
  display: flex;
  height: 100%;
}

.chart-y-axis {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding-right: 12px;
  font-size: 11px;
  color: #4b5563;
}

.chart-bars {
  flex: 1;
  display: flex;
  align-items: flex-end;
  gap: 4px;
}

.bar-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  height: 100%;
  justify-content: flex-end;
}

.bar {
  width: 100%;
  border-radius: 4px 4px 0 0;
  animation: barGrow 0.5s ease forwards;
  animation-fill-mode: backwards;
}

@keyframes barGrow {
  from { height: 0; }
}

.bar-label {
  font-size: 10px;
  color: #4b5563;
  margin-top: 8px;
}

/* Score Ring */
.score-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 24px;
}

.score-ring {
  position: relative;
  width: 140px;
  height: 140px;
}

.score-ring svg {
  transform: rotate(-90deg);
  width: 100%;
  height: 100%;
}

.score-ring .bg {
  fill: none;
  stroke: rgba(55, 65, 81, 0.8);
  stroke-width: 8;
}

.score-ring .progress {
  fill: none;
  stroke-width: 8;
  stroke-linecap: round;
  transition: stroke-dashoffset 1s ease;
}

.score-text {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
}

.score-value {
  display: block;
  font-size: 42px;
  font-weight: 700;
  color: #fff;
  line-height: 1;
}

.score-label {
  font-size: 14px;
  color: #6b7280;
}

.score-details {
  width: 100%;
}

.score-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 0;
  font-size: 13px;
  color: #9ca3af;
  border-bottom: 1px solid rgba(55, 65, 81, 0.5);
}

.score-item:last-child {
  border-bottom: none;
}

.score-item .dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.score-item .status {
  margin-left: auto;
  color: #10b981;
}

/* Actions Section */
.actions-section {
  margin-bottom: 24px;
}

.actions-section h2 {
  font-size: 16px;
  font-weight: 600;
  color: #9ca3af;
  margin-bottom: 16px;
}

.actions-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.action-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: rgba(31, 41, 55, 0.5);
  border: 1px solid rgba(55, 65, 81, 0.5);
  border-radius: 16px;
  cursor: pointer;
  transition: all 0.3s;
  animation: fadeInUp 0.5s ease forwards;
  animation-delay: var(--delay);
  opacity: 0;
}

.action-card:hover {
  border-color: rgba(99, 102, 241, 0.4);
  transform: translateY(-2px);
}

.action-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
}

.action-content {
  flex: 1;
}

.action-content h3 {
  font-size: 15px;
  font-weight: 600;
  color: #e5e7eb;
  margin: 0 0 4px;
}

.action-content p {
  font-size: 13px;
  color: #6b7280;
  margin: 0;
}

.action-arrow {
  color: #4b5563;
  transition: all 0.3s;
}

.action-card:hover .action-arrow {
  color: #a5b4fc;
  transform: translateX(4px);
}

/* Alerts Section */
.alerts-section {
  background: rgba(31, 41, 55, 0.5);
  border: 1px solid rgba(55, 65, 81, 0.5);
  border-radius: 16px;
  padding: 24px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.section-header h2 {
  font-size: 16px;
  font-weight: 600;
  color: #d1d5db;
  margin: 0;
}

.view-all {
  background: none;
  border: none;
  color: #6366f1;
  cursor: pointer;
  font-size: 13px;
  transition: color 0.2s;
}

.view-all:hover {
  color: #a5b4fc;
}

.empty-state {
  text-align: center;
  padding: 40px;
}

.empty-state svg {
  color: #10b981;
  margin-bottom: 12px;
}

.empty-state p {
  color: #6b7280;
  margin: 0;
}

.alerts-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.alert-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  background: rgba(17, 24, 39, 0.5);
  border-radius: 12px;
  border-left: 3px solid;
}

.alert-item.severity-high {
  border-left-color: #ef4444;
}

.alert-item.severity-medium {
  border-left-color: #f59e0b;
}

.alert-item.severity-low {
  border-left-color: #10b981;
}

.alert-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.alert-icon.high {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}

.alert-icon.medium {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
}

.alert-icon.low {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
}

.alert-content {
  flex: 1;
}

.alert-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 4px;
}

.alert-type {
  font-weight: 600;
  font-size: 14px;
  color: #e5e7eb;
}

.alert-time {
  font-size: 12px;
  color: #6b7280;
}

.alert-message {
  font-size: 13px;
  color: #9ca3af;
  margin: 0;
}

.alert-severity {
  font-size: 10px;
  font-weight: 600;
  padding: 4px 10px;
  border-radius: 20px;
  flex-shrink: 0;
}

.alert-severity.high {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}

.alert-severity.medium {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
}

.alert-severity.low {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
}

/* Responsive */
@media (max-width: 768px) {
  .stats-grid,
  .charts-section,
  .actions-grid {
    grid-template-columns: 1fr;
  }

  .hero-section h1 {
    font-size: 28px;
  }
}
</style>
