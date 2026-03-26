# 谛听 (DiTing) - Whispers Go 语言重构版

本项目是 [Whispers](https://github.com/Skyscanner/whispers) 的 Go 语言实现版本，重构旨在提高其在高度解耦、可扩展性以及高性能并发扫描方面的能力。

## 核心设计

本项目遵循 **Clean Architecture (整洁架构)**：

- **api/**: API 定义 (OpenAPI/Swagger)
- **cmd/**: 入口工程
  - `whispers-cli/`: 命令行工具
  - `whispers-web/`: Web 服务后端
- **internal/**: 业务核心 (私有)
  - `app/`: 编排层 (Engine)
  - `plugin/`: 解析插件接口定义
  - `scanner/`: 并发扫描逻辑
  - `ai/`: 大模型误报过滤
- **pkg/**: 公用工具库
- **docs/**: 技术文档

## 快速开始

1.  **初始化 Go 模块**：
    由于环境路径问题，请在终端手动运行：
    ```bash
    cd DiTing
    go mod init ditting
    ```

2.  **查看项目实施计划**：
    请查阅 [docs/implementation_phases.md](docs/implementation_phases.md) 获取详细的开发阶段说明。

## 设计亮点

- **高解耦插件系统**：通过接口支持任意格式解析。
- **AI 辅助审计**：集成 LLM 二次校验扫描结果，降低误报。
- **单文件部署**：使用 `go:embed` 嵌入规则文件。
