# Project Feature Tracker 设计草案

本文档定义 Reasonix 的一个拟新增能力：为“当前打开的项目目录”维护独立的特性清单（feature tracker），并支持在 Reasonix Web UI 中随时记录、查看和继续推进。

该能力的目标不是替代项目源码内的文档，而是提供一个**由 Reasonix 自己维护的、按项目隔离的特性台账系统**，用于记录：

1. 希望以后支持上的功能特性
2. 已经由 fork / 本地自定义实现支持上的功能特性
3. 已被上游覆盖并因此删除自定义实现的功能特性
4. 暂时搁置或不再计划支持的功能特性

---

## 设计目标

### 1. 按项目隔离

每个项目目录有自己的独立 feature list。

例如：

- `~/Projects/reasonix`
- `~/Projects/reasonix/plugins/example`
- `~/Projects/yak/server`

都应当被视为不同项目，分别维护自己的特性清单。

### 2. 由 Reasonix 自己维护

该数据属于 Reasonix 的工作台账，不默认属于项目源码仓库的一部分。

因此：

- 默认不直接写入项目目录
- 默认写入 `REASONIX_HOME` / `~/.reasonix` 下的专用存储目录

### 3. 适合版本化

稳定、人工维护价值高的内容应尽量使用文本格式存储，并支持纳入 home 目录版本管理。

### 4. 高频状态与稳定内容分离

例如：

- `created_at`
- `updated_at`
- `last_opened_at`

这类高频变化字段不应污染主 feature 清单的版本历史，应拆到独立状态文件。

### 5. Web UI 易于接入

未来应能以类似“快捷回复”的轻量方式接入 Reasonix Web UI：

- 查看当前项目 feature list
- 优先显示 wishlist
- 新增 / 编辑 / 改状态 / 删除

---

## 存储位置

默认存储根目录：

```text
~/.reasonix/project-tracker
```

如果设置了 `REASONIX_HOME`，则应使用：

```text
$REASONIX_HOME/project-tracker
```

推荐把这个目录视为 **Reasonix 的项目特性台账存储区**。

---

## 为什么不默认写入项目目录

不推荐默认使用如下形式：

```text
<project>/.reasonix/features.toml
```

原因：

1. 会污染项目仓库
2. 会制造与上游同步无关的 diff
3. 这些内容本质上是“agent 维护数据”，不是项目源码的一部分
4. 多项目统一查看、统计和迁移更困难

如果未来确实需要“项目内共享模式”，可以作为一个可选策略，但不应作为默认方案。

---

## 基础目录与项目 ID 规则

### 基本思路

项目逻辑 ID 不采用哈希，改为：

**相对于某个基础目录（base dir）的相对路径**

默认基础目录：

```text
~/Projects
```

并支持未来配置多个基础目录。

例如：

- 项目根目录：`/home/lee/Projects/reasonix`
- 基础目录：`/home/lee/Projects`
- 逻辑 ID：`reasonix`

再例如：

- 项目根目录：`/home/lee/Projects/reasonix/plugins/example`
- 基础目录：`/home/lee/Projects`
- 逻辑 ID：`reasonix/plugins/example`

### 为什么保留层级目录而不拍平

推荐直接保留层级目录，不把路径分隔符替换成 `_`。

推荐形式：

```text
projects/reasonix/
projects/reasonix/plugins/example/
projects/yak/server/
```

而不是：

```text
projects/reasonix_plugins_example/
```

原因：

1. 层级结构天然表达“子项目 / 子仓库”
2. 避免 `_` 编码冲突
3. 更易人工查看和维护
4. 与逻辑 ID 一致，少一层转换

### 多个基础目录

建议支持多个基础目录：

```toml
base_dirs = ["~/Projects", "~/Work", "~/Forks"]
```

解析规则建议为：

1. 将当前项目绝对路径与 `base_dirs` 逐个比较
2. 使用**最长前缀匹配**
3. 取相对路径作为项目逻辑 ID

### 不在任何基础目录下的项目

第一版建议允许 fallback。

可选策略：

1. 使用规范化绝对路径的编码结果作为临时 ID
2. 或提醒用户手动新增 base dir

为了避免阻塞使用，第一版更推荐：

- 先 fallback
- 后续再补“迁移到新 base dir”的能力

---

## 推荐目录结构

```text
~/.reasonix/project-tracker/
  config.toml
  README.md
  projects/
    reasonix/
      project.toml
      features.toml
      state.toml
    reasonix/plugins/example/
      project.toml
      features.toml
      state.toml
    yak/server/
      project.toml
      features.toml
      state.toml
```

---

## 文件职责划分

### 1. `config.toml`

作用：

- 记录 feature tracker 自己的配置
- 主要是 `base_dirs`

示例：

```toml
base_dirs = ["~/Projects"]
```

后续可扩展：

```toml
base_dirs = ["~/Projects", "~/Work"]
storage_dir = "~/.reasonix/project-tracker"
```

其中 `storage_dir` 更像文档化配置项；若实际运行中存储目录已由 `REASONIX_HOME` 决定，则不一定需要在代码里真正作为用户可改项暴露，但应在文档中说明。

### 2. `project.toml`

作用：

- 记录项目身份信息
- 只放相对稳定内容

建议字段：

```toml
id = "reasonix/plugins/example"
name = "example"
root = "/home/lee/Projects/reasonix/plugins/example"
base_dir = "/home/lee/Projects"
```

说明：

- `id` 是逻辑 ID
- `name` 是展示名
- `root` 是解析时识别到的项目绝对路径
- `base_dir` 用于调试、迁移和核对

不建议在此文件中放高频变化时间戳。

### 3. `features.toml`

作用：

- 记录真正有长期价值的 feature 台账
- 适合版本化
- 应是最核心的数据文件

建议字段：

```toml
[[features]]
id = "quick-reply"
title = "Web UI 快捷回复"
status = "custom_active"
priority = "high"
summary = "提供快捷回复插入、管理和本地自动发送偏好"
notes = "由 fork 维护；未来如上游出现等价模板插入能力，再评估是否删除自定义实现"

[[features]]
id = "project-feature-tracker"
title = "项目特性追踪列表"
status = "wishlist"
priority = "high"
summary = "按项目记录 wishlist、当前自维护特性、已被上游覆盖特性"
notes = "计划接入 Reasonix Web UI"
```

### 4. `state.toml`

作用：

- 只存高频变化、运行态数据
- 默认不建议纳入版本管理

建议字段：

```toml
created_at = "2026-07-06T12:00:00+08:00"
updated_at = "2026-07-06T12:30:00+08:00"
last_opened_at = "2026-07-06T13:10:00+08:00"
```

未来也可以加入：

- `last_selected_status`
- `last_selected_feature_id`
- `last_ui_filter`

---

## Feature 状态模型

建议固定为以下枚举：

- `wishlist`
- `planned`
- `in_progress`
- `custom_active`
- `upstream_covered`
- `dropped`

### 含义说明

#### `wishlist`

希望未来支持，但还未排期或尚未开始实现。

#### `planned`

已决定要做，但还未开始。

#### `in_progress`

正在实现中。

#### `custom_active`

当前已由 fork / 本地自定义实现支持，并仍在维护。

#### `upstream_covered`

曾经需要自定义支持，但现在已被上游覆盖，并已删除自定义实现。

#### `dropped`

不再计划支持，或确认不值得继续投入。

### 默认显示顺序

Web UI 中建议优先展示顺序：

1. `wishlist`
2. `planned`
3. `in_progress`
4. `custom_active`
5. `upstream_covered`
6. `dropped`

其中 `wishlist` 应当最优先显示，以方便随手记录未来想法。

---

## 版本化建议

推荐将以下文件纳入版本管理：

- `config.toml`
- `project.toml`
- `features.toml`

推荐将以下文件排除在版本管理之外：

- `state.toml`

原因：

- `features.toml` 是真正需要审阅和追踪的知识资产
- `state.toml` 是运行缓存和 UI 状态，不值得制造高频 diff

---

## Web UI 第一版接入建议

建议采用与快捷回复类似的“轻量入口 + 弹窗管理”模式，但目标更偏向“项目特性工作台”。

第一版建议支持：

1. 打开当前项目的 feature list
2. 优先显示 `wishlist`
3. 按状态筛选
4. 新增 feature
5. 编辑 feature
6. 改 feature 状态
7. 删除 feature

第一版不建议做：

- 跨项目聚合视图
- 父项目与子项目继承关系
- 多用户共享
- 数据库存储

先把“当前项目独立台账”做好即可。

---

## 第一版实现边界建议

为了控制复杂度，第一版建议：

1. 使用 TOML
2. 不用数据库
3. 不做复杂索引
4. 不做自动推导“已被上游覆盖”
5. 先让用户手工维护状态

换句话说，第一版重点是：

- 有地方存
- 有地方看
- 有地方改
- 结构足够稳定

而不是一开始就做复杂自动化。

---

## 与当前 `lianghong-use` 维护方式的关系

该能力与 `lianghong-use` 的上游同步策略是互补关系。

- `LIANGHONG-USE-UPSTREAM_MERGE_GUIDE.zh-CN.md`
  负责指导“如何同步上游、如何最小化残留 diff”
- Project Feature Tracker
  负责记录“到底有哪些需求、哪些仍需 fork 维护、哪些已被上游覆盖”

两者结合后：

1. 平时在 Web UI 中记录和维护项目特性
2. 与上游同步时，用 feature tracker 审计哪些特性仍需重放
3. 同步结束后，把对应 feature 状态更新为：
   - `custom_active`
   - `upstream_covered`
   - `dropped`

---

## 后续可扩展方向

后续如果第一版稳定，可考虑增加：

1. 跨项目总览
2. “最近更新的 feature”
3. `upstream_covered` 的证据字段
4. `related_files`
5. `related_commits`
6. `owner` / `labels`
7. 与任务 / goal 模式联动

但这些都应放在第一版之后。

