<div align="center">
  <img src="https://img.shields.io/badge/Status-Completed-success?style=for-the-badge&logoColor=white" />
  <img src="https://img.shields.io/badge/Language-Go%201.20+-00ADD8?style=for-the-badge&logo=go&logoColor=white" />
  <img src="https://img.shields.io/badge/Architecture-Clean%20Architecture-blueviolet?style=for-the-badge" />
  <img src="https://img.shields.io/badge/AI-DeepSeek%20Copilot-black?style=for-the-badge&logo=deepseek" />
  
  <h1>🛡️ DiTing (谛听)</h1>
  <h3>基于并发范式与 LLM 降噪的下一代静态隐私排查引擎</h3>
</div>

---

DiTing 是对传统 Python 慢速扫描流（原项目 `Whispers`）进行的一次**降维打击式重构**。
它利用 Go 语言强劲的 Goroutine 并发体系，将文件扫描执行效率提升了数十倍。在架构层面，DiTing 剥离了落后的脚本配置模式，实现了单体二进制 Web 面板嵌入，并原生首创性地接入了大模型（DeepSeek）语义二次研判诊断，一举破解了静态代码扫描“误报率极高”的行业痛点。

---

## ✨ 核心特性大盘点 (Features)

1. 🚀 **引擎极速提权 (Go Concurrency)**：废除单核死锁，支持千万级工程树形穿透快扫，零外部依赖、极低内存占用。
2. 🧠 **全语种语义解构引擎 (Dispatcher)**：内置高达 19 种语言级别的解析插件（涵盖 YAML, JSON, XML, Java, Go, Python, Nginx, Dockerfile 等），精确到具体行号与语义块。
3. 🛠️ **多维干扰降噪防线 (Anti-False Positives)**：
   - 特征过滤：通过动态变量名拦截与 Similarity Filter（相似片段打分技术）隔离常规变量。
   - 结构化盲区：特创 Breadcrumb 基于配置层次路径 (`config.db.password`) 的静默豁免。
4. 🤖 **「神之复核」LLM Copilot**：内置 DeepSeek 大模型驱动的 AI 对话审核面板，自动组装可疑断点进行上下文智能研判，彻底杀灭 False Positive (误报)。
5. 💻 **SaaS 级开箱即用套件**：通过 `//go:embed` 超级特性将前端带有 `Glassmorphism`（深色毛玻璃）炫酷动效的交互大屏塞进了唯一的 EXE 中。终端与图形化随心切换。

---

## 📂 项目模块全景 (Folder Structure)

本项目坚持严格的 **Clean Architecture (整洁架构)** 领域驱动设计，各个模块严丝合缝、极致解耦。

```text
DiTing/
├── cmd/                # 入口舱 (程序的触发点)
│   ├── ditting-cli/    # [构建出口] 提供极简 CLI 终端与自动化报表导出的黑框程序。
│   └── ditting-web/    # [构建出口] 提供内嵌了 Web GUI 可视化交互的高级运行体。
│
├── configs/rules/      # 规则库 (弹药库：所有的密钥、指纹规则如 gitkeys.yaml 均存放于此，支持热插拔)
├── docs/               # 文档归档区
│
├── internal/           # 核心业务引擎 (严禁外部业务项目引入)
│   ├── app/            # 顶层调度器：组装规则、挂载插件并发起冲锋任务。
│   ├── core/           # 领域层基架：定义 Secret、Rule 结构体和核心数据模型。
│   ├── plugin/         # 解析插件栈：19大文件格式清洗器汇聚于此。
│   ├── rule/           # 正则拦截网：包含基于 `regexp2` 的高阶捕获模型与精准阻断过滤逻辑。
│   ├── scanner/        # 并行疾跑腿：负责高速、安全、低损耗地递归遍历系统硬盘并加载代码进内存。
│   └── ui/             # 大屏终端站：挂载 HTTP 核心网关与纯原生封装的 Vue3 Web 仪表盘。
│
├── pkg/                # 基础设施公共组件 (任何人都能复用的轮子)
│   ├── config/         # 处理 .ditingrc 文件的解析与默认策略兜底。
│   └── logger/         # 标准化日志流基座。
│
└── test_all/           # 测试靶场 (包含大量硬编码、假秘钥以供引擎测漏验证)
```

---

## 🏎️ 快速起航 (How to Run)

DiTing 的部署哲学是**零依赖**。您无需预装任何庞大的 NodeJS 或 Python 依赖环境，拥有 Go 环境即可一键起飞。

### 方式 1: 启动大屏可视化 Web 版 (强烈推荐)

专为想要追求极致沉浸体验的极客与开源赛道评委打造：

```bash
# 进入工程目录
cd DiTing

# 将各种界面组件直接烙印为 exe 单程序
go build ./cmd/ditting-web

# 运行它！
.\ditting-web.exe
```
> **💡 AI 体验说明**：在运行后，控制台会自动切出浏览器。选择需要扫描的目录（例如 `E:/我的项目/whispers/DiTing/test_all`）后开始排雷。
> **要想体验本项目的灵魂特性**：在上方的输入框填入您的 DeepSeek `API-Key`，点击单条漏洞的 `[AI 复核]` 按钮即可感受 LLM 动态审判全过程！

<br>

### 方式 2: 使用极速终端模式 (适用于 CI/CD 自动构建流水线)

如果您试图将 DiTing 加入到 GitHub Actions 或者 Jenkins 节点，并需要一份机器可读的静态报表报告。

```bash
# 仅仅挂载 CLI 引擎进行编译
go build ./cmd/ditting-cli

# 命令格式:
# .\ditting-cli.exe -path [挂载的工程绝对路径] -format [生成类型 json 或 csv]

# 示例演练：
.\ditting-cli.exe -path "E:\我的项目\whispers\DiTing\test_all" -format json
```
*执行完毕后，执行目录将自动生成一份类似 `report_16888888.json` 等脱水版结果凭证文件，完美对接后续基建。*

---
**DiTing 的诞生，只为重新定义云原生时代的静态隐私守护。**
