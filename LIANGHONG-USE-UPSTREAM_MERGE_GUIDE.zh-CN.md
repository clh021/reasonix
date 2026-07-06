# `lianghong-use` 上游合并指引

本文档用于指导 `lianghong-use` 分支后续从上游 `main-v2` 同步代码时的原则、步骤和交付标准。

默认目标不是“在旧分支上继续累积 merge 历史”，而是：

1. 优先吸收上游 `main-v2` 的正式实现。
2. 只有在上游**没有覆盖**我们自己的需求时，才保留自己的代码。
3. 尽可能减少 `lianghong-use` 相对上游的长期 diff 面积。
4. 每次同步结束后，都能清楚汇报“还剩多少我们自己的代码”。
5. 尽量把我们自己的修改整理成上游更新之后的少数几次尾部提交。

---

## 一句话原则

**如果上游 `main-v2` 新增的特性已经覆盖了我们自己的新增特性，就舍弃我们自己的实现，优先采用上游版本。**

换句话说：

- 上游已覆盖：删掉我们的。
- 上游部分覆盖：只保留缺失部分。
- 上游未覆盖：保留我们的。

---

## 适用范围

本文档默认用于：

- 当前工作分支：`lianghong-use`
- 上游对比分支：`origin/main-v2`

以后在 `lianghong-use` 上只需要专注自己的需求开发；当需要跟上游同步时，按本文档执行即可。

---

## 默认同步策略

**以后默认不要直接在旧的 `lianghong-use` 上继续 merge `origin/main-v2`。**

默认流程应当是：

1. 先审计当前 `lianghong-use` 相对 `origin/main-v2` 还剩哪些自定义能力。
2. 从 `origin/main-v2` 新开一条干净整理分支。
3. 只把“上游仍未覆盖、而且确实需要保留”的最小自定义代码重新整理进去。
4. 将这些自定义重放成上游更新之后的最后几次提交。
5. 验证通过后，再让 `lianghong-use` 指向这条整理后的干净历史。

这样做的目的：

- 避免 `lianghong-use` 历史里不断堆积旧 merge 噪音
- 让“当前真正还需要维护的 fork 代码”保持一眼可见
- 让以后每次再同步上游时，都只需要面对少量尾部提交

### 只有在下面情况才考虑直接 merge

- 当前相对 `origin/main-v2` 的自定义差异极小，且基本没有历史噪音
- 或者只是临时跟上游对齐一次，之后还会马上做重整

如果当前分支已经长期领先上游很多提交，但相对上游只剩少量真实自定义文件，那么**重建干净分支通常优于继续 merge**。

---

## 同步总策略

每次同步都遵循下面五条：

1. **以远端最新上游为准**
   不要以本地旧的 `main-v2` 为准，应直接对比 `origin/main-v2`。

2. **先判断“功能是否已覆盖”，再决定保留谁**
   不是看代码写法像不像，而是看用户可见能力和行为是否已经被上游实现。

3. **保留最小自定义面**
   即使必须保留自己的功能，也要尽量：
   - 拆到独立文件
   - 收敛到单个自包含 block
   - 避免改动上游高频变动文件的大段核心逻辑

4. **最终必须给出残留清单**
   同步结束后必须明确：
   - 还保留了哪些自定义文件
   - 这些文件分别承载什么功能
   - 哪些已经被上游覆盖并丢弃

5. **优先重放“当前仍需要的能力”，不要保留旧历史本身**
   我们需要保住的是需求，不是旧提交图。

---

## 标准操作流程

### 1. 先看分支状态和上游基线

```bash
git status --short --branch
git branch --all --verbose --no-abbrev
git fetch origin
```

重点确认：

- 当前是否在 `lianghong-use`
- 本地是否有未提交改动
- 本地 `main-v2` 是否过旧
- `origin/main-v2` 是否已拉到最新

### 2. 始终用 `origin/main-v2` 作为上游基线，先做自定义面审计

先算公共基线和自定义改动范围：

```bash
git merge-base lianghong-use origin/main-v2
git diff --name-status $(git merge-base lianghong-use origin/main-v2)..lianghong-use
git diff --stat $(git merge-base lianghong-use origin/main-v2)..lianghong-use
```

这一步的目的不是直接同步，而是先搞清楚 `lianghong-use` 自己到底还剩哪些真实自定义文件。

如果看到：

- 历史上领先上游很多提交
- 但 `git diff origin/main-v2..lianghong-use` 只剩少量文件

那么默认应进入“重建干净分支”流程，而不是继续在原分支上 merge。

### 3. 逐项判断“是否已被上游覆盖”

对每个自定义改动，必须归类到下面三类之一：

#### A. 已被上游覆盖

条件：

- 上游已有同等或更完整的正式实现
- 我们的实现只是早期版本、简化版本或临时版本

处理：

- 舍弃我们的实现
- 合并时以 `origin/main-v2` 为准

#### B. 部分被上游覆盖

条件：

- 上游实现了大部分能力
- 但仍缺少我们的一小段行为或边界修复

处理：

- 只保留“上游缺失的最小差异”
- 不保留整块旧实现

#### C. 上游未覆盖

条件：

- 上游没有对应功能
- 或者功能明显不等价

处理：

- 保留我们的实现
- 同时继续压缩 diff 面

---

## 冲突处理优先级

出现冲突时，默认按以下优先级处理：

### 优先级 1：上游正式特性

如果 `main-v2` 已有完整正式实现，优先保留上游。

典型情况：

- 上游认证体系比我们自己的简化 middleware 更完整
- 上游移动端 UI 已正式实现

### 优先级 2：我们的未覆盖能力

如果上游没有该能力，则保留我们的功能实现。

典型情况：

- Web `serve` 端项目切换
- 上游没有的真实 bugfix

### 优先级 3：尽量缩小维护面

即使保留我们的功能，也要尽量做以下收敛：

- **优先新增文件**，不要大改上游核心文件
- 如果必须改大文件，尽量收敛到：
  - 单个函数
  - 单个路由接线
  - 单个脚本 block
- 前端大 HTML 文件里，优先使用：
  - 运行时注入
  - 自包含脚本
  - 避免改全局 CSS / 静态 DOM / 主 i18n 字典

---

## 推荐同步命令

### 默认推荐：重建干净分支

先从上游最新版本起一条整理分支：

```bash
git checkout -b lianghong-use-rebuild-YYYYMMDD origin/main-v2
```

然后只把仍需保留的自定义代码，按能力分组重新整理进去。

推荐分组方式：

1. 文档与忽略项
2. `serve` 项目切换 / workspace 持久化
3. `serve` quick reply
4. `serve` 前端局部补丁
5. 桌面端独立 bugfix

每组都单独提交，目标是让最终历史接近：

```text
origin/main-v2
  └─ commit A: docs / repo-local ignore
  └─ commit B: serve project switch + workspace persistence
  └─ commit C: serve quick reply
  └─ commit D: transcript final scroll fix
```

整理分支验证通过后，再决定是否替换正式分支：

```bash
git checkout lianghong-use
git reset --hard lianghong-use-rebuild-YYYYMMDD
git push --force-with-lease origin lianghong-use
```

只有在确认没有其他人依赖旧历史时，才做这一步。

### 备选方案：直接 merge 旧分支

仅当差异很小、且没有明显历史噪音时，才使用下面方式：

```bash
git checkout lianghong-use
git merge --no-ff --no-commit origin/main-v2
```

如果有冲突：

```bash
git diff --name-only --diff-filter=U
```

逐个处理，处理完后：

```bash
git add <resolved files>
git commit -m "Merge origin/main-v2 into lianghong-use"
```

---

## 合并时的判断清单

每次处理一个自定义改动时，都问这 5 个问题：

1. 上游是否已经有同等功能？
2. 上游是否比我们的实现更完整？
3. 我们保留的到底是“功能”，还是只是“旧实现方式”？
4. 能否把保留部分缩小到更小的文件或更少的行数？
5. 这段代码下次合并时是否容易再次冲突？

如果第 1 或第 2 个问题答案是“是”，默认就应该丢掉我们的版本。

---

## 降低后续维护成本的做法

### 1. 把自定义逻辑拆到独立文件

优先：

- `internal/serve/project_switch.go`
- `internal/serve/workspace.go`
- 对应测试文件独立放置

避免：

- 把大量自定义逻辑堆在 `serve.go` 里

### 2. 保留“行为补丁”，删除“全局实现改写”

优先保留：

- 范围明确的 bugfix
- 单点行为补丁

优先删除：

- 影响整个系统行为的大改动
- 只是实现方式不同、但不构成真实需求的改写

### 3. 自定义前端逻辑尽量自包含

例如在 `internal/serve/index.html` 中：

- 优先把项目切换逻辑收敛到单个脚本 block
- 尽量不要修改上游的：
  - 全局样式定义
  - 主体 HTML 结构
  - 主 i18n 字典

---

## 重建干净分支时的额外要求

### 1. 不要机械 cherry-pick 全部旧提交

如果旧分支历史已经很长，默认不要：

- 一路 `cherry-pick` 旧提交
- 试图保留旧提交顺序
- 为了保留旧实现方式而保留代码

应该做的是：

- 只保留当前仍需要的能力
- 直接以当前最终文件内容为依据重放
- 能缩小的继续缩小

### 2. 每次重放后都重新判断是否还能继续删

重放到干净分支后，经常会发现：

- 某个旧补丁其实已经没必要了
- 某段代码只剩一小部分还值得保留
- 某个修改可以从“改大文件”收敛成“新增独立文件 + 一行接线”

这时优先继续删，直到只剩真正必要的最小差异。

---

## 同步完成后的必做验证

至少做下面几项：

### Go 后端验证

```bash
go test ./internal/serve/...
```

### 桌面前端静态验证

```bash
cd desktop/frontend
pnpm typecheck
```

如果这两项都过，至少说明保留的核心自定义能力没有明显破坏。

---

## 同步后必须输出的汇报格式

每次同步完成后，必须按下面结构汇报。

### 1. 同步结果

- 是否已成功把 `origin/main-v2` 同步进新的整理结果
- 如果用了重建分支：整理分支名是什么、是否已替换正式 `lianghong-use`
- 如果用了直接 merge：merge commit 是什么
- 当前分支是否还有未提交改动

### 2. 已被上游覆盖并丢弃的内容

明确列出：

- 哪些原本属于 `lianghong-use` 的功能，现在已经由上游覆盖
- 因此本次合并直接舍弃了哪些实现

### 3. 仍保留的自定义代码

必须列出：

- 还剩多少个文件
- 每个文件各自负责什么功能

推荐命令：

```bash
git diff --name-status origin/main-v2..lianghong-use
git diff --stat origin/main-v2..lianghong-use
```

### 4. 验证结果

明确列出执行过哪些验证，例如：

- `go test ./internal/serve/...`
- `pnpm typecheck`

### 5. 风险说明

如果还有高冲突区域，要明确指出：

- 哪个文件仍然是后续合并风险点
- 为什么
- 后续建议怎么继续压缩

---

## 推荐的最终结论模板

下面这个模板可直接复用：

```md
本次已将 `origin/main-v2` 同步进 `lianghong-use`。

上游已覆盖并因此舍弃的内容：
- ...
- ...

当前相对 `origin/main-v2` 仍保留的自定义代码共 N 个文件：
- `path/to/file-a`: ...
- `path/to/file-b`: ...

本次验证：
- `go test ./internal/serve/...`
- `pnpm typecheck`

后续主要维护风险点：
- `...`
原因：...
```

---

## 当前 `lianghong-use` 的建议认知

在本文档编写时，`lianghong-use` 相对 `origin/main-v2` 的自定义代码已被收敛到很小范围，重点只剩：

1. `serve` 端项目切换能力
2. workspace 列表持久化
3. 一处桌面端 transcript 流结束后的补滚动修复

这意味着以后新增需求时，建议继续遵守下面原则：

- 新需求优先做成独立文件
- 如果只是局部 bugfix，不要做全局重写
- 每次同步都以“减少相对上游的残留代码”为目标
- 每次同步优先重建干净分支，而不是继续堆旧 merge 历史

---

## 最后原则

**`lianghong-use` 只保留“上游没有、但我们确实需要”的最小代码。**

不是为了保住自己的实现而保留代码，而是为了保住自己的需求。

---

## 自定义 CSS / 前端改动记录

以下列出 `lianghong-use` 分支上独立于上游的前端样式或行为改动，合并上游时需注意保留。

### `.session-item` 移除 `overflow: hidden`

- **文件**: `internal/serve/index.html`（第 257 行附近）
- **原始 CSS**: `.session-item { ... overflow: hidden; }`
- **改动**: 去掉 `overflow: hidden`
- **原因**: 该属性导致 serve web 界面的会话列表显示异常（列表项内容被裁切，无法正常展示完整信息）。
- **合并注意事项**: 上游 `main-v2` 合并进来时，如果上游恢复了 `overflow: hidden`，需重新移除。
