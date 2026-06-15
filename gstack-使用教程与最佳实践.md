# gstack 完整使用教程与最佳实践

> 由 Garry Tan (YC CEO 开源的 AI 工作流框架 — 让一个人像一个团队一样工作

## 📚 目录

1. [项目概述](#项目概述)
2. [快速安装](#快速安装)
3. [核心工作流](#核心工作流)
4. [所有 Skill 详细说明](#所有-skill-详细说明)
5. [最佳实践](#最佳实践)
6. [高级功能](#高级功能)
7. [常见问题](#常见问题)

---

## 项目概述

### gstack 是 Garry Tan 开源的 AI 编码工作流框架，它将 Claude Code 转变为一个虚拟工程团队。使用 gstack，一个人可以像 20 个人一样高效工作。

### 核心理念

gstack 遵循 **"Boil the Lake"** 原则 — 当 AI 使得边际成本几乎为零时，应该做完整的事情，而不是渐进的版本。

### 生产力数据

根据 Garry Tan 的真实使用数据：

- **2026 年运行速率：** ~810× 2013 年的速率（11,417 vs 14 行/天）
- **2026 年迄今：** 240× 整个 2013 年
- **并行冲刺：** 常规运行 10-15 个并行冲刺
- **贡献者：** 1,237+ 贡献（2026 年）

### 支持的 AI 代理

- Claude Code（主要目标）
- OpenAI Codex CLI
- OpenCode
- Cursor
- Factory Droid
- Slate
- Kiro
- Hermes
- GBrain

---

## 快速安装

### 前置要求
- Claude Code
- Git
- Bun v1.0+
- Node.js（仅 Windows）

### 方法 1：标准安装（30 秒）

打开 Claude Code 并粘贴：

```bash
git clone --single-branch --depth 1 https://github.com/garrytan/gstack.git ~/.claude/skills/gstack && cd ~/.claude/skills/gstack && ./setup
```

然后在 `CLAUDE.md` 中添加：

```markdown
## gstack

使用 gstack 的 /browse skill 进行所有网页浏览，永远不要使用 mcp__claude-in-chrome__* 工具。

可用技能：
/office-hours, /plan-ceo-review, /plan-eng-review, /plan-design-review, /design-consultation, /design-shotgun, /design-html, /review, /ship, /land-and-deploy, /canary, /benchmark, /browse, /connect-chrome, /qa, /qa-only, /design-review, /setup-browser-cookies, /setup-deploy, /setup-gbrain, /retro, /investigate, /document-release, /codex, /cso, /autoplan, /plan-devex-review, /devex-review, /careful, /freeze, /guard, /unfreeze, /gstack-upgrade, /learn
```

### 方法 2：团队模式（推荐共享仓库）

```bash
(cd ~/.claude/skills/gstack && ./setup --team) && ~/.claude/skills/gstack/bin/gstack-team-init required && git add .claude/ CLAUDE.md && git commit -m "require gstack for AI-assisted work"
```

### 方法 3：OpenClaw 集成

安装 gstack 后，告诉 OpenClaw 在生成 Claude Code 会话时使用 gstack skill。

**原生 OpenClaw 技能（通过 ClawHub）：

```bash
clawhub install gstack-openclaw-office-hours gstack-openclaw-ceo-review gstack-openclaw-investigate gstack-openclaw-retro
```

### 卸载

```bash
~/.claude/skills/gstack/bin/gstack-uninstall
```

---

## 核心工作流

### 完整 Sprint 流程

gstack 按照 Sprint 流程按照实际工程团队的工作流程完全一致：

```
思考 → 规划 → 构建 → 审查 → 测试 → 交付 → 反思
```

每个 skill 会将结果传递给下一个，形成完整的闭环。

### 推荐起始顺序

| 阶段 | Skill | 角色 | 产出 |
|------|-------|------|------|
| 思考 | `/office-hours` | YC 合伙人 | 设计文档 |
| 规划 | `/plan-ceo-review` | CEO/创始人 | 产品愿景 |
| 规划 | `/plan-eng-review` | 工程经理 | 技术架构 |
| 规划 | `/plan-design-review` | 高级设计师 | 设计方案 |
| 构建 | 实现代码 | 开发 |
| 审查 | `/review` | 资深工程师 | 代码审查 |
| 测试 | `/qa` | QA 主管 | 测试 + 修复 |
| 交付 | `/ship` | 发布工程师 | PR + 测试 |
| 反思 | `/retro` | 工程经理 | 回顾总结 |

### 快速入门五步走

1. 运行 `/office-hours` — 描述你正在构建什么
2. 运行 `/plan-ceo-review` — 对任何功能想法
3. 运行 `/review` — 对任何有变更的分支
4. 运行 `/qa` — 对你的 staging URL
5. 运行 `/ship` — 交付

---

## 所有 Skill 详细说明

### 🎯 规划类 Skill

#### `/office-hours` — YC Office Hours

**角色：** YC 合伙人
**这是每个项目应该开始的地方。**

##### 核心功能
- 六个强制性问题，重新构建你的产品
- 挑战你的框架假设
- 生成 2-3 个实现方案，带有诚实的工作量估算
- 写入设计文档，供下游 skill 使用

##### 工作流程
1. **重构框架** — 倾听你的痛点，不是你的功能请求
2. **前提挑战** — 提出可证伪的产品主张
3. **实现替代方案** — 生成 2-3 个具体方法
4. **设计文档** — 写入 `~/.gstack/projects/`

##### 两种模式

**创业模式** — 面向创始人：需求真实性、现状、狭窄切入点、观察与惊喜
**构建者模式** — 面向黑客松、副业、开源：热情的合作者，帮你找到最酷的版本

##### 最佳实践
- 永远从这里开始，不要跳过规划前写代码
- 诚实回答痛点问题，不要试图给 AI "正确"答案

---

#### `/plan-ceo-review` — CEO 审查

**角色：** CEO/创始人
**Brian Chesky 模式 — 找到隐藏在请求中的 10 星产品。**

##### 核心功能
- 重新思考问题，从用户角度出发
- 找到感觉不可避免、令人愉悦、甚至有点神奇的版本
- 四种范围扩展机会

##### 四种模式

| 模式 | 行为 |
|------|------|
| **SCOPE EXPANSION** | 大胆梦想，积极推荐扩展 |
| **SELECTIVE EXPANSION** | 保持当前范围为基线，中性推荐 |
| **HOLD SCOPE** | 对现有计划最大严谨性，无扩展 |
| **SCOPE REDUCTION** | 找到最小可行版本，砍掉其他 |

##### 示例

> 用户说："让卖家上传商品照片"

弱助手：添加文件选择器并保存图像。

`/plan-ceo-review`：
- 我们能从照片中识别产品吗？
- 我们能自动推断 SKU 或型号吗？
- 我们能搜索网络并自动起草标题和描述吗？
- 我们能建议哪张照片作为主图转化率最好吗？

---

#### `/plan-eng-review` — 工程审查

**角色：** 工程经理
**构建可以承载产品愿景的技术脊柱。**

##### 核心功能
- 架构和系统边界
- 数据流和状态转换
- 故障模式和边缘情况
- 信任边界和测试覆盖
- **ASCII 图** — 强制隐藏假设浮出水面

##### 关键解锁

LLM 在被迫绘制系统时变得更加完整。序列图、状态图、组件图、数据流图、测试矩阵 — 图迫使隐藏假设浮出水面。

---

#### `/plan-design-review` — 设计审查（规划阶段）

**角色：** 高级设计师
**交互式规划模式设计审查。**

##### 核心功能
- 对每个设计维度评分 0-10
- 解释 10 分是什么样子
- 编辑计划以达到目标
- AI Slop 检测
- 每个设计选择一个 AskUserQuestion

##### 审计维度
- 视觉层次
- 间距和节奏
- 排版层次
- 颜色系统
- 组件一致性
- 响应式行为
- 可访问性
- 动效

---

#### `/plan-devex-review` — 开发者体验审查

**角色：** 开发者体验负责人
**交互式 DX 审查：探索开发者角色，基准竞争对手 TTHW，设计你的神奇时刻，逐步追踪摩擦点。**

##### 三种模式

| 模式 | 行为 |
|------|------|
| **DX EXPANSION** | 完整的 DX 扩展，20-45 个强制性问题 |
| **DX POLISH** | 细化现有计划，关注摩擦点 |
| **DX TRIAGE** | 快速分类，只关注最关键问题 |

---

#### `/autoplan` — 自动规划

**角色：** 审查流水线
**一个命令，完全审查的计划。**

##### 核心功能
- 自动运行 CEO → 设计 → 工程 → DX 审查
- 编码决策原则
- 只向你呈现品味决策供批准

---

### 🎨 设计类 Skill

#### `/design-consultation` — 设计咨询

**角色：** 设计合作伙伴
**从零开始构建完整的设计系统。**

##### 核心功能
- 研究领域现状
- 提出创造性风险
- 生成逼真的产品模型
- 写入 `DESIGN.md`

---

#### `/design-shotgun` — 设计探索

**角色：** 设计探索者
**"给我看选项"模式。**

##### 工作流程
1. 生成 4-6 个 AI 模型变体
2. 在浏览器中打开比较板
3. 收集你的反馈
4. 迭代直到你喜欢某个方向
5. 味觉记忆开始偏向你的偏好

##### 味觉学习
批准和拒绝被写入持久的每项目味觉配置文件，每周衰减 5%。未来变体生成会偏向你实际选择的内容。

---

#### `/design-html` — 设计转 HTML

**角色：** 设计工程师
**将模型转换为实际工作的生产级 HTML。**

##### 核心功能
- 使用 Pretext 计算文本布局
- 文本在调整大小时实际重排
- 高度根据内容调整
- 布局是动态的
- 30KB 开销，零依赖
- 检测你的框架（React、Svelte、Vue）
- 根据设计类型智能 API 路由

---

### 🔍 审查类 Skill

#### `/review` — 代码审查

**角色：** 资深工程师
**找到通过 CI 但在生产中爆炸的 bug。**

##### 核心功能
- 自动修复明显的问题
- 标记完整性差距
- 关注生产问题，不是风格问题
- 每个发现带有具体的故障场景

##### 审查维度
- 竞态条件
- 错误处理
- 边界情况
- 空指针
- 性能回归

---

#### `/investigate` — 调试

**角色：** 调试器
**系统根本原因调试。**

##### 铁律
**没有调查就没有修复。**

##### 工作流程
1. 追踪数据流
2. 测试假设
3. 3 次失败修复后停止
4. 自动冻结到被调查的模块

---

#### `/design-review` — 设计审查（现场）

**角色：** 会写代码的设计师
**现场视觉审计 + 修复循环。**

##### 核心功能
- 80 项审计清单
- 原子提交
- 前后截图
- 修复发现的问题

---

#### `/devex-review` — 开发者体验审查（现场）

**角色：** DX 测试员
**现场开发者体验审计。**

##### 核心功能
- 实际测试你的入职流程
- 浏览文档
- 尝试入门流程
- 计时 TTHW（Time-To-Hello-World）
- 截图错误
- 与 `/plan-devex-review` 分数比较

---

#### `/codex` — 第二意见

**角色：** 第二意见
**来自 OpenAI Codex CLI 的独立审查。**

##### 三种模式

| 模式 | 行为 |
|------|------|
| **review** | 代码审查，通过/失败门 |
| **adversarial** | 对抗性挑战，主动尝试破坏你的代码 |
| **consult** | 开放咨询，会话连续性 |

##### 跨模型分析
当 `/review`（Claude）和 `/codex`（OpenAI）都审查了同一个分支时，你会得到跨模型分析，显示哪些发现重叠，哪些是每个独有的。

---

### 🧪 测试类 Skill

#### `/qa` — QA 测试

**角色：** QA 主管
**测试你的应用，找到 bug，用原子提交修复它们，重新验证。**

##### 核心功能
- 真实 Chromium 浏览器
- 真实点击，真实截图
- 每个命令 ~100ms
- 为每个修复自动生成回归测试
- 原子提交，每个 bug 一个提交

##### 浏览器命令 (`$B`)

```bash
$B navigate https://example.com      # 导航到 URL
$B click "button"                # 点击元素
$B fill "input" "value"           # 填充输入
$B snapshot                        # 截图
$B snapshot -i                     # 带注释的截图
$B diff before after              # 前后差异
$B verify "text"                     # 验证文本存在
$B form submit                    # 提交表单
$B dialog accept                  # 接受对话框
$B upload "file.txt"             # 上传文件
$B responsive 375x812             # 测试响应式
$B wait 2s                         # 等待
$B wait for "selector"             # 等待元素
$B scroll bottom                  # 滚动到底部
$B eval "js code"                # 执行 JS
$B extract "selector"               # 提取元素
$B cookies import chrome              # 导入 Chrome cookie
$B list                             # 列出所有标签页
$B close                            # 关闭当前标签页
$B new                              # 新建标签页
$B switch 1                      # 切换到标签页 1
$B handoff                          # 交接给用户
$B resume                           # 从交接恢复
$B disconnect                    # 断开浏览器
```

---

#### `/qa-only` — 仅报告

**角色：** QA 报告员
**与 `/qa` 相同的方法，但只报告。**

使用场景：
- 你想要纯 bug 报告，不需要代码更改
- 你想在修复前先查看所有问题

---

### 🚀 交付类 Skill

#### `/ship` — 交付

**角色：** 发布工程师
**一个命令，从分支到 PR。**

##### 工作流程
1. 同步 main 分支
2. 运行测试
3. 审计覆盖率
4. 推送分支
5. 打开 PR

##### 测试引导
如果你的项目没有测试框架，`/ship` 会自动引导一个。

---

#### `/land-and-deploy` — 合并部署

**角色：** 发布工程师
**一个命令，从"已批准"到"生产中验证"。**

##### 工作流程
1. 合并 PR
2. 等待 CI 和部署
3. 验证生产健康

---

#### `/canary` — 金丝雀监控

**角色：** SRE
**部署后监控循环。**

监控内容：
- 控制台错误
- 性能回归
- 页面失败

---

#### `/benchmark` — 性能基准

**角色：** 性能工程师
**页面加载时间、Core Web Vitals、资源大小基准。**

##### 核心功能
- 每个 PR 前后比较
- 随时间跟踪趋势
- 输出为表格、JSON、Markdown

---

#### `/document-release` — 发布文档

**角色：** 技术作家
**更新所有项目文档以匹配你刚刚交付的内容。**

自动捕获过时的 README。

---

### 🔒 安全类 Skill

#### `/cso` — 首席安全官

**角色：** 首席安全官
**OWASP Top 10 + STRIDE 威胁建模安全审计。**

##### 核心功能
- 零噪音：17 个误报排除
- 8/10+ 置信度门
- 独立发现验证
- 每个发现都包含具体的利用场景

##### 审计范围
- 注入攻击
- 认证问题
- 加密问题
- 访问控制问题
- 数据泄露
- 业务逻辑漏洞

---

### 🛡️ 安全与工具类 Skill

#### `/careful` — 小心模式

**角色：** 安全护栏
**在破坏性命令前警告。**

触发警告的命令：
- `rm -rf`
- `DROP TABLE`
- `force-push`
- `git reset --hard`

可以覆盖任何警告。常见构建清理列入白名单。

---

#### `/freeze` — 冻结

**角色：** 编辑锁
**将所有文件编辑限制到单个目录。**

在调试时防止意外更改范围外的代码。

---

#### `/guard` — 全面保护

**角色：** 全面安全
**`/careful` + `/freeze` 一个命令。**

生产工作的最大安全性。

---

#### `/unfreeze` — 解锁

**角色：** 解锁
**移除 `/freeze` 边界，允许到处编辑。**

---

### 🌐 浏览器类 Skill

#### `/browse` — 浏览

**角色：** QA 工程师
**给代理眼睛。**

##### 架构：
- 真实的 Chromium 守护进程，真实的点击，真实的截图
- 第一个命令启动一切（~3秒），之后每个命令 ~100-200ms
- 持久状态：登录一次，保持登录状态
- 30 分钟空闲超时后自动关闭

---

#### `/open-gstack-browser` — GStack 浏览器

**角色：** GStack 浏览器
**启动带有侧边栏的 GStack 浏览器。**

##### 核心功能
- 反机器人隐身
- 自动模型路由（动作使用 Sonnet，分析使用 Opus）
- 一键 cookie 导入
- Claude Code 集成
- 实时观看每个动作

##### 侧边栏代理
在 Chrome 侧边栏中输入自然语言，子 Claude 实例执行它。每个任务最多 5 分钟。隔离会话，不会干扰你的主 Claude Code 窗口。

---

#### `/setup-browser-cookies` — 浏览器 Cookie

**角色：** 会话管理器
**从你的真实浏览器导入 cookie。**

支持的浏览器：
- Chrome
- Arc
- Brave
- Edge

测试认证页面。第一次每个浏览器的 cookie 导入会触发 macOS Keychain 对话框。

---

#### `/pair-agent` — 配对代理

**角色：** 远程代理桥
**将远程 AI 代理与你的浏览器配对。**

支持的代理：
- OpenClaw
- Codex
- Cursor
- Hermes

##### 安全特性
- 作用域隧道
- 锁定的允许列表
- 会话令牌
- 速率限制
- 域名限制
- 活动归因

---

### 🧠 记忆与学习类 Skill

#### `/learn` — 学习

**角色：** 记忆
**管理 gstack 跨会话学到的内容。**

##### 核心功能
- 审查学到的内容
- 搜索
- 修剪
- 导出自定义模式
- 导出项目特定的模式和偏好

学习跨会话复合，所以 gstack 在你的代码库上变得更智能。

---

#### `/setup-gbrain` — GBrain 设置

**角色：** 记忆同步
**从零到运行 gbrain 不到 5 分钟。**

##### 三条路径

| 路径 | 描述 |
|------|------|
| **Supabase 现有 URL** | 你的云代理已经配置了大脑 |
| **Supabase 自动配置** | 粘贴 Supabase PAT，自动创建新项目 |
| **PGLite 本地** | 零账户，零网络，仅在这台 Mac 上隔离 |

##### 每个远程信任策略

| 策略 | 行为 |
|------|------|
| **read-write** | 代理可以搜索大脑并从这个仓库写入新页面 |
| **read-only** | 代理可以搜索但永远不写入 |
| **deny** | 完全没有 gbrain 交互 |

---

#### `/retro` — 回顾

**角色：** 工程经理
**团队感知的每周回顾。**

##### 内容
- 每人分解
- 交付连胜
- 测试健康趋势
- 成长机会

`/retro global` 跨所有你的项目和 AI 工具运行。

---

### 🔧 其他 Skill

#### `/gstack-upgrade` — 自更新

**角色：** 自更新器
**将 gstack 升级到最新版本。**

自动检测全局 vs  vendored 安装，同步两者，显示更改内容。

---

#### `/context-save` — 保存上下文

**角色：** 保存状态
**保存工作上下文（git 状态、决策、剩余工作）以便任何未来会话可以恢复。**

---

#### `/context-restore` — 恢复上下文

**角色：** 恢复状态
**从保存的上下文恢复，即使跨 Conductor 工作区交接。**

---

#### `/health` — 健康

**角色：** 代码质量仪表板
**包装类型检查器、linter、测试、死代码检测。**

计算加权 0-10 分数；随时间跟踪趋势。

---

#### `/scrape` — 抓取

**角色：** 浏览器数据提取器
**从网页提取数据。**

第一个调用通过 `$B` 原型化；匹配意图的后续调用在 ~200ms 内运行编码的浏览器技能。

---

#### `/skillify` — 技能化

**角色：** 技能编码器
**回溯你的对话，找到最后一个 `/scrape` 原型，合成脚本 + 测试 +  fixture，运行测试，在提交前询问。**

---

#### `/plan-tune` — 问题调优

**角色：** 问题调谐器
**每个问题自调整 AskUserQuestion 灵敏度。**

将问题标记为 never-ask、always-ask 或 only-for-one-way。

---

## 最佳实践

### 🎯 工作流最佳实践

1. **永远从 `/office-hours` 开始**
   - 不要跳过思考阶段直接编码
   - 花时间让框架被挑战
   - 设计文档是所有下游工作的单源真相

2. **使用正确的审查类型

| 为谁构建 | 规划阶段（代码前） | 现场审计（交付后） |
|----------|-------------------|-------------------|
| 最终用户（UI、Web 应用、移动端 | `/plan-design-review` | `/design-review` |
| 开发者（API、CLI、SDK、文档 | `/plan-devex-review` | `/devex-review` |
| 架构（数据流、性能、测试） | `/plan-eng-review` | `/review` |
| 以上所有 | `/autoplan` | — |

3. **并行冲刺最佳实践**

- 10-15 个并行冲刺是实际最大值
- 没有流程，10 个代理就是 10 个混乱源
- 像 CEO 管理团队一样管理它们：检查重要决策，让其余运行

4. **测试一切**
   - 每个 `/ship` 运行产生覆盖率审计
   - 每个 `/qa` bug 修复产生回归测试
   - 100% 测试覆盖率是目标
   - 测试让 vibe 编码安全，而不是 YOLO 编码

5. **连续检查点模式（选择加入）**

设置 `gstack-config set checkpoint_mode continuous`

- 技能在你工作时自动提交你的工作
- 前缀 `WIP:` 加上结构化的 `[gstack-context]` 正文
- 在崩溃和上下文切换中幸存
- `/context-restore` 读取这些提交来重建会话状态
- `/ship` 在 PR 前过滤压缩 WIP 提交

---

### 🔒 安全最佳实践

1. **在生产工作中始终使用 `/guard`**
   - 组合 `/careful` + `/freeze`
   - 防止意外删除和范围外编辑

2. **提示注入防御**

gstack  shipping 带有分层防御：
- 22MB ML 分类器与浏览器捆绑，本地扫描每个页面和工具输出
- Claude Haiku 转录检查对完整对话形状进行投票
- 系统提示中的随机金丝雀令牌在文本、工具参数、URL 和文件写入中捕获会话泄露尝试
- 判决组合器要求两个分类器一致才能阻止
- 选择加入 721MB DeBERTa-v3 集成通过 `GSTACK_SECURITY_ENSEMBLE=deberta` 用于 2-of-3 协议

3. **浏览器交接**
当 AI 遇到 CAPTCHA、认证墙或 MFA 提示时：
```bash
$B handoff   # 打开可见的 Chrome 在完全相同的页面
# 你解决问题
$B resume    # 从停止的地方继续
```

---

### 💡 生产力最佳实践

1. **语音输入**
gstack 技能具有语音友好的触发短语。自然地说你想要什么：
- "运行安全检查" → `/cso`
- "测试网站" → `/qa`
- "进行工程审查" → `/plan-eng-review`

2. **主动技能建议**
gstack 注意到你处于哪个阶段 — 头脑风暴、审查、调试、测试 — 并建议正确的技能。不喜欢的话可以说"停止建议"，它会跨会话记住。

3. **领域技能**
```bash
$B domain-skill save
```
代理保存每个站点的注释（例如，"LinkedIn 的 Apply 按钮位于 iframe 中"），下次访问该主机名时自动触发。隔离 → 3 次成功使用后激活 → 可选的跨项目提升。

---

## 高级功能

### 双监听器隧道架构（v1.6.0.0）

当用户运行 `pair-agent --client` 时，守护进程启动 ngrok 隧道，以便远程配对代理可以驱动浏览器。

**两个 HTTP 监听器，不是一个：

- **本地监听器** (`127.0.0.1:LOCAL_PORT`) — 始终绑定。服务引导、`/cookie-picker`、`/inspector/*`、`/welcome`、`/refs`、侧边栏代理 API 和完整命令表面。永远不转发。
- **隧道监听器** (`127.0.0.1:TUNNEL_PORT`) — 在 `/tunnel/start` 上延迟绑定，在 `/tunnel/stop` 上拆除。服务锁定的允许列表：`/connect`（配对仪式，未认证 + 速率限制）、`/command`（仅作用域令牌，进一步限制为浏览器驱动命令允许列表）和 `/sidebar-chat`。其他所有内容 404。

### 原始 CDP 转义口

```bash
$B cdp <Domain.method>
```

原始 Chrome DevTools 协议转义口，用于罕见的策划命令遗漏的情况。拒绝默认：方法必须显式添加到 `browse/src/cdp-allowlist.ts` 中，并带有一行理由。

### 模型基准测试

```bash
gstack-model-benchmark
```

跨模型基准测试：通过 Claude、GPT（通过 Codex CLI）和 Gemini 运行相同的提示；比较延迟、令牌、成本和（可选）LLM 判断的质量分数。输出为表格、JSON 或 Markdown。

---

## 常见问题

### Q: gstack 和 Claude Code 有什么区别？

A: Claude Code 是平台。gstack 是运行在其上的 opinionated 工作流技能集合。把 Claude Code 想象成操作系统，gstack 是应用程序。

### Q: 我需要所有技能吗？

A: 不需要。从 5 个核心技能开始：`/office-hours` → `/plan-ceo-review` → `/review` → `/qa` → `/ship`。其余的在你需要时添加。

### Q: 这只适用于 Web 应用吗？

A: 不。gstack 适用于任何类型的软件项目。浏览器技能对 Web 应用最有用，但规划、审查、测试和交付技能适用于所有类型的项目。

### Q: 我可以在团队中使用吗？

A: 是的。团队模式自动为所有队友自动更新。没有 vendored 文件，没有版本漂移，没有手动升级。

### Q: 我的数据去哪了？

A: 默认情况下，所有内容都保留在你的机器上。分析是选择加入的。gbrain 同步是选择加入的，所有数据都保留在你的机器上，除非你明确选择加入遥测。

### Q: 支持 Windows 吗？

A: 是的。gstack 在 macOS、Linux 和 Windows 上原生运行。

---

## 许可证

MIT 许可证。免费且开源。没有高级层，没有等待列表。

---

*本教程基于 Garry Tan 的 gstack v1.25.1.0 版本编写。

更多信息请访问：https://github.com/garrytan/gstack