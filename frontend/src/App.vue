<template>
  <div id="app">
    <Header :currentView="currentView" @navigate="navigate" />
    <main>
      <component :is="currentViewComponent" />
    </main>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import Header from './components/Header.vue'
import Overview from './views/Overview.vue'
import ScanView from './views/Scan.vue'
import MonitorView from './views/Monitor.vue'
import SettingsView from './views/Settings.vue'

const currentView = ref('overview')

const components = {
  overview: Overview,
  scan: ScanView,
  monitor: MonitorView,
  settings: SettingsView
}

const currentViewComponent = computed(() => components[currentView.value])

const navigate = (view) => {
  currentView.value = view
}

onMounted(() => {
  // 初始化 IPC 监听
  window.electronAPI?.onAlert((alert) => {
    console.log('收到告警:', alert)
  })
})
</script>

<style scoped>
main {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}
</style>
