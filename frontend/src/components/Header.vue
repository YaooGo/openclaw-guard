<template>
  <header class="app-header">
    <div class="header-left">
      <h1 class="app-title">OpenClaw 安全卫士</h1>
      <span class="app-version">v1.0.0</span>
    </div>
    <nav class="header-nav">
      <button
        v-for="item in navItems"
        :key="item.id"
        @click="$emit('navigate', item.id)"
        class="nav-item"
        :class="{ active: currentView === item.id }"
      >
        {{ item.label }}
      </button>
    </nav>
  </header>
</template>

<script setup>
const props = defineProps({
  currentView: {
    type: String,
    default: 'overview'
  }
})

defineEmits(['navigate'])

const navItems = [
  { id: 'overview', label: '概览' },
  { id: 'scan', label: '安全扫描' },
  { id: 'monitor', label: '实时监控' },
  { id: 'settings', label: '设置' }
]
</script>

<style scoped>
.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  background: rgba(0, 0, 0, 0.3);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.header-left {
  display: flex;
  align-items: baseline;
  gap: 12px;
}

.app-title {
  font-size: 20px;
  font-weight: 600;
  margin: 0;
}

.app-version {
  font-size: 12px;
  color: #888;
}

.header-nav {
  display: flex;
  gap: 8px;
}

.nav-item {
  padding: 8px 16px;
  background: transparent;
  border: none;
  color: #aaa;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.2s;
  font-size: 14px;
}

.nav-item:hover {
  color: white;
  background: rgba(255, 255, 255, 0.05);
}

.nav-item.active {
  color: white;
  background: rgba(76, 175, 80, 0.2);
}
</style>
