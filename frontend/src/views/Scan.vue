<template>
  <div class="scan-view">
    <div class="scan-header">
      <h1>🛡️ 安全扫描</h1>
      <p class="subtitle">深度检测 AI 助手配置中的安全风险</p>
    </div>

    <!-- 未开始 -->
    <div v-if="!scanning && !result" class="scan-start">
      <div class="scan-info">
        <h3>扫描范围</h3>
        <ul>
          <li>🔍 MCP Server 配置审计</li>
          <li>🔐 Claude 工具权限审查</li>
          <li>🗝️ 多平台 API Key 泄漏检测（Anthropic / OpenAI / AWS / GitHub 等）</li>
          <li>💻 Shell 历史危险命令审计</li>
          <li>📄 .env 文件密钥扫描</li>
          <li>🔒 SSH 安全配置检查</li>
          <li>👤 对话历史 PII 数据扫描</li>
          <li>📁 敏感文件与目录权限审查</li>
        </ul>
      </div>
      <button @click="startScan" class="btn btn-primary">开始扫描</button>
    </div>

    <!-- 扫描中 -->
    <div v-if="scanning" class="scan-progress">
      <div class="spinner"></div>
      <p class="progress-title">正在扫描中...</p>

      <!-- 步骤进度条 -->
      <div class="progress-bar-wrap">
        <div class="progress-bar" :style="{ width: progressPct + '%' }"></div>
      </div>
      <p class="progress-pct">{{ progressStep }} / {{ progressTotal }} 模块</p>

      <!-- 步骤列表 -->
      <div class="steps-list">
        <div
          v-for="(step, i) in progressSteps"
          :key="i"
          class="step-item"
          :class="{
            done:    i < progressStep - 1,
            active:  i === progressStep - 1,
            pending: i >= progressStep
          }"
        >
          <span class="step-dot"></span>
          <span class="step-label">{{ step }}</span>
        </div>
      </div>
    </div>

    <!-- 结果 -->
    <div v-if="result" class="scan-result">
      <div class="result-summary" :class="getSummaryClass()">
        <h2>扫描完成</h2>
        <div class="summary-stats">
          <div class="stat-item high">
            <span class="stat-number">{{ result.summary.high_risks }}</span>
            <span class="stat-label">高危</span>
          </div>
          <div class="stat-item medium">
            <span class="stat-number">{{ result.summary.medium_risks }}</span>
            <span class="stat-label">中危</span>
          </div>
          <div class="stat-item low">
            <span class="stat-number">{{ result.summary.low_risks }}</span>
            <span class="stat-label">低危</span>
          </div>
          <div class="stat-item suggestions">
            <span class="stat-number">{{ result.summary.total_suggestions }}</span>
            <span class="stat-label">建议</span>
          </div>
        </div>
      </div>

      <!-- 无风险 -->
      <div v-if="result.risks.length === 0" class="all-clear">
        <div class="all-clear-icon">✅</div>
        <p>未发现安全风险，系统状态良好！</p>
      </div>

      <!-- 风险列表 -->
      <div v-if="result.risks.length > 0" class="risks-section">
        <h3>风险详情</h3>
        <div class="risk-list">
          <div
            v-for="risk in result.risks"
            :key="risk.id"
            class="risk-item"
            :class="'risk-' + risk.level"
          >
            <div class="risk-header">
              <span class="risk-icon">{{ getRiskIcon(risk.level) }}</span>
              <span class="risk-title">{{ risk.title }}</span>
              <span class="risk-badge" :class="risk.level">{{ riskLevelText(risk.level) }}</span>
              <span v-if="risk.fixed" class="risk-fixed">已修复</span>
            </div>
            <p class="risk-description">{{ risk.description }}</p>
            <p v-if="risk.path" class="risk-path">📂 {{ risk.path }}</p>
            <p v-if="risk.fix_action" class="risk-fix">
              <strong>修复建议：</strong>{{ risk.fix_action }}
            </p>
            <button
              v-if="!risk.fixed && risk.fix_action"
              @click="fixRisk(risk)"
              class="btn btn-small"
            >
              查看修复方法
            </button>
          </div>
        </div>
      </div>

      <!-- 优化建议 -->
      <div v-if="result.suggestions.length > 0" class="suggestions-section">
        <h3>优化建议</h3>
        <div class="suggestion-list">
          <div
            v-for="sug in result.suggestions"
            :key="sug.id"
            class="suggestion-item"
          >
            <div class="suggestion-header">
              <h4>{{ sug.title }}</h4>
              <span class="suggestion-priority" :class="'priority-' + sug.priority">
                {{ getPriorityText(sug.priority) }}
              </span>
            </div>
            <p>{{ sug.description }}</p>
          </div>
        </div>
      </div>

      <div class="result-actions">
        <button @click="exportReport" class="btn btn-secondary">导出报告</button>
        <button @click="rescan" class="btn btn-primary">重新扫描</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import api from '../services/api'

const scanning = ref(false)
const result = ref(null)

// 进度相关
const progressStep  = ref(0)
const progressTotal = ref(9)
const progressPct   = ref(0)
const progressSteps = ref([
  '审计 MCP Server 配置',
  '检查 Claude 权限设置',
  '扫描 API Key 泄漏',
  '分析 Shell 历史记录',
  '扫描 .env 文件密钥',
  '检查 SSH 安全配置',
  '扫描对话历史 PII 数据',
  '审查敏感文件权限',
  '审查目录访问权限',
])

// 监听扫描进度事件
let removeProgressListener = null

onMounted(() => {
  if (window.electronAPI?.onAlert) {
    // 复用 onAlert 通道，但过滤 scan-progress 事件
    window.electronAPI.onScanProgress?.((data) => {
      progressStep.value = data.step
      progressTotal.value = data.total
      progressPct.value = Math.round((data.step / data.total) * 100)
    })
  }

  // 通过 guard-event 监听进度
  if (window.electronAPI?.onGuardEvent) {
    removeProgressListener = window.electronAPI.onGuardEvent((msg) => {
      if (msg.event === 'scan-progress') {
        progressStep.value = msg.data.step
        progressTotal.value = msg.data.total
        progressPct.value = Math.round((msg.data.step / msg.data.total) * 100)
      }
    })
  }
})

onUnmounted(() => {
  if (typeof removeProgressListener === 'function') removeProgressListener()
})

const startScan = async () => {
  scanning.value = true
  result.value = null
  progressStep.value = 0
  progressPct.value = 0

  try {
    const data = await api.scan()
    result.value = data.result || data
  } catch (error) {
    console.error('扫描失败:', error)
    alert('扫描失败: ' + error.message)
  } finally {
    scanning.value = false
    progressStep.value = progressTotal.value
    progressPct.value = 100
  }
}

const rescan = () => {
  result.value = null
  startScan()
}

const exportReport = async () => {
  try {
    const data = await api.exportReport()
    const blob = new Blob([data.report], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `security-report-${new Date().toISOString().slice(0, 10)}.txt`
    a.click()
    URL.revokeObjectURL(url)
  } catch (error) {
    console.error('导出失败:', error)
  }
}

const fixRisk = (risk) => {
  alert('修复方法：\n\n' + risk.fix_action)
}

const getSummaryClass = () => {
  if (!result.value) return ''
  if (result.value.summary.high_risks > 0) return 'summary-danger'
  if (result.value.summary.medium_risks > 0) return 'summary-warning'
  return 'summary-safe'
}

const getRiskIcon = (level) => ({ high: '🔴', medium: '🟡', low: '🟢' }[level] || '⚪')
const riskLevelText = (level) => ({ high: '高危', medium: '中危', low: '低危' }[level] || level)
const getPriorityText = (p) => ({ high: '高优先级', medium: '中优先级', low: '低优先级' }[p] || p)
</script>

<style scoped>
.scan-view {
  max-width: 860px;
  margin: 0 auto;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

.scan-header {
  text-align: center;
  margin-bottom: 40px;
}
.scan-header h1 { font-size: 32px; margin-bottom: 8px; }
.subtitle { color: #6b7280; font-size: 16px; }

/* ── 开始界面 ── */
.scan-start { text-align: center; padding: 20px; }
.scan-info {
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 16px;
  padding: 28px 36px;
  margin-bottom: 28px;
  text-align: left;
}
.scan-info h3 { margin-bottom: 18px; font-size: 16px; color: #d1d5db; }
.scan-info ul { list-style: none; padding: 0; }
.scan-info li {
  padding: 8px 0;
  font-size: 15px;
  color: #9ca3af;
  border-bottom: 1px solid rgba(255,255,255,0.05);
}
.scan-info li:last-child { border-bottom: none; }

/* ── 扫描中 ── */
.scan-progress { text-align: center; padding: 40px 20px; }
.spinner {
  width: 52px; height: 52px;
  margin: 0 auto 20px;
  border: 3px solid rgba(255,255,255,0.08);
  border-top-color: #6366f1;
  border-radius: 50%;
  animation: spin 0.9s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

.progress-title { font-size: 18px; margin-bottom: 20px; color: #e5e7eb; }

.progress-bar-wrap {
  width: 100%;
  max-width: 480px;
  height: 6px;
  background: rgba(255,255,255,0.08);
  border-radius: 99px;
  margin: 0 auto 8px;
  overflow: hidden;
}
.progress-bar {
  height: 100%;
  background: linear-gradient(90deg, #6366f1, #8b5cf6);
  border-radius: 99px;
  transition: width 0.4s ease;
}
.progress-pct { color: #6b7280; font-size: 13px; margin-bottom: 28px; }

.steps-list {
  display: inline-flex;
  flex-direction: column;
  gap: 10px;
  text-align: left;
}
.step-item {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  color: #4b5563;
  transition: color 0.3s;
}
.step-item.done  { color: #10b981; }
.step-item.active { color: #a5b4fc; font-weight: 600; }

.step-dot {
  width: 8px; height: 8px;
  border-radius: 50%;
  background: #374151;
  flex-shrink: 0;
  transition: background 0.3s;
}
.step-item.done  .step-dot { background: #10b981; }
.step-item.active .step-dot {
  background: #6366f1;
  box-shadow: 0 0 8px #6366f1;
  animation: pulse 1.2s infinite;
}
@keyframes pulse { 0%,100%{opacity:1} 50%{opacity:0.4} }

/* ── 结果 ── */
.result-summary {
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 16px;
  padding: 28px;
  margin-bottom: 28px;
  text-align: center;
}
.result-summary.summary-danger { border-color: rgba(239,68,68,0.4); background: rgba(239,68,68,0.05); }
.result-summary.summary-warning { border-color: rgba(245,158,11,0.4); background: rgba(245,158,11,0.05); }
.result-summary.summary-safe { border-color: rgba(16,185,129,0.4); background: rgba(16,185,129,0.05); }
.result-summary h2 { margin-bottom: 20px; font-size: 22px; }

.summary-stats { display: flex; justify-content: center; gap: 36px; }
.stat-item { text-align: center; }
.stat-number { display: block; font-size: 38px; font-weight: 700; }
.stat-label { font-size: 13px; color: #6b7280; }
.stat-item.high .stat-number   { color: #ef4444; }
.stat-item.medium .stat-number { color: #f59e0b; }
.stat-item.low .stat-number    { color: #10b981; }
.stat-item.suggestions .stat-number { color: #6366f1; }

/* All clear */
.all-clear { text-align: center; padding: 48px; color: #10b981; }
.all-clear-icon { font-size: 56px; margin-bottom: 12px; }
.all-clear p { font-size: 16px; }

/* Risks */
.risks-section, .suggestions-section { margin-bottom: 28px; }
.risks-section h3, .suggestions-section h3 { font-size: 18px; margin-bottom: 16px; color: #d1d5db; }

.risk-item {
  background: rgba(255,255,255,0.04);
  border-radius: 10px;
  padding: 16px 18px;
  margin-bottom: 10px;
  border-left: 4px solid;
}
.risk-item.risk-high   { border-left-color: #ef4444; }
.risk-item.risk-medium { border-left-color: #f59e0b; }
.risk-item.risk-low    { border-left-color: #10b981; }

.risk-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}
.risk-icon { font-size: 18px; }
.risk-title { font-weight: 600; flex: 1; color: #e5e7eb; }

.risk-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 99px;
  font-weight: 600;
}
.risk-badge.high   { background: rgba(239,68,68,0.2);  color: #fca5a5; }
.risk-badge.medium { background: rgba(245,158,11,0.2); color: #fcd34d; }
.risk-badge.low    { background: rgba(16,185,129,0.2); color: #6ee7b7; }

.risk-fixed {
  background: #10b981;
  color: white;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}
.risk-description { color: #9ca3af; margin-bottom: 6px; font-size: 14px; }
.risk-path { font-family: 'Monaco','Menlo',monospace; font-size: 12px; color: #6b7280; margin-bottom: 6px; }
.risk-fix { color: #9ca3af; font-size: 13px; background: rgba(255,255,255,0.03); padding: 8px 10px; border-radius: 6px; }

/* Suggestions */
.suggestion-item {
  background: rgba(99,102,241,0.07);
  border: 1px solid rgba(99,102,241,0.2);
  border-radius: 10px;
  padding: 16px 18px;
  margin-bottom: 10px;
}
.suggestion-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}
.suggestion-header h4 { margin: 0; color: #c7d2fe; font-size: 15px; }
.suggestion-item p { color: #9ca3af; font-size: 14px; margin: 0; }

.suggestion-priority {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 99px;
  font-weight: 600;
}
.priority-high   { background: rgba(239,68,68,0.2);  color: #fca5a5; }
.priority-medium { background: rgba(245,158,11,0.2); color: #fcd34d; }
.priority-low    { background: rgba(16,185,129,0.2); color: #6ee7b7; }

/* Actions */
.result-actions {
  display: flex;
  gap: 14px;
  justify-content: center;
  margin-top: 28px;
}

.btn {
  padding: 12px 32px;
  border: none;
  border-radius: 10px;
  font-size: 15px;
  cursor: pointer;
  transition: all 0.2s;
  font-weight: 500;
}
.btn-primary {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
}
.btn-primary:hover { filter: brightness(1.1); transform: translateY(-1px); }
.btn-secondary {
  background: rgba(255,255,255,0.08);
  color: #d1d5db;
  border: 1px solid rgba(255,255,255,0.1);
}
.btn-secondary:hover { background: rgba(255,255,255,0.12); }
.btn-small {
  padding: 6px 14px;
  font-size: 13px;
  margin-top: 10px;
  background: rgba(99,102,241,0.2);
  color: #a5b4fc;
  border: 1px solid rgba(99,102,241,0.3);
  border-radius: 6px;
}
.btn-small:hover { background: rgba(99,102,241,0.35); }
</style>
