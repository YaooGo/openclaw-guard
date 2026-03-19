const { contextBridge, ipcRenderer } = require('electron')

// 暴露安全的 API 给渲染进程
contextBridge.exposeInMainWorld('electronAPI', {
  // 发送请求到 Go 后端
  sendRequest: (request) => ipcRenderer.invoke('send-request', request),

  // 监听响应
  onResponse: (callback) => {
    ipcRenderer.on('guard-response', (event, response) => {
      callback(response)
    })
  },

  // 移除响应监听
  removeResponseHandler: (callback) => {
    ipcRenderer.removeListener('guard-response', callback)
  },

  // 监听告警事件
  onAlert: (callback) => {
    ipcRenderer.on('guard-event', (event, message) => {
      if (message.event === 'alert') {
        callback(message.data)
      }
    })
  },

  // 监听所有 guard 事件（包括 scan-progress）
  onGuardEvent: (callback) => {
    const handler = (event, message) => callback(message)
    ipcRenderer.on('guard-event', handler)
    // 返回移除函数
    return () => ipcRenderer.removeListener('guard-event', handler)
  },

  // 导出报告
  exportReport: (content) => ipcRenderer.invoke('export-report', content),

  // 平台信息
  platform: process.platform
})
