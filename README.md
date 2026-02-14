# A toy project written by Opus4.5, Codex and Gemini

# OJ 在线评测系统

一个支持 AI 智能判题的现代化在线评测系统（Online Judge）。

## 特性

- **传统 OJ 功能**：用户管理、题目管理、代码提交、自动评测、排行榜
- **账号统一分配**：普通用户账号由管理员在后台创建与分配
- **自助修改密码**：用户可在个人中心修改密码
- **文件操作题目**：可要求从指定文件读入并输出到指定文件
- **AI 智能判题**：调用 DeepSeek API 分析代码，检测是否使用指定算法/语言（不满足要求时分数封顶 50）
- **灵活配置**：每道题目可独立配置 AI 判题要求
- **现代化 UI**：采用 Swiss 风格重构界面，统一状态色、表格居中与极简导航
- **分屏题面编辑**：题目详情页支持题面/代码可拖拽分屏，内置 Monaco 编辑器并可调字号与 Tab Size
- **后台双栏编辑**：题目管理与比赛管理的 Markdown 描述编辑均支持左编辑右预览
- **测试点上传增强**：支持单文件/Zip 上传进度显示、连续上传无需刷新
- **题面图片上传**：题目编辑页支持上传题面图片并返回 Markdown，可选择插入到题目描述/输入格式/输出格式/提示
- **整题重测**：题目编辑页支持一键将该题历史提交重新入队评测
- **比赛功能**：支持 OI / IOI 赛制，按用户/分组配置参赛范围，支持固定起止与窗口期+个人时长两种计时模式
- **比赛榜单切换**：管理员可切换查看 `赛时|赛后`、`赛时`、`赛后` 三种排行榜视图并按当前模式导出 CSV
- **权限分层**：`super_admin`（超级管理员）与 `admin`（普通管理员）均可进行题目/比赛/系统编辑，仅 `super_admin` 可调整用户管理员身份
- **隐藏题可见性**：固定起止比赛在开赛后可见；窗口期比赛需先点击“开始比赛”后可见；比赛结束后参赛用户仍可访问
- **通过标识**：题目列表与比赛题目展示已通过标识（比赛以开始后通过为准）
- **提交记录权限**：提交列表与提交详情需登录访问，普通用户仅可查看自己的提交
- **OI 比赛公平机制**：进行中的 OI 比赛对普通用户隐藏详细评测结果（显示为 `Submitted`）

## 技术栈

### 后端
- Go 1.21+
- Gin Web 框架
- GORM + SQLite
- JWT 认证

### 前端
- Vue 3 + Vite
- Element Plus
- Pinia 状态管理
- Vue Router
- Monaco Editor + splitpanes

## 快速开始

### 开发环境要求

- Go 1.21+
- Node.js 18+
- GCC/G++ (用于判题)

### 后端启动

```bash
cd backend

# 安装依赖
go mod tidy

# 运行
go run ./cmd/server
```

后端默认运行在 http://localhost:8080

### 前端启动

```bash
cd frontend

# 安装依赖
npm install

# 运行开发服务器
npm run dev
```

前端默认运行在 http://localhost:3000

### 默认账号

- 超级管理员：`admin` / `admin123`

普通用户账号请由管理员在 **管理后台 → 用户管理** 中创建后分配给学生。
建议管理员和用户首次登录后在 **个人中心** 修改密码。

## 配置说明

### 后端配置 (`backend/configs/config.yaml`)

```yaml
server:
  port: 8080
  mode: debug  # debug 或 release

database:
  driver: sqlite
  path: ./data/oj.db

judge:
  sandbox: simple  # simple, isolate, docker
  workers: 2
  timeout: 30
```

### 环境变量（可选）

创建 `.env` 文件：

```env
JWT_SECRET=your_jwt_secret_here
```

说明：
- `.env` 目前只用于 `JWT_SECRET`。
- **AI 判题配置不从 `.env` 或 `config.yaml` 读取**，只通过管理后台写入数据库设置。

### AI 判题配置

AI 判题的 API Key **不再需要通过环境变量配置**，可以在运行时通过管理后台动态设置：

1. 登录管理员账号
2. 进入 **管理后台** → **系统设置**
3. 配置 AI 判题参数（API Key、模型、超时时间等）
4. 点击保存

这种方式更加灵活，可以随时修改配置而无需重启服务。

## 部署

### 服务器要求

- Linux (Debian 12 / Ubuntu 20.04+)
- 2核 2G 内存（最低配置）
- 公网 IP / 域名

### 部署步骤

> 提示：SQLite 驱动依赖 CGO，建议在 Linux 环境构建后端（需要 gcc）。

**初次部署可以直接运行下面的命令**

```bash
bash deploy/scripts/deploy_fresh_local.sh <server> <domain> user port
```

1. **构建后端**
```bash
cd backend
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o oj-server ./cmd/server
```

2. **构建前端**
```bash
cd frontend
npm run build
```

3. **上传到服务器**
```bash
scp backend/oj-server user@server:/opt/oj/
scp -r frontend/dist/* user@server:/opt/oj/static/
scp -r backend/configs user@server:/opt/oj/
```

4. **配置 systemd 服务**
```bash
# 复制 deploy/systemd/oj.service 到 /etc/systemd/system/
systemctl enable oj
systemctl start oj
```

5. **配置 Nginx**
```bash
# 复制 deploy/nginx/oj.conf 到 /etc/nginx/sites-available/
# 修改域名后启用
ln -s /etc/nginx/sites-available/oj /etc/nginx/sites-enabled/
nginx -t && systemctl reload nginx
```

部署大文件上传建议保持以下配置：
- `client_max_body_size 200m`
- `proxy_send_timeout 600s`
- `proxy_read_timeout 600s`

详细部署说明请参考 `PROJECT_DESIGN.md`。

## 更迭

### 本地

```bash
bash start-backend.sh
bash start-frontend.sh
```

### 服务器

```bash
cd backend && CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o oj-server ./cmd/server

cd frontend && npm ci && npm run build

ssh <user@server> "sudo systemctl stop oj"
scp ./backend/oj-server <user@server>:/opt/oj/oj-server.new
ssh <user@server> "mv /opt/oj/oj-server.new /opt/oj/oj-server && chown www-data:www-data /opt/oj/oj-server"
ssh <user@server> "rm -rf /opt/oj/static/*"
scp -r ./frontend/dist/* <user@server>:/opt/oj/static/
ssh <user@server> "sudo systemctl restart oj"
```

注意：
- **不要**覆盖 `/opt/oj/configs/config.yaml` 和 `/opt/oj/.env`，避免数据路径或 JWT 密钥变化。
- `deploy/scripts/deploy_fresh_local.sh` 会执行 `--wipe`，仅适用于全新部署，不用于更新。

## AI 判题功能

### 系统配置

首先需要在管理后台配置 AI 服务：

1. 登录管理员账号，进入 **管理后台** → **系统设置**
2. 开启 "启用 AI 判题"
3. 填入 API Key（推荐使用 [DeepSeek](https://platform.deepseek.com/)，性价比高）
4. 可选：修改 API 地址、模型名称、超时时间
5. 点击 "保存设置"

### 题目配置

在创建/编辑题目时，可配置以下 AI 判题选项：

| 配置项 | 说明 |
|--------|------|
| `enabled` | 是否启用 AI 判题 |
| `required_algorithm` | 要求使用的算法（如"动态规划"） |
| `required_language` | 要求使用的语言 |
| `forbidden_features` | 禁止使用的特性 |
| `custom_prompt` | 自定义判题说明 |
| `strict_mode` | 严格模式（AI 不通过则 WA）|

### 工作流程

1. 用户提交代码
2. 系统先进行传统评测（编译、运行测试点）
3. 如果题目启用 AI 判题，调用 AI API 分析代码
4. 综合两者结果给出最终判定；若 AI 判定未满足要求，则分数封顶 50

## 项目结构

```
oj-system/
├── backend/                 # 后端服务
│   ├── cmd/server/          # 入口
│   ├── internal/
│   │   ├── config/          # 配置
│   │   ├── handler/         # HTTP 处理器
│   │   ├── service/         # 业务逻辑
│   │   ├── repository/      # 数据访问
│   │   ├── model/           # 数据模型
│   │   ├── middleware/      # 中间件
│   │   ├── judge/           # 判题核心
│   │   └── router/          # 路由
│   └── configs/             # 配置文件
│
├── frontend/                # 前端项目
│   ├── src/
│   │   ├── api/             # API 调用
│   │   ├── components/      # 组件
│   │   ├── views/           # 页面
│   │   ├── stores/          # 状态管理
│   │   └── router/          # 路由
│   └── package.json
│
├── deploy/                  # 部署相关
│   ├── nginx/
│   ├── systemd/
│   └── scripts/
│
└── PROJECT_DESIGN.md        # 详细设计文档
```

## License

MIT
