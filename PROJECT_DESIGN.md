# OJ（Online Judge）系统设计文档

## 1. 项目概述

### 1.1 项目目标
开发一个支持 AI 智能判题的在线评测系统（OJ），在传统 OJ 功能基础上，增加基于大模型的代码分析能力，能够检测用户提交的代码是否符合题目要求的算法或编程语言规范。

### 1.2 核心特性
- **传统 OJ 功能**：用户登录与管理、题目管理、代码提交、自动评测、排行榜等
- **比赛功能**：支持 OI / IOI 赛制，按题目与用户/分组配置参赛范围，支持固定起止与窗口期+个人时长两种计时模式
- **比赛提交上限**：OI / IOI 统一为单场比赛每位用户最多提交 `99` 次
- **权限分层**：`super_admin` 与 `admin` 均可进行题目/比赛/系统编辑，仅 `super_admin` 可调整用户管理员身份
- **AI 智能判题**：调用大模型 API（如 DeepSeek）分析代码，检查是否使用指定算法/语言
- **灵活配置**：每道题目可独立配置是否启用 AI 判题及具体要求
- **双重结果**：同时返回测试点评测结果和 AI 分析结果
- **自助修改密码**：用户可在个人中心修改密码
- **文件操作题目**：可要求从指定文件读入并输出到指定文件
- **文件 IO 评测稳态**：文件读写题在写入输入文件前会先创建工作目录，避免目录缺失导致 `System Error`
- **评测性能优化**：编译型语言在单次提交中仅编译一次，按测试点复用运行
- **隐藏题可见性**：`admin/super_admin` 可直接查看并提交；普通参赛用户在固定起止开赛后或窗口期开始个人会话后可见，比赛结束后参赛用户仍可访问
- **UI 体系升级**：采用 Swiss 风格重构，统一状态色、表格对齐规范与后台编辑体验
- **OI 比赛遮罩机制**：进行中的 OI 比赛对普通用户隐藏详细评测信息，仅显示 `Submitted`
- **OI 成绩即时可见**：固定起止比赛结束后立即显示总分；窗口赛在用户个人时长结束后立即显示总分
- **窗口期开赛确认**：窗口期比赛点击“开始比赛”后需二次确认，确认后才进入个人计时
- **IOI 提交计数展示**：比赛详情页显示“已提交次数/提交上限（99）”
- **比赛榜单题号规则**：管理员比赛成绩表与导出 CSV 的题目列按比赛内顺序显示 `A/B/C/...`（每场从 `A` 开始）
- **帮助页 Markdown 渲染**：`/help` 页面正文改为 Markdown 渲染输出，确保反引号/表格等语法可正确展示
- **评测中断与删除**：管理员可在提交列表终止指定评测，或删除指定提交记录
- **考试压测工具链**：提供 `stress-testing/` 工作流与脚本，支持批量建号建赛、提交延迟与服务器压力采集

### 1.3 部署环境
- **开发环境**：Windows 主机 / WSL
- **生产环境**：2核2G 小型服务器（公网访问）

---

## 2. 技术选型

考虑到服务器资源有限（2核2G），选择轻量级技术栈：

### 2.1 后端
| 组件 | 技术选型 | 说明 |
|------|---------|------|
| Web 框架 | **Go + Gin** | 当前实现 |
| 数据库 | **SQLite + GORM** | 当前实现 |
| 判题队列 | **内存队列（channel）** | `judge/queue` |
| 定时维护 | **后台协程 + ticker** | 赛后统计同步、启动全量修复 |

### 2.2 前端
| 组件 | 技术选型 | 说明 |
|------|---------|------|
| 框架 | **Vue 3 + Vite** | 当前实现 |
| UI 库 | **Element Plus** | 当前实现 |
| 代码编辑器 | **Monaco Editor** | 题目页与提交详情使用 |
| 布局组件 | **splitpanes** | 题面/代码分屏 |

### 2.3 判题沙箱
| 组件 | 技术选型 | 说明 |
|------|---------|------|
| 沙箱 | **isolate** 或 **Docker** | isolate 更轻量，Docker 更通用 |
| 编译器 | GCC, G++, Python, Java 等 | 按需安装 |

### 2.4 AI 接口
| 组件 | 技术选型 | 说明 |
|------|---------|------|
| 大模型 | **DeepSeek API** | 成本低、支持中文、代码理解能力强 |
| 调用方式 | HTTP REST API | 异步调用，避免阻塞判题流程 |

### 2.5 推荐方案（资源受限场景）
```
后端：Go + Gin + SQLite
前端：Vue 3 + Element Plus（构建为静态文件）
判题：isolate 沙箱（Linux）
AI：DeepSeek API
部署：单机部署，Nginx 反向代理
```

---

## 3. 系统架构

### 3.1 整体架构图
```
┌─────────────────────────────────────────────────────────────────┐
│                          用户浏览器                               │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Nginx 反向代理                            │
│              (静态文件服务 + API 转发 + HTTPS)                    │
└─────────────────────────────────────────────────────────────────┘
                                │
                ┌───────────────┴───────────────┐
                ▼                               ▼
┌───────────────────────────┐    ┌───────────────────────────────┐
│     前端静态文件            │    │         后端 API 服务          │
│     (Vue 3 构建产物)        │    │         (Go/Gin)              │
└───────────────────────────┘    └───────────────────────────────┘
                                                │
                ┌───────────────┬───────────────┼───────────────┐
                ▼               ▼               ▼               ▼
        ┌───────────┐   ┌───────────┐   ┌───────────┐   ┌───────────┐
        │  SQLite   │   │ 判题队列   │   │ 判题沙箱   │   │ DeepSeek  │
        │  数据库    │   │  (内存)   │   │ (isolate) │   │   API     │
        └───────────┘   └───────────┘   └───────────┘   └───────────┘
```

### 3.2 判题流程
```
用户提交代码
      │
      ▼
┌─────────────────┐
│ 1. 保存提交记录  │
│ 状态: Pending   │
└─────────────────┘
      │
      ▼
┌─────────────────┐
│ 2. 加入判题队列  │
└─────────────────┘
      │
      ▼
┌─────────────────┐     ┌─────────────────────────────────────┐
│ 3. 判题 Worker  │────▶│ 3.1 传统评测（单次编译 + 运行测试点）│
│    取出任务     │     └─────────────────────────────────────┘
└─────────────────┘                    │
      │                                ▼
      │                 ┌─────────────────────────────────────┐
      │                 │ 3.2 AI 评测（如果题目启用）          │
      │                 │     - 调用 DeepSeek API             │
      │                 │     - 分析算法/语言是否符合要求      │
      │                 └─────────────────────────────────────┘
      │                                │
      ▼                                ▼
┌─────────────────────────────────────────────────────────────┐
│ 4. 综合结果                                                  │
│    - 传统评测未通过 → 返回传统评测结果                        │
│    - 传统评测通过且 AI 未通过：严格模式 -> WA；非严格仅提示    │
│    - AI 未通过时最终分数封顶 50                               │
└─────────────────────────────────────────────────────────────┘
      │
      ▼
┌─────────────────┐
│ 5. 更新数据库   │
│ 通知用户结果    │
└─────────────────┘
```

---

## 4. 模块设计

### 4.1 模块划分
```
oj-system/
├── backend/                 # 后端服务
│   ├── cmd/                 # 入口
│   │   └── server/
│   │       └── main.go
│   ├── configs/             # 配置文件
│   ├── internal/
│   │   ├── config/          # 配置管理
│   │   ├── handler/         # HTTP 处理器
│   │   ├── service/         # 业务逻辑
│   │   ├── repository/      # 数据访问
│   │   ├── model/           # 数据模型
│   │   ├── middleware/      # 中间件
│   │   ├── judge/           # 判题核心
│   │   │   ├── sandbox/     # 沙箱执行
│   │   │   ├── ai/          # AI 判题
│   │   │   └── queue/       # 判题队列
│   │   └── utils/           # 工具函数
│   └── go.mod
│
├── frontend/                # 前端项目
│   ├── src/
│   │   ├── api/             # API 调用
│   │   ├── components/      # 通用组件
│   │   ├── views/           # 页面
│   │   ├── stores/          # 状态管理
│   │   ├── router/          # 路由
│   │   ├── utils/           # 工具函数
│   │   └── assets/          # 静态资源
│   └── package.json
│
├── deploy/                  # 部署相关
│   ├── nginx/
│   ├── systemd/
│   └── scripts/
│
├── stress-testing/          # 压力测试工具与工作流
│   ├── WORKFLOW.md          # 考试场景压测步骤
│   ├── prepare_exam.py      # 批量建号/按组建赛
│   ├── run_exam_load.py     # 提交压测主脚本
│   ├── collect_server_metrics.py # 服务器指标采集
│   ├── analyze_results.py   # 压测结果分析
│   └── requirements.txt
│
├── data/                    # 数据目录（运行时）
│   ├── problems/            # 题目数据（测试点）
│   ├── submissions/         # 提交代码
│   └── db/                  # 数据库文件
│
└── docs/                    # 文档
```

---

## 5. API 接口设计

### 5.1 用户模块 `/api/v1/user`

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| POST | `/login` | 用户登录 | 公开 |
| GET | `/profile` | 获取个人信息 | 登录 |
| PUT | `/profile` | 更新个人信息 | 登录 |
| PUT | `/password` | 修改密码 | 登录 |

**账号分配说明**
普通用户账号由管理员在后台统一创建与分配，客户端不再提供注册入口。

**登录响应**
```json
{
    "code": 200,
    "data": {
        "token": "jwt_token_string",
        "user": {
            "id": 1,
            "username": "string",
            "role": "user|admin|super_admin"
        }
    }
}
```

### 5.2 题目模块 `/api/v1/problem`

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | `/list` | 获取题目列表 | 公开（默认仅公开题，管理员可见全部） |
| GET | `/:id` | 获取题目详情 | 公开（隐藏题需权限） |
| POST | `/` | 创建题目 | 管理员 |
| PUT | `/:id` | 更新题目 | 管理员 |
| DELETE | `/:id` | 删除题目 | 管理员 |
| POST | `/:id/image` | 上传题面图片 | 管理员 |
| GET | `/:id/image/:filename` | 获取题面图片 | 公开（受题目可见性约束） |
| POST | `/:id/testcase` | 上传测试数据 | 管理员 |
| POST | `/:id/testcase/zip` | Zip 批量上传测试数据 | 管理员 |
| POST | `/:id/rejudge` | 整题重测（历史提交重新入队） | 管理员 |
| GET | `/:id/testcases` | 获取测试点列表 | 管理员 |
| DELETE | `/:id/testcases` | 清空测试点 | 管理员 |

**隐藏题可见性**：隐藏题仅对管理员（`admin/super_admin`）或比赛开始后的参赛用户可见，赛后参赛用户仍可访问；进行中的比赛中，非管理员访问隐藏题时不展示题目标签与难度。

**上传链路参数（当前实现）**：
- Nginx：`client_max_body_size 200m`，`proxy_send_timeout/proxy_read_timeout = 600s`
- 后端 Gin：`MaxMultipartMemory = 256MB`
- 前端：普通请求超时 `180s`，测试点上传超时 `600s` 并展示上传进度
- 题面图片上传：单图大小上限 `10MB`，支持 `png/jpg/jpeg/gif/webp/bmp`

**题目数据结构**
```json
{
    "id": 1,
    "title": "两数之和",
    "description": "题目描述（支持 Markdown）",
    "input_format": "输入格式说明",
    "output_format": "输出格式说明",
    "hint": "提示（可选）",
    "samples": [
        {
            "input": "1 2\n3 4",
            "output": "3\n7"
        }
    ],
    "time_limit": 1000,           // 时间限制，单位 ms
    "memory_limit": 256,          // 内存限制，单位 MB
    "difficulty": "easy|medium|hard",
    "tags": ["数组", "哈希表"],

    // ========== 文件操作 ==========
    "file_io_enabled": true,      // 是否启用文件操作
    "file_input_name": "data.in", // 输入文件名
    "file_output_name": "data.out", // 输出文件名
    
    // ========== AI 判题配置 ==========
    "ai_judge_config": {
        "enabled": true,                    // 是否启用 AI 判题
        "required_algorithm": "动态规划",    // 要求使用的算法（可选）
        "required_language": "C++",         // 要求使用的语言（可选）
        "forbidden_features": ["STL sort"], // 禁止使用的特性（可选）
        "custom_prompt": "额外的判题说明"    // 自定义 prompt（可选）
    }
}
```

### 5.3 提交模块 `/api/v1/submission`

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| POST | `/` | 提交代码 | 登录 |
| GET | `/:id` | 获取提交详情 | 登录 |
| GET | `/list` | 获取提交列表 | 登录 |
| GET | `/my` | 获取我的提交 | 登录 |

**说明**：
- 管理员可查看所有提交；普通用户仅可查看自己的提交。
- 隐藏题提交权限：管理员（`admin/super_admin`）可直接提交；普通用户需满足隐藏题可见性规则后才可提交。
- 进行中的 OI 比赛中，普通用户看到的相关提交会被遮罩为 `Submitted`，分数与详细评测信息隐藏。
- `memory_used` 按程序运行期间虚拟内存峰值（`VmPeak`）统计（KB）；超过 `memory_limit`（MB）返回 `Memory Limit Exceeded`。
- Linux 运行前会设置 `ulimit -v` 与 `ulimit -s` 为 `memory_limit * 1024`（KB）。

**提交请求**
```json
POST /api/v1/submission
{
    "problem_id": 1,
    "language": "cpp",          // cpp, c, python, java, go
    "code": "源代码内容"
}
```

**提交结果（带 AI 判题）**
```json
{
    "id": 12345,
    "problem_id": 1,
    "user_id": 1,
    "language": "cpp",
    "status": "Wrong Answer",   // Accepted, Wrong Answer, TLE, MLE, RE, CE, Pending, Judging, Submitted
    "time_used": 15,            // 最大运行时间 ms
    "memory_used": 1024,        // 最大内存使用 KB
    "created_at": "2025-01-28T10:00:00Z",
    
    // 测试点详情
    "testcase_results": [
        {"id": 1, "status": "Accepted", "time": 10, "memory": 512},
        {"id": 2, "status": "Accepted", "time": 15, "memory": 1024},
        {"id": 3, "status": "Accepted", "time": 12, "memory": 800}
    ],
    
    // ========== AI 判题结果 ==========
    "ai_judge_result": {
        "enabled": true,
        "passed": false,                     // AI 判定是否通过
        "algorithm_detected": "暴力枚举",     // 检测到的算法
        "language_check": "passed",          // 语言检查结果
        "reason": "题目要求使用动态规划算法，但检测到代码使用了暴力枚举方法。",
        "details": {
            "required": "动态规划",
            "detected": "暴力枚举",
            "confidence": 0.92
        }
    },
    
    // 最终判定说明
    "final_message": "测试点全部通过，但未使用要求的算法，判定为 Wrong Answer"
}
```

### 5.4 排行榜模块 `/api/v1/rank`

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | `/` | 获取排行榜 | 公开 |

### 5.5 比赛模块 `/api/v1/contest`

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | `/list` | 获取比赛列表 | 登录 |
| GET | `/:id` | 获取比赛详情 | 登录 |
| POST | `/:id/start` | 开始窗口期个人比赛 | 登录 |

**比赛详情响应结构**
```json
{
    "contest": {
        "id": 1,
        "title": "期中赛",
        "type": "oi",
        "timing_mode": "window",
        "duration_minutes": 180,
        "submission_limit": 99,
        "start_at": "2026-03-01T08:00:00Z",
        "end_at": "2026-03-01T11:00:00Z",
        "problem_ids": [1, 2, 3],
        "allowed_users": [10, 11],
        "allowed_groups": ["ClassA"]
    },
    "problems": [
        {"id": 1, "title": "A+B", "difficulty": "easy", "has_accepted": true}
    ],
    "session": {
        "started": true,
        "can_start": false,
        "in_live": true,
        "start_at": "2026-03-01T08:30:00Z",
        "end_at": "2026-03-01T11:30:00Z",
        "remaining_seconds": 5400
    },
    "my_live_total": 180,
    "my_post_total": 180,
    "my_submission_count": 7
}
```

**说明**：
- `has_accepted` 仅在以下情况展示：管理员、IOI 赛制、或比赛已结束
- `has_submitted` 表示是否在比赛时间范围内提交过该题
- `timing_mode` 支持 `fixed`（固定起止）与 `window`（窗口期 + 个人固定时长）
- `submission_limit` 为系统固定值 `99`（每位用户每场比赛最多提交 99 次），OI / IOI 均生效
- 窗口期比赛中，用户点击“开始比赛”后需先确认，再调用 `POST /contest/:id/start` 启动个人计时
- `my_live_total` / `my_post_total` 分别表示赛时与赛后得分（赛后为订正总分口径，包含赛时基线）
- `my_submission_count` 为当前用户在该比赛“赛时阶段”的已提交次数（用于 IOI 详情页展示 `已提交/上限`）
- 管理员排行榜支持 `board_mode=combined|live|post` 三种视图
- 管理员排行榜列标题按 `A/B/C/...`（列序对应 `problem_ids` 顺序）展示，不直接显示题库题号
- 窗口期比赛中，管理员排行榜在比赛结束前显示用户“剩余时间/未开始”，比赛结束后不显示该列
- 窗口期比赛中，管理员可重置指定用户的开赛会话，用户状态恢复为“未开始”
- 管理员可终止指定用户比赛（强制交卷）：窗口期会立即截断该用户会话结束时间；固定起止会写入该用户个人结束时间，后续提交按赛后口径处理

### 5.6 统计模块 `/api/v1/statistics`

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | `/` | 获取系统统计（题目/用户/提交数量） | 公开 |

**统计响应**
```json
{
    "problems": 10,
    "users": 30,
    "submissions": 120
}
```

### 5.7 管理模块 `/api/v1/admin`

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | `/users` | 用户管理列表 | 管理员 |
| POST | `/users` | 创建用户（管理员分配账号） | 管理员 |
| POST | `/users/batch` | 批量导入用户 | 管理员 |
| PUT | `/users/:id` | 更新用户信息 | 管理员 |
| PUT | `/users/:id/role` | 修改用户角色 | 超级管理员 |
| POST | `/submissions/:id/abort` | 终止指定提交评测 | 管理员 |
| DELETE | `/submissions/:id` | 删除指定提交记录 | 管理员 |
| POST | `/contests` | 创建比赛 | 管理员 |
| PUT | `/contests/:id` | 更新比赛 | 管理员 |
| DELETE | `/contests/:id` | 删除比赛 | 管理员 |
| POST | `/contests/:id/users/:user_id/reset-start` | 重置用户窗口期开赛状态 | 管理员 |
| POST | `/contests/:id/users/:user_id/force-finish` | 终止用户比赛（强制交卷） | 管理员 |
| GET | `/contests/:id/leaderboard` | 比赛排行榜（管理员，支持 `board_mode`） | 管理员 |
| GET | `/contests/:id/export` | 导出比赛成绩（支持 `board_mode`） | 管理员 |
| POST | `/contests/:id/refresh` | 刷新比赛统计（赛后同步） | 管理员 |
| GET | `/settings/ai` | 获取 AI 设置 | 管理员 |
| PUT | `/settings/ai` | 更新 AI 设置 | 管理员 |
| POST | `/settings/ai/test` | 测试 AI 连接 | 管理员 |

**创建用户请求**
```json
POST /api/v1/admin/users
{
    "username": "string",      // 用户名，3-20 字符
    "email": "string",         // 邮箱（可选）
    "password": "string",      // 初始密码，6-20 字符
    "student_id": "string",    // 学号（可选）
    "group": "ClassA"          // 分组（可选）
}
```

**批量导入请求**
```json
POST /api/v1/admin/users/batch
{
    "users": [
        {
            "username": "student01",
            "password": "pass123",
            "email": "",
            "student_id": "2025001",
            "group": "ClassA"
        }
    ]
}
```

**更新用户请求**
```json
PUT /api/v1/admin/users/:id
{
    "username": "string",      // 可选
    "email": "string",         // 可选（可置空）
    "student_id": "string",    // 可选（可置空）
    "group": "ClassA",         // 可选（可置空）
    "role": "user",            // 可选（仅超级管理员可修改）
    "password": "string"       // 可选（重置密码）
}
```

**创建比赛请求**
```json
POST /api/v1/admin/contests
{
    "title": "期中赛",
    "description": "可选",
    "type": "oi",                         // oi 或 ioi
    "timing_mode": "window",              // fixed 或 window（可选，默认 fixed）
    "duration_minutes": 180,              // timing_mode=window 时必填，单位分钟
    "start_at": "2026-03-01T08:00:00Z",
    "end_at": "2026-03-01T11:00:00Z",
    "problem_ids": [1, 2, 3],
    "allowed_users": [10, 11],
    "allowed_groups": ["ClassA"]
}
```

**说明**：`allowed_users` 与 `allowed_groups` 至少填写一个，否则普通用户无法看到/参加比赛。

---

## 6. AI 判题系统设计

### 6.1 DeepSeek API 调用

**请求示例**
```json
POST https://api.deepseek.com/v1/chat/completions
Headers:
  Authorization: Bearer <API_KEY>
  Content-Type: application/json

{
    "model": "deepseek-chat",
    "messages": [
        {
            "role": "system",
            "content": "你是一个代码分析专家。请分析用户提交的代码，判断其使用的算法和编程特性。请严格按照指定的 JSON 格式输出结果。"
        },
        {
            "role": "user",
            "content": "请分析以下代码：\n\n```cpp\n用户代码\n```\n\n题目要求：\n- 必须使用算法：动态规划\n- 允许的语言：C++\n\n请输出 JSON 格式的分析结果。"
        }
    ],
    "response_format": {"type": "json_object"},
    "temperature": 0.1
}
```

### 6.2 AI 返回格式（JSON Schema）

```json
{
    "algorithm_analysis": {
        "detected_algorithms": ["动态规划", "记忆化搜索"],
        "primary_algorithm": "动态规划",
        "confidence": 0.95,
        "evidence": "代码中使用了 dp 数组进行状态转移，第15-20行包含典型的动态规划递推公式"
    },
    "language_features": {
        "language": "C++",
        "standard": "C++11",
        "used_features": ["vector", "algorithm头文件"],
        "forbidden_features_used": []
    },
    "requirement_check": {
        "algorithm_match": true,
        "language_match": true,
        "all_requirements_met": true
    },
    "summary": "代码使用动态规划算法解决问题，符合题目要求"
}
```

### 6.3 Prompt 模板

```markdown
# 角色
你是一个专业的代码分析专家，擅长识别代码中使用的算法和编程技术。

# 任务
分析用户提交的代码，判断是否符合题目的特定要求。

# 题目信息
- 题目标题：{{problem_title}}
- 题目描述：{{problem_description}}

# 题目要求
{{#if required_algorithm}}
- 必须使用的算法：{{required_algorithm}}
{{/if}}
{{#if required_language}}
- 必须使用的编程语言：{{required_language}}
{{/if}}
{{#if forbidden_features}}
- 禁止使用的特性：{{forbidden_features}}
{{/if}}
{{#if custom_prompt}}
- 额外要求：{{custom_prompt}}
{{/if}}

# 用户提交的代码
```{{language}}
{{code}}
```

# 输出要求
请严格按照以下 JSON 格式输出分析结果，不要输出其他内容：

{
    "algorithm_analysis": {
        "detected_algorithms": ["检测到的算法列表"],
        "primary_algorithm": "主要使用的算法",
        "confidence": 0.0-1.0的置信度,
        "evidence": "判断依据说明"
    },
    "language_features": {
        "language": "编程语言",
        "used_features": ["使用的语言特性"],
        "forbidden_features_used": ["使用了的禁止特性"]
    },
    "requirement_check": {
        "algorithm_match": true/false,
        "language_match": true/false,
        "all_requirements_met": true/false
    },
    "summary": "一句话总结"
}
```

### 6.4 AI 判题配置项

在创建/编辑题目时，可配置以下 AI 判题选项：

| 配置项 | 类型 | 说明 |
|--------|------|------|
| `enabled` | bool | 是否启用 AI 判题 |
| `required_algorithm` | string | 要求使用的算法（如"动态规划"、"DFS"、"贪心"等） |
| `required_language` | string | 要求使用的语言 |
| `forbidden_features` | []string | 禁止使用的特性（如"STL sort"、"递归"等） |
| `custom_prompt` | string | 自定义判题说明 |
| `strict_mode` | bool | 严格模式：AI 判定失败则直接 WA |

### 6.5 评分修正规则

当题目启用 AI 判题且 **AI 判定未满足要求** 时，最终得分会进行封顶处理：

```
score = min(score, 50)
```

---

## 7. 数据库设计

### 7.1 用户表 `users`
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100),
    password_hash VARCHAR(255) NOT NULL,
    student_id VARCHAR(50),
    role VARCHAR(20) DEFAULT 'user',        -- user, admin
    `group` VARCHAR(50),
    solved_count INTEGER DEFAULT 0,
    accepted_count INTEGER DEFAULT 0,
    submit_count INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### 7.2 题目表 `problems`
```sql
CREATE TABLE problems (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    input_format TEXT,
    output_format TEXT,
    hint TEXT,
    samples TEXT,                           -- JSON 格式
    time_limit INTEGER DEFAULT 1000,        -- ms
    memory_limit INTEGER DEFAULT 256,       -- MB
    difficulty VARCHAR(20),
    tags TEXT,                              -- JSON 数组
    
    -- AI 判题配置
    ai_judge_config TEXT,                   -- JSON 格式
    file_io_enabled BOOLEAN DEFAULT FALSE,
    file_input_name VARCHAR(100),
    file_output_name VARCHAR(100),
    
    is_public BOOLEAN DEFAULT TRUE,
    created_by INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (created_by) REFERENCES users(id)
);
```

### 7.3 测试数据表 `testcases`
```sql
CREATE TABLE testcases (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    problem_id INTEGER NOT NULL,
    input_file VARCHAR(255) NOT NULL,       -- 输入文件路径
    output_file VARCHAR(255) NOT NULL,      -- 输出文件路径
    score INTEGER DEFAULT 0,                -- 该测试点分数
    is_sample BOOLEAN DEFAULT FALSE,
    order_num INTEGER DEFAULT 0,
    
    FOREIGN KEY (problem_id) REFERENCES problems(id) ON DELETE CASCADE
);
```

### 7.4 提交表 `submissions`
```sql
CREATE TABLE submissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    problem_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    language VARCHAR(20) NOT NULL,
    code TEXT NOT NULL,
    status VARCHAR(30) DEFAULT 'Pending',
    time_used INTEGER,                      -- ms
    memory_used INTEGER,                    -- KB
    score INTEGER DEFAULT 0,
    
    -- 测试点结果（JSON）
    testcase_results TEXT,
    
    -- AI 判题结果（JSON）
    ai_judge_result TEXT,
    
    -- 编译错误信息
    compile_error TEXT,
    final_message TEXT,
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (problem_id) REFERENCES problems(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_submissions_user ON submissions(user_id);
CREATE INDEX idx_submissions_problem ON submissions(problem_id);
CREATE INDEX idx_submissions_status ON submissions(status);
```

### 7.5 比赛表 `contests`
```sql
CREATE TABLE contests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    type VARCHAR(10) NOT NULL,            -- oi | ioi
    timing_mode VARCHAR(20) DEFAULT 'fixed', -- fixed | window
    duration_minutes INTEGER DEFAULT 0,    -- 个人比赛时长（分钟）
    submission_limit INTEGER DEFAULT 99,   -- 比赛总提交次数上限（固定 99）
    start_at DATETIME,
    end_at DATETIME,
    problem_ids TEXT,                     -- JSON 列表
    allowed_users TEXT,                   -- JSON 列表
    allowed_groups TEXT,                  -- JSON 列表
    is_stats_synced BOOLEAN DEFAULT FALSE,
    created_by INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### 7.6 比赛会话表 `contest_participations`
```sql
CREATE TABLE contest_participations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    contest_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    start_at DATETIME,
    end_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(contest_id, user_id)
);
```

---

## 8. 前端页面设计

### 8.1 页面列表

| 页面 | 路由 | 说明 |
|------|------|------|
| 首页 | `/` | 系统介绍、公告 |
| 题目列表 | `/problems` | 题目列表、筛选、搜索 |
| 题目详情 | `/problem/:id` | 题目内容、代码提交 |
| 比赛列表 | `/contests` | 比赛列表（需登录，不展示固定提交上限列） |
| 比赛详情 | `/contest/:id` | 比赛详情、题目列表（需登录；IOI 显示已提交/上限） |
| 提交列表 | `/submissions` | 提交记录（需登录；管理员可终止评测或删除记录） |
| 提交详情 | `/submission/:id` | 提交结果、AI 分析详情（需登录） |
| 排行榜 | `/rank` | 用户排名 |
| 评测帮助 | `/help` | 公示评测系统环境、编译命令、版本、赛制（OI/IOI、fixed/window）与资源限制规则 |
| 个人中心 | `/profile` | 个人信息、提交统计 |
| 登录 | `/login` | 用户登录 |
| 管理后台 | `/admin/*` | 题目管理、用户管理 |
| 比赛创建（管理） | `/admin/contest/create` | 创建比赛（描述双栏编辑预览） |
| 比赛编辑（管理） | `/admin/contest/:id/edit` | 编辑比赛（描述双栏编辑预览） |

### 8.2 核心组件

```
components/
├── common/
│   ├── Navbar.vue           # 导航栏
│   ├── Footer.vue           # 页脚
│   ├── MarkdownPreview.vue  # Markdown + LaTeX 渲染
│   └── CodeEditor.vue       # Monaco 编辑器封装
├── problem/
│   ├── StatusBadge.vue      # 题目状态标识
│   └── DifficultyBadge.vue  # 难度标签
├── submission/
│   ├── TestcaseResults.vue  # 测试点结果
│   └── AIJudgeResult.vue    # AI 判题结果展示
└── views/
    ├── problem/ProblemDetail.vue   # 题面/代码分屏
    ├── submission/SubmissionDetail.vue # 评测仪表盘
    ├── contest/ContestDetail.vue   # 比赛详情、窗口期开赛、IOI 提交计数与赛时/赛后榜切换
    ├── Help.vue                    # 帮助页（系统/命令/赛制/资源限制）
    ├── admin/ProblemEdit.vue       # 双栏 Markdown 编辑 + 题面图片上传 + 测试点管理（含上传进度、整题重测）
    └── admin/ContestEdit.vue       # 比赛描述双栏 Markdown 编辑与预览
```

---

## 9. 开发任务清单

### 阶段一：基础框架搭建（后端）
- [ ] 初始化 Go 项目，配置项目结构
- [ ] 配置管理模块（读取 YAML 配置）
- [ ] 数据库连接与 ORM 配置（GORM + SQLite）
- [ ] 用户认证模块（JWT）
- [ ] 基础中间件（日志、错误处理、CORS、认证）
- [ ] 用户模块 API（登录、个人信息、管理员创建用户）

### 阶段二：核心业务开发
- [ ] 题目模块 API（CRUD、测试数据上传）
- [ ] 提交模块 API（提交代码、查询结果）
- [ ] 判题队列实现
- [ ] 传统判题沙箱集成（isolate/Docker）
- [ ] 各语言编译运行支持（C/C++/Python/Java）

### 阶段三：AI 判题系统
- [ ] DeepSeek API 封装
- [ ] Prompt 模板管理
- [ ] AI 判题结果解析
- [ ] AI 判题与传统判题结果整合
- [ ] 异步判题流程优化

### 阶段四：前端开发
- [ ] 初始化 Vue 3 项目
- [ ] 路由配置与布局
- [ ] 用户认证页面（登录）
- [ ] 题目列表与详情页
- [ ] 代码编辑器集成（Monaco Editor）
- [ ] 提交与结果展示
- [ ] AI 判题结果展示组件
- [ ] 排行榜页面
- [ ] 管理后台页面

### 阶段五：测试与优化
- [ ] 单元测试编写
- [ ] 接口测试
- [ ] 性能测试与优化
- [ ] 安全性检查

### 阶段六：部署上线
- [ ] 编写部署脚本
- [ ] Nginx 配置
- [ ] HTTPS 证书配置
- [ ] 监控与日志配置
- [ ] 数据备份方案

---

## 10. 部署指南

### 10.1 服务器环境准备

以下命令在 Debian 12 / Ubuntu 20.04+ 适用。

```bash
# 1. 更新系统
sudo apt update && sudo apt upgrade -y

# 2. 安装必要软件
sudo apt install -y nginx sqlite3 git

# 3. 安装编译器（用于判题）
sudo apt install -y gcc g++ python3
# 如需 Java 判题再安装 JDK
sudo apt install -y default-jdk

# 4. 安装 isolate 沙箱（推荐）
sudo apt install -y isolate
# 或者安装 Docker（备选）
# curl -fsSL https://get.docker.com | sh

# 5. 创建应用目录
sudo mkdir -p /opt/oj
sudo chown $USER:$USER /opt/oj
```

### 10.2 后端部署

```bash
# 1. 构建后端（在开发机上）
cd backend
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o oj-server ./cmd/server

# 2. 上传到服务器
scp oj-server user@server:/opt/oj/
scp -r configs user@server:/opt/oj/

# 3. 创建 systemd 服务
sudo tee /etc/systemd/system/oj.service << EOF
[Unit]
Description=OJ Backend Service
After=network.target

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/opt/oj
ExecStart=/opt/oj/oj-server -config /opt/oj/configs/config.yaml
Restart=always
RestartSec=5

# 环境变量
EnvironmentFile=/opt/oj/.env

# 日志
StandardOutput=append:/var/log/oj/oj.log
StandardError=append:/var/log/oj/oj-error.log

# 资源限制（适用于小型服务器）
MemoryLimit=512M
CPUQuota=100%

[Install]
WantedBy=multi-user.target
EOF

# 4. 创建日志目录并启动服务
sudo mkdir -p /var/log/oj
sudo chown -R www-data:www-data /var/log/oj /opt/oj

sudo systemctl daemon-reload
sudo systemctl enable oj
sudo systemctl start oj
```

### 10.3 前端部署

```bash
# 1. 构建前端（在开发机上）
cd frontend
npm run build

# 2. 上传到服务器
scp -r dist/* user@server:/opt/oj/static/
```

### 10.4 Nginx 配置

```nginx
# /etc/nginx/sites-available/oj
server {
    listen 80;
    server_name your-domain.com;
    client_max_body_size 200m;
    
    # 重定向到 HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    # SSL 证书（使用 Let's Encrypt）
    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;
    
    # 前端静态文件
    location / {
        root /opt/oj/static;
        try_files $uri $uri/ /index.html;
        
        # 缓存静态资源
        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2)$ {
            expires 30d;
            add_header Cache-Control "public, immutable";
        }
    }
    
    # API 代理
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_send_timeout 600s;
        proxy_read_timeout 600s;
        
        # WebSocket 支持（如果需要实时推送）
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
    
    # 安全头
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
}
```

```bash
# 启用配置
sudo ln -s /etc/nginx/sites-available/oj /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx

# 获取 SSL 证书
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d your-domain.com
```

### 10.5 配置文件示例

```yaml
# configs/config.yaml
server:
  port: 8080
  mode: release              # debug, release

database:
  driver: sqlite
  path: /opt/oj/data/oj.db

judge:
  sandbox: isolate           # isolate, docker
  workers: 2                 # 判题并发数（根据 CPU 核心数设置）
  timeout: 30                # 单个判题最大时间（秒）
  
# AI 设置仅通过管理后台写入数据库读取，当前代码不会从 config.yaml 读取此段
ai:
  enabled: false
  provider: deepseek
  api_key: ""
  api_url: https://api.deepseek.com/v1/chat/completions
  model: deepseek-chat
  timeout: 60
  
paths:
  problems: /opt/oj/data/problems
  submissions: /opt/oj/data/submissions
  
jwt:
  secret: ${JWT_SECRET}      # 从环境变量读取
  expire: 72h
```

### 10.6 环境变量配置

```bash
# /opt/oj/.env
JWT_SECRET=your_jwt_secret_here

# 加载环境变量（在 systemd 服务中）
# 在 /etc/systemd/system/oj.service 的 [Service] 部分添加：
EnvironmentFile=/opt/oj/.env
```

**注意**：AI 判题的 API Key **不再需要通过环境变量配置**，部署后可通过管理后台动态设置：
1. 访问网站，使用管理员账号登录
2. 进入 **管理后台** → **系统设置**
3. 配置 AI 判题参数并保存

### 10.7 资源优化（2核2G 服务器）

```yaml
# 针对低配服务器的优化建议

# 1. 限制判题并发
judge:
  workers: 1                 # 只用 1 个 worker，避免 OOM

# 2. 使用 SQLite 而非 PostgreSQL
database:
  driver: sqlite

# 3. 不使用 Redis，用内存队列
cache:
  type: memory

# 4. 限制 isolate 资源
# /etc/isolate/default.conf
max-processes = 10
max-fsize = 10240            # 10MB
wall-time = 30

# 5. 启用 swap（应急）
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
```

### 10.8 数据备份

```bash
# 创建备份脚本 /opt/oj/backup.sh
#!/bin/bash
BACKUP_DIR=/opt/oj/backups
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

# 备份数据库（以 /opt/oj/configs/config.yaml 中 database.path 为准）
cp /opt/oj/data/oj.db $BACKUP_DIR/oj_$DATE.db

# 备份题目数据
tar -czf $BACKUP_DIR/problems_$DATE.tar.gz /opt/oj/data/problems

# 保留最近 7 天的备份
find $BACKUP_DIR -name "*.db" -mtime +7 -delete
find $BACKUP_DIR -name "*.tar.gz" -mtime +7 -delete

# 添加定时任务
# crontab -e
# 0 3 * * * /opt/oj/backup.sh
```

---

## 11. 安全注意事项

### 11.1 代码执行安全
- 使用 isolate 沙箱隔离用户代码
- 限制系统调用（禁止网络、文件系统访问）
- 限制 CPU 时间和内存使用
- 限制输出大小

### 11.2 API 安全
- 所有敏感接口需要 JWT 认证
- 实现请求频率限制（Rate Limiting）
- 输入验证和 SQL 注入防护
- CSRF 防护

### 11.3 数据安全
- 密码使用 bcrypt 加密存储
- AI API Key 存储在数据库中，通过管理后台配置
- JWT Secret 使用环境变量，不硬编码
- 定期备份数据库
- HTTPS 加密传输

---

## 12. 后续扩展方向

- **比赛模块**：支持 ACM/IOI 赛制
- **讨论区**：题目讨论、题解分享
- **代码查重**：检测抄袭
- **更多 AI 能力**：代码建议、错误分析、自动生成测试数据
- **Docker 化部署**：便于迁移和扩展
- **微服务拆分**：判题服务独立，支持水平扩展

---

## 附录 A：常用算法关键词（供 AI 判题参考）

| 算法类别 | 关键词 |
|---------|--------|
| 排序 | 冒泡排序、选择排序、插入排序、快速排序、归并排序、堆排序 |
| 搜索 | DFS、BFS、二分查找、A* |
| 动态规划 | DP、状态转移、记忆化搜索、背包问题 |
| 贪心 | 贪心、局部最优 |
| 图论 | Dijkstra、Floyd、Bellman-Ford、最小生成树、拓扑排序 |
| 数据结构 | 链表、栈、队列、哈希表、树、堆、并查集、线段树、树状数组 |
| 字符串 | KMP、字典树、后缀数组、Manacher |
| 数学 | 素数筛、GCD、快速幂、组合数学 |

---

*文档版本：v1.1*
*创建日期：2025-01-28*
*最后更新：2026-02-11*
