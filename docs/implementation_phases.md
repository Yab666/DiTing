# 谛听 (DiTing) 项目实施阶段与过程说明 (对照原版 Whispers)

本项目旨在将原版 Python 编写的 [Whispers](https://github.com/Skyscanner/whispers) 静态分析工具重构为高性能、高解耦的 Go 语言版本。以下是对照原版功能的实施阶段说明。

## 第一阶段：骨架搭建与文件遍历 (对应原版 Core & Traverse)

**目标**：建立项目基础架构，实现高效的文件系统扫描。

*   **Go 实施内容**：
    *   `internal/scanner`: 实现高并发的文件遍历逻辑（对应原版 `traverse.py`）。
    *   `internal/app/engine.go`: 实现核心编排引擎（对应原版 `core.py`）。
    *   `cmd/whispers-cli`: 实现基础命令行交互（对应原版 `cli.py`）。
*   **功能对齐**：支持递归扫描目录、排除指定文件/路径、基本的命令行参数处理。

## 第二阶段：基础插件与规则引擎 (对应原版 Plugins & Rules)

**目标**：实现结构化数据的解析与秘密检测逻辑。

*   **Go 实施内容**：
    *   `internal/plugin`: 实现 `Parser` 接口，并完成 YAML、JSON、XML 等基础解析插件（对应原版 `plugins/yml.py`, `json.py`, `xml.py`）。
    *   `internal/rule`: 移植原版的规则判定逻辑，包括正则匹配、长度校验、Base64 检查等（对应原版 `secrets.py` 和 `rules/` 目录）。
    *   `configs/rules`: 移植并优化原版的所有 YAML 规则。
*   **功能对齐**：能够准确识别 YAML/JSON/XML 配置文件中的敏感信息，确保检出率不低于原版。

## 第三阶段：高级语言解析与 AST (对应原版 Python/JS/Go Plugins)

**目标**：实现对代码文件的深度语义解析。

*   **Go 实施内容**：
    *   `internal/plugin/python`: 使用抽象语法树 (AST) 模式解析 Python 代码（对应原版 `plugins/python.py`）。
    *   `internal/plugin/plaintext`: 实现对 JavaScript、Java、Go 等语言的通用文本匹配插件（对应原版 `plugins/javascript.py`, `java.py`, `go.py`）。
*   **功能对齐**：支持对 Python 代码中的硬编码字符串进行 AST 级深度识别，而非简单的正则表达式。

## 第四阶段：AI 误报过滤与 Web 管理 (新增增强功能)

**目标**：利用 LLM 降低误报率，并提供可视化管理界面。

*   **Go 实施内容**：
    *   `internal/ai`: 封装 LLM API 调用逻辑，对初步发现的可疑结果进行二次审查（原版无此功能）。
    *   `cmd/whispers-web`: 构建后端服务，提供扫描记录管理、规则配置及可视化报告页面。
*   **价值体现**：解决原版工具在复杂逻辑下误报较多的痛点，提升用户体验。

## 第五阶段：并发优化与集成交付 (最终对齐与性能超越)

**目标**：通过并发榨干机器性能，完成全流程验证。

*   **Go 实施内容**：
    *   **性能榨取**：优化 Goroutine 池，确保扫描速度相比 Python 原版有量级提升。
    *   **单文件打包**：利用 `go:embed` 将所有静态资源（规则、Web 前端）打包进一个二级制文件。
    *   **集成测试**：使用 `test/fixtures` 的所有用例进行回归测试，确保功能完全覆盖原版。
*   **最终交付**：提供 Makefile、Docker 一键构建方案。

---

## 核心模块对照表 (Whispers vs. DiTing)

| 功能模块 | 原版 Python (Whispers) | 新版 Go (DiTing) | 说明 |
| :--- | :--- | :--- | :--- |
| **基础骨架** | `core.py`, `cli.py` | `internal/app/`, `cmd/` | Go 版本采用 Clean Architecture，解耦更彻底 |
| **文件遍历** | `traverse.py` | `internal/scanner/` | Go 版本利用 Goroutine 实现原生并发扫描 |
| **解析插件** | `plugins/` (类实现) | `internal/plugin/` (接口实现) | 统一的 `Parser` 接口，扩展更简单 |
| **规则引擎** | `secrets.py`, `rules/` | `internal/rule/`, `configs/rules/` | 逻辑移植并优化，支持熵计算与正则 |
| **高级解析** | `plugins/python.py` (AST) | `internal/plugin/python` | 保持 AST 解析能力，确保对代码的深理解 |
| **误报过滤** | - (无) | `internal/ai/` | **[新增]** 集成 LLM 智能审计，大幅降低误报 |
| **管理界面** | - (无) | `cmd/whispers-web` | **[新增]** 可视化后台，支持规则配置与记录查询 |
| **部署形态** | Pip 安装 / 环境依赖 | **单二进制文件** | 利用 `go:embed` 实现零依赖一键运行 |
