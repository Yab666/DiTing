<div align="center">
  <img src="./internal/ui/web/assets/logo.png" width="200" />
  <br>
  <img src="https://img.shields.io/badge/Status-Stable-success?style=for-the-badge&logoColor=white" />
  <img src="https://img.shields.io/badge/Language-Go%201.20+-00ADD8?style=for-the-badge&logo=go&logoColor=white" />
  <img src="https://img.shields.io/badge/Architecture-Clean%20Architecture-blueviolet?style=for-the-badge" />
  <img src="https://img.shields.io/badge/AI-DeepSeek%20Integration-black?style=for-the-badge&logo=deepseek" />
  
  <h1>🛡️ DiTing (谛听)</h1>
  <h3>基于 Go 语言并发体系与大模型辅助研判的静态隐私排查引擎</h3>
</div>

---

## 📖 项目简介

DiTing（谛听）是一款用 Go 语言实现的静态代码扫描工具，旨在识别源码、配置文件及日志中硬编码的密钥、密码及敏感隐私信息。本项目在设计思路上参考了 Python 项目 `Whispers`，并通过以下技术手段提升了扫描性能与研判准确性：

1. **高性能并发引擎**：利用 Go 的 Goroutine 并发模型，实现多文件同步扫描。
2. **AI 辅助研判**：接入 DeepSeek 大模型接口，对疑似泄露点进行上下文语义分析，辅助降低误报率。
3. **多维解析插件**：内置 20 余种解析器，支持对结构化数据（YAML, JSON, XML）及多种编程语言源码的深度提取。

---

## ✨ 技术特性

*   **全场景解析能力**：支持包括 YAML、JSON、XML、Docker、Python、Go、Java、Shell、Pip、NPM 等在内的多种格式解析，能够精准定位到敏感信息所在的行号。
*   **双重噪声过滤机制**：
    *   **特征权重过滤**：基于变量名相似度评分（Similarity Filter）过滤常见的非敏感占位符。
    *   **上下文研判**：集成 AI Copilot，通过分析命中点周围的代码逻辑，判断泄露的真实性与危害性。
*   **灵活的交互模式**：
    *   **Web 可视化面板**：内置基于 Vue 3 的交互大屏，支持实时日志输出、风险图表展示及 AI 在线研判。
    *   **CLI 命令行工具**：支持集成至 CI/CD 流水线，输出标准化的 JSON 或 CSV 审计报告。
*   **轻量化分发**：单二进制文件运行，无需预装复杂依赖，支持跨平台快速部署。

---

## 📂 架构设计

本项目遵循 **Clean Architecture (整洁架构)** 模式，保持核心逻辑与外部接口的完全解耦：

```text
DiTing/
├── cmd/                # 程序入口
│   ├── ditting-cli/    # 命令行审计工具
│   └── ditting-web/    # 交互式 Web 服务
├── internal/           # 核心业务逻辑
│   ├── app/            # 任务调度与引擎组装
│   ├── core/           # 核心模型与领域定义
│   ├── plugin/         # 各类语言与格式解析插件
│   ├── rule/           # 规则匹配引擎与过滤逻辑
│   ├── scanner/        # 并行文件扫描器
│   └── ui/             # Web API 与前端静态资源嵌入
├── configs/rules/      # 敏感信息检测规则库 (YAML)
├── pkg/                # 通用基础设施组件
└── test_all/           # 自动化测试靶场
```

---

## 🏎️ 快速起航

### 1. 编译与运行 (Web 版)

适用于需要交互式查看扫描结果与 AI 研判的场景：

```bash
# 构建 Web 服务端
go build ./cmd/ditting-web

# 运行服务
./ditting-web
```
运行后程序将自动打开浏览器并定位至 `http://127.0.0.1:8080`。

### 2. 命令行使用 (CLI 版)

适用于脚本自动化或持续集成流水线：

```bash
# 构建命令行工具
go build ./cmd/ditting-cli

# 执行扫描并导出 JSON 报表
./ditting-cli -path "/your/project/path" -format json
```

---

## 📝 开发者说明

- **规则扩展**：可在 `configs/rules/` 目录下新增 YAML 规则。
- **解析器增加**：在 `internal/plugin/` 中实现 `Parser` 接口即可扩展支持新的文件格式。

---
**DiTing：构建更清晰、更准确的静态代码安全检测新标准。**
