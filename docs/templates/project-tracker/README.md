# Project Tracker 存储目录说明

该目录用于存放 Reasonix 的“按项目隔离的特性追踪数据”（Project Feature Tracker）。

默认位置建议为：

```text
~/.reasonix/project-tracker
```

如果设置了 `REASONIX_HOME`，则对应位置为：

```text
$REASONIX_HOME/project-tracker
```

该目录中的数据**默认属于 Reasonix 自己维护的工作台账**，而不是项目源码仓库本身的一部分。

---

## 推荐目录结构

```text
project-tracker/
  config.toml
  README.md
  projects/
    reasonix/
      project.toml
      features.toml
      state.toml
```

---

## 文件职责

### `config.toml`

用于记录 project tracker 的配置，例如：

```toml
base_dirs = ["~/Projects"]
```

可选地，也可以在注释或文档中说明：

```toml
# storage_dir = "~/.reasonix/project-tracker"
```

这主要用于帮助用户理解数据实际存放位置，而不一定要求运行时必须依赖此字段解析。

### `project.toml`

记录项目的稳定身份信息，例如：

```toml
id = "reasonix"
name = "reasonix"
root = "/home/lee/Projects/reasonix"
base_dir = "/home/lee/Projects"
```

### `features.toml`

记录该项目的特性清单，是最核心、最值得版本化的文件。

### `state.toml`

记录高频变化的运行态信息，例如：

- `created_at`
- `updated_at`
- `last_opened_at`

该文件默认更适合作为**非版本化状态文件**使用。

---

## 版本管理建议

推荐纳入版本管理：

- `config.toml`
- `project.toml`
- `features.toml`

推荐排除版本管理：

- `state.toml`

原因：

- `features.toml` 属于长期维护的知识资产
- `state.toml` 更像缓存 / 最近状态，容易产生低价值 diff

---

## 为什么默认不放项目目录

不推荐默认把这些数据直接写入项目目录，例如：

```text
<project>/.reasonix/features.toml
```

主要原因：

1. 会污染项目仓库
2. 会制造与上游同步无关的 diff
3. 这些数据本质上是 Reasonix 自己维护的工作台账

如果未来需要共享模式，可以在此默认方案之外扩展“项目内存储模式”。

