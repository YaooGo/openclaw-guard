// API 服务 - 通过 IPC 与 Electron 主进程通信

class ApiService {
  constructor() {
    this.requestId = 0
  }

  async sendRequest(action, data = {}) {
    const id = `req_${++this.requestId}`

    // electronAPI.sendRequest 使用 ipcRenderer.invoke，直接返回 Promise，无需再注册事件监听
    const timeoutPromise = new Promise((_, reject) => {
      setTimeout(() => reject(new Error('请求超时')), 15000)
    })

    const requestPromise = window.electronAPI?.sendRequest({ id, action, data })

    try {
      const response = await Promise.race([requestPromise, timeoutPromise])
      return response
    } catch (err) {
      throw err
    }
  }

  // 扫描相关
  async scan() {
    return this.sendRequest('scan')
  }

  async exportReport() {
    return this.sendRequest('export_report')
  }

  // 监控相关
  async getMonitorStatus() {
    return this.sendRequest('get_monitor_status')
  }

  async setMonitorEnabled(enabled) {
    return this.sendRequest('set_monitor_enabled', { enabled })
  }

  async getLogs() {
    return this.sendRequest('get_logs')
  }

  async clearLogs() {
    return this.sendRequest('clear_logs')
  }

  // 配置相关
  async getConfig() {
    return this.sendRequest('get_config')
  }

  async updateConfig(config) {
    return this.sendRequest('update_config', config)
  }
}

export default new ApiService()
