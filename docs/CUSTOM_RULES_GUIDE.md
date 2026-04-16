# 🛠️ DiTing (谛听) 自定义规则配置指南

DiTing 采用了 **YAML 驱动的规则引擎**，支持在不修改任何源代码的情况下，通过配置 YAML 文件来扩展查杀能力。

## 1. 规则存存放目录
所有规则文件统一存放在以下目录：
`configs/rules/`

引擎在启动时会自动扫描该目录下的所有 `.yaml` 文件并动态加载。

---

## 2. 规则基本结构

每个 YAML 文件可以包含一个或多个规则。规则的核心是基于正则表达式的 `Key` (变量名) 和 `Value` (具体内容) 匹配。

### 标准模板
```yaml
rule-id:                 # 规则唯一标识符
  description: "..."     # 规则详细描述
  message: "..."         # 检出时的警告消息
  severity: "CRITICAL"   # 严重等级：CRITICAL | MAJOR | MINOR | INFO
  
  # 变量名匹配
  key:
    regex: "regexp"      # 匹配变量名的正则表达式
    ignorecase: True     # 是否忽略大小写
    
  # 内容匹配
  value:
    regex: "regexp"      # 匹配变量值的正则表达式
    minlen: 8            # 最小命中长度过滤（防止误报）
    
  # 相似度过滤
  similar: 0.7           # 如果 Key 和 Value 相似度高于此值，则视为占位符并过滤
```

---

## 3. 常见规则配置示例

### 案例 A：识别公司内部机密 Token
如果您公司内部有一种以 `CORP_` 开头且长度为 32 位的 Token。

```yaml
corp-internal-token:
  description: "公司内部专用 Token 审计"
  message: "发现公司生产环境加密 Token，请确保其未被同步到外网！"
  severity: "CRITICAL"
  key:
    regex: "(CORP_TOKEN|INTERNAL_KEY)"
  value:
    regex: "[a-zA-Z0-9]{32}"
    minlen: 32
```

### 案例 B：敏感机密文件发现
通过匹配虚拟 Key `file` 来识别敏感文件（如私钥、证书、配置文件）。

```yaml
private-key-file:
  description: "RSA 私钥文件"
  message: "发现潜在的非加密私钥文件，可能导致链路被破解"
  severity: "MAJOR"
  key:
    regex: "^file$"
  value:
    regex: ".*\\.(pem|key|p12)$"
```

---

## 4. 如何生效

### 🚀 热加载机制
DiTing 的设计初衷是**开箱即用**：
1. **无需重新编译**：您可以直接在运行中的系统目录下修改或新增 YAML 文件。
2. **即时重扫**：
   - **CLI 模式**：下次运行扫描命令时会自动读取新规则。
   - **Web 模式**：无需重启后台，直接在网页点击“启动雷达扫描”，后端会自动完成规则的重载。

---

## 5. 开发建议
1. **测试正则表达式**：建议在提交规则前使用在线工具（如 regex101）进行验证。
2. **合理设置等级**：只有确认会导致代码泄露的项才建议设为 `CRITICAL`，一般的配置追踪建议设为 `INFO`。
3. **分文件管理**：建议按业务类型（如 `cloud_provider.yaml`, `database.yaml`）进行分类存放，便于维护。
