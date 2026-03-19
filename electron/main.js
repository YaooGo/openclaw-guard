const { app, BrowserWindow, ipcMain, dialog } = require('electron')
const path = require('path')
const fs = require('fs')
const os = require('os')

let mainWindow = null
let fileWatchers = []  // 存储所有活跃的 watcher 实例
let monitorEnabled = false
let operationCount = { total: 0, blocked: 0, allowed: 0 }
let operationLogs = []
const MAX_LOGS = 100

// 白名单和黑名单路径
const whitelist = ['~/Projects', '~/code', '~/Documents', '~/Desktop']
const blacklist = ['~/.ssh', '~/.gnupg', '/etc', '/usr', '/System']

// 危险命令
const dangerousCommands = ['rm -rf', 'rm -r /', 'mkfs', 'dd if=', ':(){ :|:& };:', 'chmod 000']

// 创建主窗口
function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1000,
    height: 700,
    minWidth: 800,
    minHeight: 600,
    title: 'OpenClaw 安全卫士',
    backgroundColor: '#1a1a2e',
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false
    },
    titleBarStyle: 'hiddenInset'
  })

  if (process.env.NODE_ENV === 'development') {
    mainWindow.loadURL('http://localhost:5173')
  } else {
    mainWindow.loadFile(path.join(__dirname, 'dist/index.html'))
  }

  mainWindow.on('closed', () => {
    mainWindow = null
  })
}

// 展开 ~ 为用户目录
function expandPath(p) {
  if (p.startsWith('~/')) {
    return path.join(os.homedir(), p.slice(2))
  }
  return p
}

// 检查路径是否在黑名单中
function isBlacklisted(filePath) {
  const expanded = expandPath(filePath)
  for (const blocked of blacklist) {
    const expandedBlocked = expandPath(blocked)
    // 使用 startsWith + 分隔符检查，避免 /etc 误匹配 /etcfoo
    if (
      expanded === expandedBlocked ||
      expanded.startsWith(expandedBlocked + path.sep)
    ) {
      return true
    }
  }
  return false
}

// 检查路径是否在白名单中
function isWhitelisted(filePath) {
  const expanded = expandPath(filePath)
  for (const allowed of whitelist) {
    const expandedAllowed = expandPath(allowed)
    if (expanded.startsWith(expandedAllowed)) {
      return true
    }
  }
  return false
}

// 记录操作日志
function logOperation(type, filePath, command = null, allowed = true, reason = '') {
  const log = {
    id: `log_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
    type,
    path: filePath,
    command,
    allowed,
    reason,
    timestamp: Math.floor(Date.now() / 1000)
  }

  operationLogs.unshift(log)
  if (operationLogs.length > MAX_LOGS) {
    operationLogs = operationLogs.slice(0, MAX_LOGS)
  }

  operationCount.total++
  if (allowed) {
    operationCount.allowed++
  } else {
    operationCount.blocked++
  }

  // 发送告警到渲染进程
  if (mainWindow && !allowed) {
    mainWindow.webContents.send('guard-event', {
      event: 'alert',
      data: {
        type,
        severity: 'high',
        message: `${reason}: ${filePath}`,
        operation: log
      }
    })
  }
}

// 启动文件监控
function startFileMonitoring() {
  if (fileWatchers.length > 0) {
    stopFileMonitoring()
  }

  const homeDir = os.homedir()
  const watchDirs = [
    path.join(homeDir, '.claude'),
    path.join(homeDir, '.openclaw'),
    path.join(homeDir, 'Projects'),
    path.join(homeDir, 'code')
  ].filter(dir => {
    try {
      return fs.existsSync(dir)
    } catch {
      return false
    }
  })

  console.log('Starting file monitoring:', watchDirs)

  // 使用 fs.watch 监控文件变化
  const watched = new Set()

  watchDirs.forEach(dir => {
    try {
      // 递归监控目录
      const watcher = fs.watch(dir, { recursive: true }, (eventType, filename) => {
        if (!filename) return

        const fullPath = path.join(dir, filename)

        // 避免重复事件
        const key = `${eventType}:${fullPath}`
        if (watched.has(key)) return
        watched.add(key)
        setTimeout(() => watched.delete(key), 100)

        // 检查是否在黑名单
        if (isBlacklisted(fullPath)) {
          logOperation('access', fullPath, null, false, '黑名单路径')
          return
        }

        // 记录操作
        logOperation(eventType === 'rename' ? 'write' : 'read', fullPath, null, true, '正常访问')
      })

      watcher.on('error', (err) => {
        console.error('Watcher error:', err)
      })

      // 将 watcher 添加到全局数组，确保可以关闭
      fileWatchers.push(watcher)
    } catch (e) {
      console.error('Failed to watch:', dir, e)
    }
  })

  monitorEnabled = true
}

// 停止文件监控
function stopFileMonitoring() {
  for (const watcher of fileWatchers) {
    try {
      watcher.close()
    } catch (e) {}
  }
  fileWatchers = []
  monitorEnabled = false
}

// ─── 扫描辅助工具 ────────────────────────────────────────────────────────────

// 安全读取文件，失败返回 null
function safeRead(filePath) {
  try {
    if (fs.existsSync(filePath)) return fs.readFileSync(filePath, 'utf8')
  } catch (e) {}
  return null
}

// 安全 JSON 解析
function safeParseJSON(content) {
  try { return JSON.parse(content) } catch (e) { return null }
}

// 检查文件权限（仅 owner 可读写）
function checkFilePerms(filePath, expectedMode = 0o600) {
  try {
    if (!fs.existsSync(filePath)) return null
    const stats = fs.statSync(filePath)
    const mode = stats.mode & 0o777
    return (mode & 0o077) !== 0 ? mode : null
  } catch (e) { return null }
}

// ─── 9 大扫描模块 ─────────────────────────────────────────────────────────────

// 模块 1: MCP Server 配置审计
function scanMCPConfig(homeDir, risks, suggestions) {
  const mcpPaths = [
    path.join(homeDir, 'Library', 'Application Support', 'Claude', 'claude_desktop_config.json'),
    path.join(homeDir, '.claude', 'claude_desktop_config.json'),
    path.join(homeDir, '.config', 'claude', 'claude_desktop_config.json'),
  ]

  // 高危 MCP 命令特征
  const dangerousMcpPatterns = [
    { re: /ngrok|frp|cloudflare tunnel/i, msg: '使用了内网穿透工具，可能暴露本地服务' },
    { re: /curl\s+.*\|\s*(bash|sh)/i,     msg: '执行了远程脚本下载并运行的命令' },
    { re: /rm\s+-rf/i,                    msg: '包含危险的 rm -rf 命令' },
    { re: /sudo/i,                         msg: '以 sudo 提权运行' },
  ]

  // 高危权限
  const highRiskTools = ['computer_use', 'bash', 'filesystem', 'execute_code', 'terminal']

  for (const cfgPath of mcpPaths) {
    const content = safeRead(cfgPath)
    if (!content) continue

    const cfg = safeParseJSON(content)
    if (!cfg || !cfg.mcpServers) continue

    for (const [serverName, serverCfg] of Object.entries(cfg.mcpServers)) {
      // 检查远程 URL（非本地服务）
      const url = serverCfg.url || ''
      if (url && !/localhost|127\.0\.0\.1/.test(url)) {
        risks.push({
          id: `mcp_remote_${serverName}`,
          level: 'high',
          title: `MCP 服务 "${serverName}" 连接到远程服务器`,
          description: `服务地址: ${url}，使用远程 MCP 服务可能导致数据泄露`,
          path: cfgPath,
          fix_action: `确认 ${serverName} 是可信服务，或改用本地部署版本`
        })
      }

      // 检查命令中的危险模式
      const cmd = [serverCfg.command || '', ...(serverCfg.args || [])].join(' ')
      for (const { re, msg } of dangerousMcpPatterns) {
        if (re.test(cmd)) {
          risks.push({
            id: `mcp_danger_${serverName}`,
            level: 'high',
            title: `MCP 服务 "${serverName}" 包含危险操作`,
            description: `${msg}，命令: ${cmd.slice(0, 100)}`,
            path: cfgPath,
            fix_action: `检查并移除 MCP 服务器 "${serverName}" 的危险配置`
          })
          break
        }
      }

      // 检查高危工具授权
      const tools = serverCfg.allowedTools || serverCfg.tools || []
      for (const tool of tools) {
        if (highRiskTools.some(t => String(tool).toLowerCase().includes(t))) {
          risks.push({
            id: `mcp_tool_${serverName}_${tool}`,
            level: 'medium',
            title: `MCP 服务 "${serverName}" 授权了高危工具: ${tool}`,
            description: `工具 "${tool}" 可能允许 AI 执行任意系统命令`,
            path: cfgPath,
            fix_action: '在 MCP 配置中限制工具权限，仅授权必要的工具'
          })
        }
      }
    }

    suggestions.push({
      id: 'mcp_review',
      title: '定期审计 MCP 服务器配置',
      description: 'MCP 服务器可以访问本地资源，建议定期检查已授权的服务列表',
      priority: 'medium'
    })
  }
}

// 模块 2: Claude 权限设置审计
function scanClaudePermissions(homeDir, risks, suggestions) {
  const settingsPaths = [
    path.join(homeDir, '.claude', 'settings.json'),
    path.join(homeDir, '.claude', 'settings.local.json'),
  ]

  const highRiskPermissions = [
    'Bash',
    'computer_use',
    'WebFetch',
    'WebSearch',
  ]

  for (const settingsPath of settingsPaths) {
    const content = safeRead(settingsPath)
    if (!content) continue
    const cfg = safeParseJSON(content)
    if (!cfg) continue

    // 检查 allowedTools
    const allowedTools = cfg.allowedTools || cfg.permissions?.allowedTools || []
    for (const tool of allowedTools) {
      if (highRiskPermissions.some(r => String(tool).includes(r))) {
        risks.push({
          id: `claude_perm_${String(tool).replace(/[^a-z0-9]/gi, '_')}`,
          level: 'medium',
          title: `Claude 授权了高权限工具: ${tool}`,
          description: `settings.json 中 allowedTools 包含 "${tool}"，该工具可执行系统命令或访问网络`,
          path: settingsPath,
          fix_action: '仅在需要时临时授权高风险工具，或精确限定其作用域'
        })
      }
    }

    // 检查是否禁用了安全提示
    if (cfg.dangerouslySkipPermissions === true || cfg.skipPermissionPrompts === true) {
      risks.push({
        id: 'claude_skip_permissions',
        level: 'high',
        title: 'Claude 已禁用权限确认提示',
        description: 'dangerouslySkipPermissions=true 将让 AI 跳过所有操作确认，极度危险',
        path: settingsPath,
        fix_action: '将 dangerouslySkipPermissions 设置为 false 或删除该配置项'
      })
    }

    // 未设置 allowedTools（无限制）
    if (!allowedTools || allowedTools.length === 0) {
      suggestions.push({
        id: 'claude_no_allowlist',
        title: '建议为 Claude 配置工具白名单',
        description: '未设置 allowedTools 意味着 AI 可以请求任意工具权限，建议明确限制',
        priority: 'high'
      })
    }
  }
}

// 模块 3: 多平台 API Key 泄漏检测
function scanApiKeyLeaks(homeDir, risks) {
  const scanTargets = [
    path.join(homeDir, '.claude', 'config.json'),
    path.join(homeDir, '.claude', 'settings.json'),
    path.join(homeDir, '.claude', 'settings.local.json'),
    path.join(homeDir, '.openclaw', 'config.json'),
    path.join(homeDir, '.zshrc'),
    path.join(homeDir, '.bashrc'),
    path.join(homeDir, '.bash_profile'),
    path.join(homeDir, '.profile'),
    path.join(homeDir, '.npmrc'),
    path.join(homeDir, '.netrc'),
  ]

  // 各平台 API Key 特征
  const keyPatterns = [
    { re: /sk-ant-[a-zA-Z0-9\-_]{20,}/g,          label: 'Anthropic API Key (sk-ant-)' },
    { re: /sk-[a-zA-Z0-9]{48}/g,                   label: 'OpenAI API Key (sk-)' },
    { re: /AKIA[0-9A-Z]{16}/g,                     label: 'AWS Access Key ID' },
    { re: /[0-9a-zA-Z/+]{40}(?=[^a-zA-Z0-9/+]|$)/g, label: 'AWS Secret Key (40字符)' },
    { re: /ghp_[a-zA-Z0-9]{36}/g,                  label: 'GitHub Personal Access Token (ghp_)' },
    { re: /github_pat_[a-zA-Z0-9_]{82}/g,          label: 'GitHub Fine-grained Token' },
    { re: /sk_live_[a-zA-Z0-9]{24,}/g,             label: 'Stripe Live Secret Key' },
    { re: /sk_test_[a-zA-Z0-9]{24,}/g,             label: 'Stripe Test Secret Key' },
    { re: /xoxb-[0-9]{11}-[0-9]{11}-[a-zA-Z0-9]{24}/g, label: 'Slack Bot Token' },
    { re: /AIza[0-9A-Za-z\-_]{35}/g,               label: 'Google API Key (AIza)' },
    { re: /EAA[a-zA-Z0-9]{50,}/g,                  label: 'Facebook Access Token' },
  ]

  for (const target of scanTargets) {
    const content = safeRead(target)
    if (!content) continue

    const found = new Set()
    for (const { re, label } of keyPatterns) {
      re.lastIndex = 0
      const matches = content.match(re)
      if (matches && !found.has(label)) {
        found.add(label)
        risks.push({
          id: `apikey_${label.replace(/[^a-z0-9]/gi, '_')}_${path.basename(target)}`,
          level: 'high',
          title: `发现明文 ${label}`,
          description: `在 ${target} 中检测到 ${label}，共 ${matches.length} 处`,
          path: target,
          fix_action: '立即将密钥移至环境变量或密钥管理工具（如 1Password、macOS Keychain）'
        })
      }
    }
  }
}

// 模块 4: Shell 历史危险命令审计
function scanShellHistory(homeDir, risks) {
  const historyFiles = [
    path.join(homeDir, '.bash_history'),
    path.join(homeDir, '.zsh_history'),
    path.join(homeDir, '.sh_history'),
  ]

  const dangerPatterns = [
    { re: /rm\s+-rf\s+\/(?!\w)/,          level: 'high',   label: '删除根目录 (rm -rf /)' },
    { re: /rm\s+-rf\s+~\s*$/,             level: 'high',   label: '删除用户主目录 (rm -rf ~)' },
    { re: /:()\s*\{\s*:\|:\s*&\s*\};:/,   level: 'high',   label: 'Fork Bomb' },
    { re: /dd\s+if=\/dev\/(zero|random)/,  level: 'high',   label: '磁盘擦除命令 (dd)' },
    { re: /mkfs\./,                        level: 'high',   label: '格式化文件系统 (mkfs)' },
    { re: /chmod\s+-R\s+777\s+\//,        level: 'medium', label: '递归开放根目录权限' },
    { re: /curl.*\|\s*(bash|sh)/,         level: 'medium', label: '远程脚本直接执行 (curl|bash)' },
    { re: /wget.*\|\s*(bash|sh)/,         level: 'medium', label: '远程脚本直接执行 (wget|bash)' },
    { re: /base64\s+-d.*\|\s*(bash|sh)/,  level: 'medium', label: 'Base64 解码后直接执行' },
    { re: /python.*-c\s+["']import\s+os/, level: 'low',    label: 'Python 内联执行系统命令' },
  ]

  for (const histFile of historyFiles) {
    const content = safeRead(histFile)
    if (!content) continue

    const lines = content.split('\n')
    const reported = new Set()

    for (const line of lines) {
      const cmd = line.replace(/^:\s*\d+:\d+;/, '').trim() // 去掉 zsh history 时间戳
      for (const { re, level, label } of dangerPatterns) {
        if (re.test(cmd) && !reported.has(label)) {
          reported.add(label)
          risks.push({
            id: `history_${label.replace(/[^a-z0-9]/gi, '_')}`,
            level,
            title: `Shell 历史中发现危险命令: ${label}`,
            description: `在 ${path.basename(histFile)} 中检测到"${label}"执行记录`,
            path: histFile,
            fix_action: '确认该命令是否为 AI 助手生成并执行，建议定期清理 Shell 历史'
          })
        }
      }
    }
  }
}

// 模块 5: .env 文件密钥泄漏
function scanEnvFiles(homeDir, risks) {
  // 扫描常见项目目录下的 .env 文件
  const searchDirs = [
    path.join(homeDir, 'Projects'),
    path.join(homeDir, 'code'),
    path.join(homeDir, 'Desktop'),
    path.join(homeDir, 'Documents'),
  ].filter(d => { try { return fs.existsSync(d) } catch(e) { return false } })

  const secretPatterns = [
    /(?:SECRET|KEY|TOKEN|PASSWORD|PASSWD|PWD|API_KEY|PRIVATE)\s*=\s*[^\s#]{8,}/i,
    /sk-ant-[a-zA-Z0-9\-_]{20,}/,
    /sk-[a-zA-Z0-9]{48}/,
    /AKIA[0-9A-Z]{16}/,
    /ghp_[a-zA-Z0-9]{36}/,
  ]

  let envFilesScanned = 0

  function scanDir(dir, depth = 0) {
    if (depth > 2 || envFilesScanned > 50) return
    try {
      const entries = fs.readdirSync(dir, { withFileTypes: true })
      for (const entry of entries) {
        if (entry.name.startsWith('.') && entry.name !== '.env' && entry.name !== '.env.local') continue
        const fullPath = path.join(dir, entry.name)
        if (entry.isDirectory() && depth < 2 && entry.name !== 'node_modules' && entry.name !== '.git') {
          scanDir(fullPath, depth + 1)
        } else if (entry.isFile() && /^\.env(\.\w+)?$/.test(entry.name)) {
          const content = safeRead(fullPath)
          if (!content) continue
          envFilesScanned++
          for (const re of secretPatterns) {
            if (re.test(content)) {
              risks.push({
                id: `env_secret_${fullPath.replace(/[^a-z0-9]/gi, '_')}`,
                level: 'high',
                title: `.env 文件包含明文密钥`,
                description: `在 ${fullPath} 中发现可能的 API Key 或密码`,
                path: fullPath,
                fix_action: '将密钥移至系统 Keychain，并在 .gitignore 中添加 .env'
              })
              break
            }
          }
        }
      }
    } catch (e) {}
  }

  for (const dir of searchDirs) {
    scanDir(dir, 0)
  }
}

// 模块 6: SSH known_hosts 安全配置检查
function scanSSHConfig(homeDir, risks, suggestions) {
  const sshConfig = safeRead(path.join(homeDir, '.ssh', 'config'))

  if (sshConfig) {
    if (!sshConfig.includes('HashKnownHosts yes')) {
      suggestions.push({
        id: 'ssh_hash_known_hosts',
        title: '建议开启 SSH HashKnownHosts',
        description: '在 ~/.ssh/config 添加 "HashKnownHosts yes" 可防止历史连接记录泄露服务器地址',
        priority: 'low'
      })
    }
    if (/StrictHostKeyChecking\s+no/i.test(sshConfig)) {
      risks.push({
        id: 'ssh_strict_host_no',
        level: 'high',
        title: 'SSH 已禁用主机密钥校验',
        description: 'StrictHostKeyChecking=no 会使 SSH 连接易受中间人攻击',
        path: path.join(homeDir, '.ssh', 'config'),
        fix_action: '将 StrictHostKeyChecking 设置为 yes 或 ask'
      })
    }
  }

  // 所有 SSH 私钥文件
  const sshDir = path.join(homeDir, '.ssh')
  try {
    if (fs.existsSync(sshDir)) {
      const files = fs.readdirSync(sshDir)
      for (const file of files) {
        if (/\.(pub|known_hosts|config|authorized_keys)$/.test(file)) continue
        const fullPath = path.join(sshDir, file)
        const badMode = checkFilePerms(fullPath, 0o600)
        if (badMode !== null) {
          risks.push({
            id: `perm_ssh_${file}`,
            level: 'high',
            title: `SSH 私钥文件权限过宽: ${file}`,
            description: `${fullPath} 权限为 ${(badMode).toString(8)}，应为 600`,
            path: fullPath,
            fix_action: `运行: chmod 600 "${fullPath}"`
          })
        }
      }
    }
  } catch (e) {}
}

// 模块 7: Claude 对话历史 PII 扫描
function scanClaudeHistoryPII(homeDir, risks) {
  const historyPaths = [
    path.join(homeDir, '.claude', 'history.json'),
    path.join(homeDir, '.openclaw', 'logs'),
  ]

  // PII 特征
  const piiPatterns = [
    { re: /1[3-9]\d{9}/g,                                        label: '手机号' },
    { re: /[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])\d{2}\d{3}[\dXx]/g, label: '身份证号' },
    { re: /\b(?:25[0-5]|2[0-4]\d|[01]?\d\d?)(?:\.(?:25[0-5]|2[0-4]\d|[01]?\d\d?)){3}\b/g, label: 'IP 地址' },
    { re: /[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}/g,  label: '邮箱地址' },
    { re: /password\s*[:=]\s*\S{4,}/gi,                           label: '明文密码字段' },
    { re: /(?:SECRET|PRIVATE_KEY|API_KEY)\s*[:=]\s*\S{8,}/gi,     label: '密钥字段' },
  ]

  for (const histPath of historyPaths) {
    const content = safeRead(histPath)
    if (!content) continue

    const found = []
    for (const { re, label } of piiPatterns) {
      re.lastIndex = 0
      const matches = content.match(re)
      if (matches) found.push(`${label}(${matches.length}处)`)
    }

    if (found.length > 0) {
      risks.push({
        id: 'claude_history_pii',
        level: 'medium',
        title: 'Claude 历史记录包含 PII 数据',
        description: `在 ${path.basename(histPath)} 中发现: ${found.join('、')}`,
        path: histPath,
        fix_action: '清理历史记录，并避免在与 AI 的对话中粘贴敏感个人信息'
      })
    }
  }
}

// 模块 8: 敏感文件权限检查（扩展版）
function scanSensitiveFilePerms(homeDir, risks) {
  const checks = [
    { file: path.join(homeDir, '.gnupg', 'secring.gpg'),  mode: 0o600, label: 'GPG 密钥环' },
    { file: path.join(homeDir, '.netrc'),                  mode: 0o600, label: '.netrc 认证文件' },
    { file: path.join(homeDir, '.npmrc'),                  mode: 0o600, label: '.npmrc（可能含 token）' },
    { file: path.join(homeDir, '.pypirc'),                 mode: 0o600, label: '.pypirc（PyPI token）' },
    { file: path.join(homeDir, '.aws', 'credentials'),    mode: 0o600, label: 'AWS 凭证文件' },
    { file: path.join(homeDir, '.aws', 'config'),         mode: 0o600, label: 'AWS 配置文件' },
  ]

  for (const { file, mode, label } of checks) {
    const badMode = checkFilePerms(file, mode)
    if (badMode !== null) {
      risks.push({
        id: `perm_${path.basename(file)}`,
        level: 'high',
        title: `${label} 权限过于宽松`,
        description: `${file} 权限为 ${(badMode).toString(8)}，应为 ${mode.toString(8)}`,
        path: file,
        fix_action: `运行: chmod ${mode.toString(8)} "${file}"`
      })
    }
  }
}

// 模块 9: 目录权限检查
function scanDirPerms(homeDir, risks) {
  const dirChecks = [
    { dir: path.join(homeDir, '.claude'),  mode: 0o700, label: 'Claude 配置目录' },
    { dir: path.join(homeDir, '.ssh'),     mode: 0o700, label: 'SSH 目录' },
    { dir: path.join(homeDir, '.gnupg'),   mode: 0o700, label: 'GPG 目录' },
    { dir: path.join(homeDir, '.aws'),     mode: 0o700, label: 'AWS 配置目录' },
  ]

  for (const { dir, mode, label } of dirChecks) {
    try {
      if (!fs.existsSync(dir)) continue
      const stats = fs.statSync(dir)
      const actual = stats.mode & 0o777
      if ((actual & 0o077) !== 0) {
        risks.push({
          id: `perm_dir_${path.basename(dir)}`,
          level: 'medium',
          title: `${label} 权限过于宽松`,
          description: `${dir} 权限为 ${actual.toString(8)}，应为 700`,
          path: dir,
          fix_action: `运行: chmod 700 "${dir}"`
        })
      }
    } catch (e) {}
  }
}

// ─── 主扫描函数 ───────────────────────────────────────────────────────────────

async function performScan(onProgress) {
  const homeDir = os.homedir()
  const risks = []
  const suggestions = []

  const steps = [
    { label: '正在审计 MCP Server 配置...', fn: () => scanMCPConfig(homeDir, risks, suggestions) },
    { label: '正在检查 Claude 权限设置...',  fn: () => scanClaudePermissions(homeDir, risks, suggestions) },
    { label: '正在扫描 API Key 泄漏...',     fn: () => scanApiKeyLeaks(homeDir, risks) },
    { label: '正在分析 Shell 历史记录...',   fn: () => scanShellHistory(homeDir, risks) },
    { label: '正在扫描 .env 文件密钥...',    fn: () => scanEnvFiles(homeDir, risks) },
    { label: '正在检查 SSH 安全配置...',     fn: () => scanSSHConfig(homeDir, risks, suggestions) },
    { label: '正在扫描对话历史 PII 数据...', fn: () => scanClaudeHistoryPII(homeDir, risks) },
    { label: '正在审查敏感文件权限...',       fn: () => scanSensitiveFilePerms(homeDir, risks) },
    { label: '正在审查目录访问权限...',       fn: () => scanDirPerms(homeDir, risks) },
  ]

  for (let i = 0; i < steps.length; i++) {
    const { label, fn } = steps[i]
    if (onProgress) onProgress(label, i + 1, steps.length)
    try { fn() } catch (e) { console.error('Scan module error:', e) }
  }

  // 去重（相同 id 保留第一个）
  const seen = new Set()
  const deduped = risks.filter(r => {
    if (seen.has(r.id)) return false
    seen.add(r.id)
    return true
  })

  // 按级别排序：high → medium → low
  const order = { high: 0, medium: 1, low: 2 }
  deduped.sort((a, b) => (order[a.level] ?? 3) - (order[b.level] ?? 3))

  const dedupedSuggs = [...new Map(suggestions.map(s => [s.id, s])).values()]

  return {
    summary: {
      high_risks:        deduped.filter(r => r.level === 'high').length,
      medium_risks:      deduped.filter(r => r.level === 'medium').length,
      low_risks:         deduped.filter(r => r.level === 'low').length,
      total_risks:       deduped.length,
      total_suggestions: dedupedSuggs.length,
    },
    risks:       deduped,
    suggestions: dedupedSuggs,
  }
}


// 清空日志
function clearLogs() {
  operationLogs = []
  operationCount = { total: 0, blocked: 0, allowed: 0 }
}

// IPC 处理
ipcMain.handle('send-request', async (event, request) => {
  console.log('Received request:', request.action)

  // 扫描
  if (request.action === 'scan') {
    try {
      const result = await performScan((label, step, total) => {
        if (mainWindow && !mainWindow.isDestroyed()) {
          mainWindow.webContents.send('guard-event', {
            event: 'scan-progress',
            data: { label, step, total }
          })
        }
      })
      return { result }
    } catch (e) {
      console.error('Scan error:', e)
      return { result: { summary: { high_risks: 0, medium_risks: 0, low_risks: 0, total_risks: 0, total_suggestions: 0 }, risks: [], suggestions: [] } }
    }
  }

  // 监控状态
  if (request.action === 'get_monitor_status') {
    return {
      status: {
        enabled: monitorEnabled,
        uptime: monitorEnabled ? Math.floor(process.uptime()) : 0,
        total_operations: operationCount.total,
        blocked_count: operationCount.blocked,
        allowed_count: operationCount.allowed
      }
    }
  }

  // 开关监控
  if (request.action === 'set_monitor_enabled') {
    const enabled = request.data?.enabled
    if (enabled) {
      startFileMonitoring()
    } else {
      stopFileMonitoring()
    }
    return { success: true, enabled }
  }

  // 日志
  if (request.action === 'get_logs') {
    return { logs: operationLogs }
  }

  // 清空日志
  if (request.action === 'clear_logs') {
    clearLogs()
    return { success: true }
  }

  // 配置
  if (request.action === 'get_config') {
    return {
      whitelist_paths: whitelist,
      blacklist_paths: blacklist,
      dangerous_commands: dangerousCommands,
      monitor_enabled: monitorEnabled,
      scan_on_start: true,
      auto_fix: false
    }
  }

  // 更新配置
  if (request.action === 'update_config') {
    if (request.data?.whitelist_paths) {
      whitelist.length = 0
      whitelist.push(...request.data.whitelist_paths)
    }
    if (request.data?.blacklist_paths) {
      blacklist.length = 0
      blacklist.push(...request.data.blacklist_paths)
    }
    return { success: true }
  }

  // 导出报告
  if (request.action === 'export_report') {
    return { report: 'Security Report Content' }
  }

  return {}
})

// 应用生命周期
app.whenReady().then(() => {
  createWindow()

  // 启动时自动开启监控
  startFileMonitoring()

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow()
    }
  })
})

app.on('window-all-closed', () => {
  stopFileMonitoring()
  if (process.platform !== 'darwin') {
    app.quit()
  }
})

app.on('will-quit', () => {
  stopFileMonitoring()
})
