---
description: how to release a new version to GitHub
---

## 标准 Git 工作流

### 日常开发 - 功能分支

1. 从 main 创建新分支
```bash
git checkout main && git pull
git checkout -b feat/your-feature-name
```

2. 开发并提交
```bash
git add .
git commit -m "feat: 你的功能描述"
```

3. 推送分支，发起 PR
```bash
git push origin feat/your-feature-name
# 然后去 GitHub 上开一个 PR → main
```

4. PR 合并后，同步本地 main
```bash
git checkout main && git pull
git branch -d feat/your-feature-name
```

---

### 发布新版本 - 打 Tag 触发自动构建

5. 在 main 上更新版本号
```bash
git checkout main && git pull
# 修改 VERSION 文件和 electron/package.json 中的 version 字段
echo "1.0.1" > VERSION
git add . && git commit -m "chore: bump version to v1.0.1"
git push
```

6. 打 Tag 推送，触发自动发布
// turbo
```bash
git tag v1.0.1
git push origin v1.0.1
```

等约 5 分钟，GitHub Actions 自动完成：
- 编译 Go 后端 + 打包前端 + 生成 .dmg
- 自动创建 Release 并上传安装包
