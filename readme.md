# mHost - macOS Host 文件管理工具

一个专为 macOS 设计的 host 文件管理工具，支持多 profile 配置和一键切换。

## 功能特性

- 🔧 **多 Profile 管理** - 创建和管理多个 host 配置文件
- 🚀 **一键切换** - 快速在不同 host 配置间切换
- 🛡️ **安全可靠** - 自动备份，失败回滚
- 🎨 **现代界面** - 基于 Fyne 的原生 macOS 界面
- ✅ **格式验证** - 智能验证 IP 地址和域名格式
- 📝 **操作日志** - 完整的操作历史记录

## 技术栈

- **后端**: Go 语言
- **UI 框架**: Fyne
- **目标平台**: macOS 10.14+
- **架构支持**: Intel x64 & Apple Silicon

## 系统要求

- macOS 10.14 或更高版本
- 管理员权限（修改 /etc/hosts 文件）

## 快速开始

### 安装依赖

```bash
# 安装 Go (如果尚未安装)
brew install go

# 安装 Fyne 依赖
go mod tidy
```

### 构建应用

```bash
# 构建可执行文件
go build -o mHost ./cmd/mhost

# 或者直接运行
go run ./cmd/mhost
```

### 打包为 macOS 应用

```bash
# 安装 fyne 打包工具
go install fyne.io/fyne/v2/cmd/fyne@latest

# 打包为 .app 文件
fyne package -os darwin
```

## 项目结构

```
mHost/
├── cmd/
│   └── mhost/          # 主程序入口
├── internal/
│   ├── core/           # 核心业务逻辑
│   ├── ui/             # Fyne UI 组件
│   ├── config/         # 配置管理
│   └── utils/          # 工具函数
├── pkg/
│   └── hosts/          # Host 文件操作库
├── assets/             # 资源文件
├── doc/                # 文档
└── README.md
```

## 使用说明

1. **创建 Profile**: 点击「新建 Profile」按钮，输入名称和描述
2. **编辑 Host 条目**: 在右侧编辑面板添加或修改 host 条目
3. **切换 Profile**: 选择左侧 Profile 列表中的项目，点击「应用」
4. **备份恢复**: 应用会自动备份原始 hosts 文件，支持一键恢复

## 开发计划

详细的开发计划和功能需求请参考 [需求文档](./doc/requirements.md)。

### 开发阶段

- [x] 需求分析和文档编写
- [ ] 核心功能开发（Go 后端）
- [ ] UI 界面实现（Fyne）
- [ ] 权限管理和安全性
- [ ] 测试和优化
- [ ] 打包和发布

## 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 本项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 联系方式

如有问题或建议，请通过 Issue 联系我们。

---

**注意**: 本工具需要管理员权限来修改系统 hosts 文件，请确保从可信来源下载和使用。