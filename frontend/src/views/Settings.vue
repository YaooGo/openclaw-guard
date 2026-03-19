<template>
  <div class="settings-view">
    <div class="settings-header">
      <h1>⚙️ 设置</h1>
    </div>

    <div class="settings-section">
      <h3>基本设置</h3>
      <div class="setting-item">
        <label>
          <input type="checkbox" v-model="config.scan_on_start" @change="saveConfig">
          启动时自动扫描
        </label>
      </div>
      <div class="setting-item">
        <label>
          <input type="checkbox" v-model="config.monitor_enabled" @change="saveConfig">
          启用实时监控
        </label>
      </div>
      <div class="setting-item">
        <label>
          <input type="checkbox" v-model="config.auto_fix" @change="saveConfig">
          自动修复安全问题（开发中）
        </label>
      </div>
    </div>

    <div class="settings-section">
      <h3>路径白名单</h3>
      <p class="section-desc">这些目录的操作将被允许</p>
      <div class="path-list">
        <div v-for="(path, index) in config.whitelist_paths" :key="index" class="path-item">
          <span class="path-text">{{ path }}</span>
          <button @click="removeWhitelistPath(index)" class="btn-remove">×</button>
        </div>
      </div>
      <div class="path-add">
        <input
          v-model="newWhitelistPath"
          @keyup.enter="addWhitelistPath"
          placeholder="输入路径，如 ~/Projects"
          class="path-input"
        >
        <button @click="addWhitelistPath" class="btn-add">添加</button>
      </div>
    </div>

    <div class="settings-section">
      <h3>路径黑名单</h3>
      <p class="section-desc">这些目录的操作将被拦截</p>
      <div class="path-list">
        <div v-for="(path, index) in config.blacklist_paths" :key="index" class="path-item danger">
          <span class="path-text">{{ path }}</span>
          <button @click="removeBlacklistPath(index)" class="btn-remove">×</button>
        </div>
      </div>
      <div class="path-add">
        <input
          v-model="newBlacklistPath"
          @keyup.enter="addBlacklistPath"
          placeholder="输入路径，如 ~/.ssh"
          class="path-input"
        >
        <button @click="addBlacklistPath" class="btn-add">添加</button>
      </div>
    </div>

    <div class="settings-section">
      <h3>危险命令检测</h3>
      <p class="section-desc">检测这些命令时会发出警告</p>
      <div class="path-list">
        <div v-for="(cmd, index) in config.dangerous_commands" :key="index" class="path-item danger">
          <span class="path-text">{{ cmd }}</span>
          <button @click="removeDangerousCommand(index)" class="btn-remove">×</button>
        </div>
      </div>
      <div class="path-add">
        <input
          v-model="newDangerousCommand"
          @keyup.enter="addDangerousCommand"
          placeholder="输入危险命令，如 rm -rf"
          class="path-input"
        >
        <button @click="addDangerousCommand" class="btn-add">添加</button>
      </div>
    </div>

    <div class="settings-actions">
      <button @click="resetConfig" class="btn btn-secondary">重置默认</button>
      <button @click="saveConfig" class="btn btn-primary">保存设置</button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../services/api'

const config = ref({
  scan_on_start: true,
  monitor_enabled: true,
  auto_fix: false,
  whitelist_paths: [],
  blacklist_paths: [],
  dangerous_commands: []
})

const newWhitelistPath = ref('')
const newBlacklistPath = ref('')
const newDangerousCommand = ref('')

const loadConfig = async () => {
  try {
    const data = await api.getConfig()
    config.value = { ...config.value, ...data }
  } catch (error) {
    console.error('加载配置失败:', error)
  }
}

const saveConfig = async () => {
  try {
    await api.updateConfig(config.value)
    alert('设置已保存')
  } catch (error) {
    console.error('保存配置失败:', error)
    alert('保存失败: ' + error.message)
  }
}

const addWhitelistPath = () => {
  if (newWhitelistPath.value.trim()) {
    config.value.whitelist_paths.push(newWhitelistPath.value.trim())
    newWhitelistPath.value = ''
  }
}

const removeWhitelistPath = (index) => {
  config.value.whitelist_paths.splice(index, 1)
}

const addBlacklistPath = () => {
  if (newBlacklistPath.value.trim()) {
    config.value.blacklist_paths.push(newBlacklistPath.value.trim())
    newBlacklistPath.value = ''
  }
}

const removeBlacklistPath = (index) => {
  config.value.blacklist_paths.splice(index, 1)
}

const addDangerousCommand = () => {
  if (newDangerousCommand.value.trim()) {
    config.value.dangerous_commands.push(newDangerousCommand.value.trim())
    newDangerousCommand.value = ''
  }
}

const removeDangerousCommand = (index) => {
  config.value.dangerous_commands.splice(index, 1)
}

const resetConfig = () => {
  if (confirm('确定要重置为默认设置吗？')) {
    config.value = {
      scan_on_start: true,
      monitor_enabled: true,
      auto_fix: false,
      whitelist_paths: ['~/Projects', '~/code', '~/Documents', '~/Desktop'],
      blacklist_paths: ['~/.ssh', '~/.gnupg', '/etc', '/usr', '/System'],
      dangerous_commands: ['rm -rf', 'rm -r /', 'mkfs', 'dd if=', 'format']
    }
  }
}

onMounted(() => {
  loadConfig()
})
</script>

<style scoped>
.settings-view {
  max-width: 800px;
  margin: 0 auto;
}

.settings-header {
  margin-bottom: 30px;
}

.settings-header h1 {
  font-size: 28px;
}

.settings-section {
  background: rgba(255, 255, 255, 0.03);
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 20px;
}

.settings-section h3 {
  font-size: 18px;
  margin-bottom: 16px;
}

.section-desc {
  color: #888;
  font-size: 14px;
  margin-bottom: 16px;
}

.setting-item {
  margin-bottom: 12px;
}

.setting-item label {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
}

.setting-item input[type="checkbox"] {
  width: 18px;
  height: 18px;
}

.path-list {
  margin-bottom: 16px;
}

.path-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 6px;
  padding: 10px 14px;
  margin-bottom: 8px;
}

.path-item.danger {
  border-left: 3px solid #f44336;
}

.path-text {
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 14px;
}

.btn-remove {
  width: 24px;
  height: 24px;
  border: none;
  background: rgba(244, 63, 54, 0.2);
  color: #f44336;
  border-radius: 4px;
  cursor: pointer;
  font-size: 18px;
  line-height: 1;
}

.btn-remove:hover {
  background: rgba(244, 63, 54, 0.4);
}

.path-add {
  display: flex;
  gap: 10px;
}

.path-input {
  flex: 1;
  padding: 10px 14px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  color: white;
  font-size: 14px;
}

.path-input:focus {
  outline: none;
  border-color: #4CAF50;
}

.btn-add {
  padding: 10px 20px;
  background: #4CAF50;
  border: none;
  border-radius: 6px;
  color: white;
  cursor: pointer;
}

.btn-add:hover {
  background: #45a049;
}

.settings-actions {
  display: flex;
  gap: 15px;
  justify-content: center;
  margin-top: 30px;
}

.btn {
  padding: 12px 30px;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  cursor: pointer;
}

.btn-primary {
  background: #4CAF50;
  color: white;
}

.btn-primary:hover {
  background: #45a049;
}

.btn-secondary {
  background: rgba(255, 255, 255, 0.1);
  color: white;
}

.btn-secondary:hover {
  background: rgba(255, 255, 255, 0.2);
}
</style>
