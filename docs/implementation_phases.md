# 谛听 (DiTing) 项目实施阶段与过程说明

本项目旨在将 Whispers 静态代码分析工具重构为高性能、高解耦的 Go 语言版本。以下是项目的实施阶段与开发过程说明。

## 第一阶段：基础设施与核心引擎构建 (Foundation & Core Engine)

**目标**：建立项目骨架，实现基础的扫描逻辑。

1.  **环境初始化**：
    *   建立基于 Clean Architecture 的目录结构。
    *   初始化 Go Module。
2.  **领域模型定义 (`internal/core`)**：
    *   定义 `Secret`（发现的密钥）、`Rule`（匹配规则）、`Result`（扫描结果）等核心结构体。
3.  **核心接口协议 (`internal/plugin`)**：
    *   定义 `Parser` 接口，规定各类文件解析器必须实现的方法。
4.  **扫描引擎开发 (`internal/app`)**：
    *   实现 `Engine` 结构，负责串联文件遍历、内容解析与规则匹配。
    *   采用依赖注入模式，使引擎不直接依赖具体的解析器实现。

## 第二阶段：插件系统与规则集成 (Plugins & Rules)

**目标**：支持多种文件格式解析，并集成安全规则。

1.  **解析器实现**：
    *   开发 YAML、JSON、Python、Shell 等常见格式的解析插件库。
    *   实现文本匹配插件（正则引擎）。
2.  **规则集加载 (`configs`)**：
    *   集成硬编码的敏感信息正则规则。
    *   使用 `//go:embed` 将默认规则文件打包进二进制文件，实现零依赖运行。
3.  **并发优化 (`internal/scanner`)**：
    *   实现基于 Goroutine 的并发扫描器，支持大数据量下的快速遍历。

## 第三阶段：AI 误报过滤模块 (AI-Driven False Positive Filtering)

**目标**：结合 LLM 降低扫描结果的误报率。

1.  **AI 模块开发 (`internal/ai`)**：
    *   封装与主流大模型（如 GPT、DeepSeek）的 API 交互逻辑。
    *   设计针对密钥特征的 Prompt 模板。
2.  **判定逻辑集成**：
    *   在扫描流程末端接入 AI 二次验证，对可疑结果进行“人工级”审查。

## 第四阶段：多维度输出与 Web 管理端 (Output & Management UI)

**目标**：提供专业报告并开启可视化管理。

1.  **报告模块 (`internal/report`)**：
    *   实现 SARIF (静态分析结果交换格式) 输出。
    *   支持 JSON 原始数据及可视化可视化 HTML 报告。
2.  **Web API 与 UI (`internal/ui` & `cmd/whispers-web`)**：
    *   构建轻量级 RESTful API 接口。
    *   集成 Web 管理后台，提供规则管理、历史记录查询及扫描可视化。

## 第五阶段：集成测试与成果交付 (Integration & Finalization)

**目标**：确保系统稳定性并完成最终构建。

1.  **集成测试 (`test/fixtures`)**：
    *   准备针对各类假密钥的测试用例库，验证全流程检出率。
2.  **自动化构建**：
    *   编写 Makefile 和 Dockerfile，实现跨平台一键编译。
3.  **文档完善**：
    *   完成 API 文档及用户操作指南。
