# OJ 系统技术参考文档

本文档详细说明系统各模块的功能、接口和数据结构，便于开发调试。

---

## 目录

1. [项目结构](#1-项目结构)
2. [后端模块详解](#2-后端模块详解)
3. [API 接口文档](#3-api-接口文档)
4. [数据模型](#4-数据模型)
5. [判题系统](#5-判题系统)
6. [AI 判题系统](#6-ai-判题系统)
7. [前端结构](#7-前端结构)
8. [调试指南](#8-调试指南)

---

## 1. 项目结构

```
OJ/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go              # 程序入口
│   ├── configs/
│   │   ├── config.yaml              # 开发配置
│   │   └── config.production.yaml   # 生产配置
│   └── internal/
│       ├── config/
│       │   └── config.go            # 配置加载
│       ├── model/
│       │   ├── user.go              # 用户模型
│       │   ├── problem.go           # 题目模型
│       │   ├── submission.go        # 提交模型
│       │   ├── contest.go           # 比赛模型
│       │   ├── contest_participation.go # 窗口期比赛会话模型
│       │   ├── setting.go           # 系统设置模型
│       │   └── response.go          # 响应结构
│       ├── repository/
│       │   ├── database.go          # 数据库初始化
│       │   ├── user_repo.go         # 用户数据访问
│       │   ├── problem_repo.go      # 题目数据访问
│       │   ├── submission_repo.go   # 提交数据访问
│       │   ├── contest_repo.go      # 比赛数据访问
│       │   ├── contest_participation_repo.go # 比赛会话数据访问
│       │   └── setting_repo.go      # 设置数据访问
│       ├── service/
│       │   ├── user_service.go      # 用户业务逻辑
│       │   ├── problem_service.go   # 题目业务逻辑
│       │   ├── submission_service.go# 提交业务逻辑
│       │   ├── contest_service.go   # 比赛业务逻辑
│       │   ├── setting_service.go   # 设置业务逻辑
│       │   └── maintenance_service.go # 统计维护任务
│       ├── handler/
│       │   ├── user_handler.go      # 用户 HTTP 处理
│       │   ├── problem_handler.go   # 题目 HTTP 处理
│       │   ├── submission_handler.go# 提交 HTTP 处理
│       │   ├── contest_handler.go   # 比赛 HTTP 处理
│       │   ├── setting_handler.go   # 设置 HTTP 处理
│       │   ├── statistics_handler.go# 公开统计 HTTP 处理
│       │   └── utils.go             # 处理器工具函数
│       ├── middleware/
│       │   ├── auth.go              # JWT 认证中间件
│       │   ├── cors.go              # 跨域中间件
│       │   └── ratelimit.go         # 限流中间件
│       ├── router/
│       │   └── router.go            # 路由配置
│       ├── judge/
│       │   ├── judger.go            # 判题主逻辑
│       │   ├── queue/
│       │   │   └── queue.go         # 判题队列
│       │   ├── sandbox/
│       │   │   └── sandbox.go       # 代码执行沙箱
│       │   └── ai/
│       │       └── deepseek.go      # AI 判题客户端
│       └── utils/
│           ├── jwt.go               # JWT 工具
│           └── password.go          # 密码工具
│
├── frontend/
│   └── src/
│       ├── api/                     # API 调用封装
│       ├── components/              # Vue 组件
│       ├── views/                   # 页面视图
│       ├── stores/                  # Pinia 状态
│       └── router/                  # Vue Router
│
└── deploy/                          # 部署配置
```

---

## 2. 后端模块详解

### 2.1 配置模块 (`internal/config`)

**文件**: `config.go`

**功能**: 加载 YAML 配置文件，支持环境变量替换

**主要结构**:
```go
type Config struct {
    Server   ServerConfig   // 服务器配置
    Database DatabaseConfig // 数据库配置
    Judge    JudgeConfig    // 判题配置
    AI       AIConfig       // AI 配置
    Paths    PathsConfig    // 路径配置
    JWT      JWTConfig      // JWT 配置
}
```

**关键函数**:
| 函数 | 说明 |
|------|------|
| `Load(path string) (*Config, error)` | 加载配置文件 |
| `replaceEnvVars(content string) string` | 替换 `${VAR}` 形式的环境变量 |

**全局变量**: `GlobalConfig *Config` - 加载后可全局访问

**注意**：
- 当前代码不会从 `config.yaml` 的 `ai` 段读取 AI 配置。
- AI 判题设置仅通过管理后台写入数据库 `settings` 表读取。

---

### 2.2 数据模型 (`internal/model`)

#### 2.2.1 用户模型 (`user.go`)

```go
type User struct {
    ID           uint      `json:"id"`
    Username     string    `json:"username"`
    Email        string    `json:"email"`
    PasswordHash string    `json:"-"`          // 不输出到 JSON
    StudentID    string    `json:"student_id"`
    Role         string    `json:"role"`       // "user" | "admin"
    Group        string    `json:"group"`
    SolvedCount  int       `json:"solved_count"`
    AcceptedCount int      `json:"accepted_count"`
    SubmitCount  int       `json:"submit_count"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

**请求/响应结构**:
- `AdminCreateUserRequest` - 管理员创建用户请求
- `AdminCreateUsersRequest` - 管理员批量创建用户请求
- `AdminUpdateUserRequest` - 管理员更新用户请求
- `UserLoginRequest` - 登录请求
- `ChangePasswordRequest` - 修改密码请求
- `UserLoginResponse` - 登录响应（含 token）
- `UserInfo` - 用户信息（不含密码）

#### 2.2.2 题目模型 (`problem.go`)

```go
type Problem struct {
    ID            uint           `json:"id"`
    Title         string         `json:"title"`
    Description   string         `json:"description"`
    InputFormat   string         `json:"input_format"`
    OutputFormat  string         `json:"output_format"`
    Hint          string         `json:"hint"`
    Samples       SampleList     `json:"samples"`       // JSON 序列化
    TimeLimit     int            `json:"time_limit"`    // 毫秒
    MemoryLimit   int            `json:"memory_limit"`  // MB
    Difficulty    string         `json:"difficulty"`    // easy|medium|hard
    Tags          StringList     `json:"tags"`          // JSON 序列化
    AIJudgeConfig *AIJudgeConfig `json:"ai_judge_config"`
    FileIOEnabled bool           `json:"file_io_enabled"`
    FileInputName string         `json:"file_input_name"`
    FileOutputName string        `json:"file_output_name"`
    IsPublic      *bool          `json:"is_public"`
    CreatedBy     uint           `json:"created_by"`
    SubmitCount   int            `json:"submit_count"`
    AcceptedCount int            `json:"accepted_count"`
    HasAccepted   bool           `json:"has_accepted"`
}

type AIJudgeConfig struct {
    Enabled           bool     `json:"enabled"`
    RequiredAlgorithm string   `json:"required_algorithm,omitempty"`
    RequiredLanguage  string   `json:"required_language,omitempty"`
    ForbiddenFeatures []string `json:"forbidden_features,omitempty"`
    CustomPrompt      string   `json:"custom_prompt,omitempty"`
    StrictMode        bool     `json:"strict_mode"`
}

type Testcase struct {
    ID         uint   `json:"id"`
    ProblemID  uint   `json:"problem_id"`
    InputFile  string `json:"input_file"`   // 文件路径
    OutputFile string `json:"output_file"`  // 文件路径
    Score      int    `json:"score"`
    IsSample   bool   `json:"is_sample"`
    OrderNum   int    `json:"order_num"`
}
```

#### 2.2.3 提交模型 (`submission.go`)

```go
// 提交状态常量
const (
    StatusPending             = "Pending"
    StatusJudging             = "Judging"
    StatusAccepted            = "Accepted"
    StatusWrongAnswer         = "Wrong Answer"
    StatusTimeLimitExceeded   = "Time Limit Exceeded"
    StatusMemoryLimitExceeded = "Memory Limit Exceeded"
    StatusRuntimeError        = "Runtime Error"
    StatusCompileError        = "Compile Error"
    StatusSystemError         = "System Error"
)

type Submission struct {
    ID              uint               `json:"id"`
    ProblemID       uint               `json:"problem_id"`
    UserID          uint               `json:"user_id"`
    Language        string             `json:"language"`      // c|cpp|python|java|go
    Code            string             `json:"code"`
    Status          string             `json:"status"`
    TimeUsed        int                `json:"time_used"`     // 毫秒
    MemoryUsed      int                `json:"memory_used"`   // KB
    Score           int                `json:"score"`         // 0-100
    TestcaseResults TestcaseResultList `json:"testcase_results"`
    AIJudgeResult   *AIJudgeResult     `json:"ai_judge_result"`
    CompileError    string             `json:"compile_error"`
    FinalMessage    string             `json:"final_message"`
    CreatedAt       time.Time          `json:"created_at"`
    ProblemTitle    string             `json:"problem_title"`
    Username        string             `json:"username"`
}

type TestcaseResult struct {
    ID      int    `json:"id"`      // 测试点序号
    Status  string `json:"status"`  // 状态
    Time    int    `json:"time"`    // 毫秒
    Memory  int    `json:"memory"`  // KB
    Message string `json:"message,omitempty"`
}

type AIJudgeResult struct {
    Enabled           bool            `json:"enabled"`
    Passed            bool            `json:"passed"`
    AlgorithmDetected string          `json:"algorithm_detected,omitempty"`
    LanguageCheck     string          `json:"language_check,omitempty"`
    Reason            string          `json:"reason,omitempty"`
    Summary           string          `json:"summary,omitempty"`
    Details           *AIJudgeDetails `json:"details,omitempty"`
}
```

#### 2.2.4 响应结构 (`response.go`)

```go
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

type PageData struct {
    Total int64       `json:"total"`
    Page  int         `json:"page"`
    Size  int         `json:"size"`
    List  interface{} `json:"list"`
}

// 响应码
const (
    CodeSuccess      = 200
    CodeBadRequest   = 400
    CodeUnauthorized = 401
    CodeForbidden    = 403
    CodeNotFound     = 404
    CodeServerError  = 500
)
```

---

### 2.3 数据访问层 (`internal/repository`)

#### 2.3.1 数据库初始化 (`database.go`)

**关键函数**:
| 函数 | 说明 |
|------|------|
| `InitDatabase(cfg *DatabaseConfig) error` | 初始化数据库连接，执行自动迁移 |
| `GetDB() *gorm.DB` | 获取数据库实例 |

**自动迁移的表**: `users`, `contests`, `problems`, `testcases`, `submissions`, `settings`

#### 2.3.2 用户仓库 (`user_repo.go`)

```go
type UserRepository struct{}

func NewUserRepository() *UserRepository
```

| 方法 | 说明 |
|------|------|
| `Create(user *User) error` | 创建用户 |
| `GetByID(id uint) (*User, error)` | 根据 ID 获取 |
| `GetByUsername(username string) (*User, error)` | 根据用户名获取 |
| `GetByEmail(email string) (*User, error)` | 根据邮箱获取 |
| `Update(user *User) error` | 更新用户 |
| `List(page, size int) ([]User, int64, error)` | 分页列表 |
| `ExistsByUsername(username string) bool` | 检查用户名是否存在 |
| `ExistsByEmail(email string) bool` | 检查邮箱是否存在 |
| `IncrementSolvedCount(userID uint) error` | 增加解题数 |
| `IncrementSubmitCount(userID uint) error` | 增加提交数 |

#### 2.3.3 题目仓库 (`problem_repo.go`)

```go
type ProblemRepository struct{}

func NewProblemRepository() *ProblemRepository
```

| 方法 | 说明 |
|------|------|
| `Create(problem *Problem) error` | 创建题目 |
| `GetByID(id uint) (*Problem, error)` | 根据 ID 获取 |
| `Update(problem *Problem) error` | 更新题目 |
| `Delete(id uint) error` | 删除题目（含测试用例） |
| `List(page, size int, difficulty, tag, keyword string) ([]Problem, int64, error)` | 分页列表（支持筛选） |
| `ListAll(page, size int, difficulty, tag, keyword string) ([]Problem, int64, error)` | 分页列表（包含隐藏题） |
| `GetTestcases(problemID uint) ([]Testcase, error)` | 获取测试用例 |
| `CreateTestcase(testcase *Testcase) error` | 创建测试用例 |
| `DeleteTestcases(problemID uint) error` | 删除所有测试用例 |
| `IncrementSubmitCount(problemID uint) error` | 增加提交数 |
| `IncrementAcceptedCount(problemID uint) error` | 增加通过数 |

#### 2.3.4 设置仓库 (`setting_repo.go`)

```go
type SettingRepository struct{}

func NewSettingRepository() *SettingRepository
```

| 方法 | 说明 |
|------|------|
| `Get(key string) (string, error)` | 获取设置值 |
| `Set(key, value string) error` | 设置值 |
| `GetAll() (map[string]string, error)` | 获取所有设置 |
| `GetMultiple(keys []string) (map[string]string, error)` | 获取多个设置 |

#### 2.3.5 提交仓库 (`submission_repo.go`)

```go
type SubmissionRepository struct{}

func NewSubmissionRepository() *SubmissionRepository
```

| 方法 | 说明 |
|------|------|
| `Create(submission *Submission) error` | 创建提交 |
| `GetByID(id uint) (*Submission, error)` | 根据 ID 获取 |
| `Update(submission *Submission) error` | 更新提交 |
| `List(page, size int, problemID, userID uint, status string) ([]SubmissionListItem, int64, error)` | 分页列表 |
| `GetPendingSubmissions(limit int) ([]Submission, error)` | 获取待判题提交 |
| `UpdateStatus(id uint, status string) error` | 更新状态 |
| `HasAccepted(userID, problemID uint) bool` | 用户是否已通过该题 |

---

### 2.4 业务逻辑层 (`internal/service`)

#### 2.4.1 用户服务 (`user_service.go`)

```go
type UserService struct{}

func NewUserService() *UserService
```

| 方法 | 说明 |
|------|------|
| `CreateUserByAdmin(req *AdminCreateUserRequest) (*User, error)` | 管理员创建用户（检查重复、密码加密） |
| `CreateUsersBatch(req *AdminCreateUsersRequest) (int, []map[string]interface{})` | 管理员批量创建用户 |
| `UpdateUserByAdmin(userID uint, req *AdminUpdateUserRequest) (*User, error)` | 管理员更新用户信息 |
| `Login(req *UserLoginRequest) (*UserLoginResponse, error)` | 登录（验证密码、生成 Token） |
| `GetProfile(userID uint) (*UserInfo, error)` | 获取个人信息 |
| `UpdateProfile(userID uint, email, studentID string) error` | 更新个人信息 |
| `ChangePassword(userID uint, oldPassword, newPassword string) error` | 修改个人密码 |
| `GetRankList(page, size int) ([]UserInfo, int64, error)` | 获取排行榜 |
| `GetUserList(page, size int) ([]User, int64, error)` | 获取用户列表（管理员） |
| `SetUserRole(userID uint, role string) error` | 设置用户角色（管理员） |

#### 2.4.2 题目服务 (`problem_service.go`)

```go
type ProblemService struct{}

func NewProblemService() *ProblemService
```

| 方法 | 说明 |
|------|------|
| `Create(req *ProblemCreateRequest, createdBy uint) (*Problem, error)` | 创建题目 |
| `GetByID(id uint) (*Problem, error)` | 获取题目详情 |
| `Update(id uint, req *ProblemCreateRequest) (*Problem, error)` | 更新题目 |
| `Delete(id uint) error` | 删除题目 |
| `List(page, size int, difficulty, tag, keyword string) ([]ProblemListItem, int64, error)` | 获取题目列表 |
| `GetTestcases(problemID uint) ([]Testcase, error)` | 获取测试用例 |
| `AddTestcase(problemID uint, inputReader, outputReader io.Reader, score int, isSample bool) error` | 添加测试用例 |
| `DeleteTestcases(problemID uint) error` | 删除所有测试用例 |

#### 2.4.3 设置服务 (`setting_service.go`)

```go
type SettingService struct{}

func GetSettingService() *SettingService  // 单例模式
```

| 方法 | 说明 |
|------|------|
| `Get(key string) string` | 获取设置值（带缓存） |
| `Set(key, value string) error` | 设置值（更新缓存） |
| `GetAISettings() *AISettings` | 获取 AI 设置 |
| `UpdateAISettings(req *UpdateAISettingsRequest) error` | 更新 AI 设置 |
| `GetAISettingsForDisplay() *AISettings` | 获取用于显示的 AI 设置（隐藏 API Key） |
| `HasAIAPIKey() bool` | 检查是否配置了 API Key |

**设置键名常量**:
- `ai_enabled` - AI 判题是否启用
- `ai_provider` - AI 提供商
- `ai_api_key` - API Key
- `ai_api_url` - API 地址
- `ai_model` - 模型名称
- `ai_timeout` - 超时时间

#### 2.4.4 提交服务 (`submission_service.go`)

```go
type SubmissionService struct{}

func NewSubmissionService() *SubmissionService
```

| 方法 | 说明 |
|------|------|
| `Submit(req *SubmissionCreateRequest, userID uint) (*Submission, error)` | 提交代码 |
| `GetByID(id uint, userID uint, isAdmin bool) (*Submission, error)` | 获取提交详情（非管理员仅能查看本人） |
| `List(page, size int, problemID, filterUserID uint, status string, viewerID uint, isAdmin bool) (*PageData, error)` | 获取提交列表（支持权限过滤与比赛期遮罩） |
| `UpdateResult(submission *Submission) error` | 更新判题结果 |
| `GetPendingSubmissions(limit int) ([]Submission, error)` | 获取待判题提交 |

---

### 2.5 HTTP 处理器 (`internal/handler`)

#### 2.5.1 工具函数 (`utils.go`)

| 函数 | 说明 |
|------|------|
| `getIntQuery(c *gin.Context, key string, defaultVal int) int` | 获取 int 查询参数 |
| `getUintQuery(c *gin.Context, key string) uint` | 获取 uint 查询参数 |
| `getUintParam(c *gin.Context, key string) uint` | 获取 uint 路径参数 |
| `getIntFormValue(c *gin.Context, key string, defaultVal int) int` | 获取 int 表单值 |

---

### 2.6 中间件 (`internal/middleware`)

#### 2.6.1 认证中间件 (`auth.go`)

| 中间件/函数 | 说明 |
|------------|------|
| `AuthMiddleware()` | 强制 JWT 认证 |
| `AdminMiddleware()` | 要求管理员权限 |
| `OptionalAuthMiddleware()` | 可选认证（不强制） |
| `GetUserID(c *gin.Context) uint` | 从上下文获取用户 ID |
| `GetUsername(c *gin.Context) string` | 从上下文获取用户名 |
| `GetRole(c *gin.Context) string` | 从上下文获取角色 |
| `IsAdmin(c *gin.Context) bool` | 判断是否管理员 |

**上下文存储的值**:
- `userID` (uint) - 用户 ID
- `username` (string) - 用户名
- `role` (string) - 用户角色

#### 2.6.2 跨域中间件 (`cors.go`)

`CORSMiddleware()` - 设置 CORS 头，允许跨域请求

#### 2.6.3 限流中间件 (`ratelimit.go`)

| 中间件 | 说明 |
|--------|------|
| `RateLimitMiddleware(limit int, window time.Duration)` | 通用限流 |
| `SubmitRateLimitMiddleware()` | 提交限流（每分钟 10 次） |

---

### 2.7 工具函数 (`internal/utils`)

#### 2.7.1 JWT 工具 (`jwt.go`)

```go
type Claims struct {
    UserID   uint   `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}
```

| 函数 | 说明 |
|------|------|
| `InitJWT(secret string)` | 初始化 JWT 密钥 |
| `GenerateToken(userID uint, username, role string) (string, error)` | 生成 Token |
| `ParseToken(tokenString string) (*Claims, error)` | 解析 Token |

#### 2.7.2 密码工具 (`password.go`)

| 函数 | 说明 |
|------|------|
| `HashPassword(password string) (string, error)` | 密码哈希（bcrypt） |
| `CheckPassword(password, hash string) bool` | 验证密码 |

---

## 3. API 接口文档

### 3.1 用户模块 `/api/v1/user`

**账号分配说明**：普通用户账号由管理员在后台创建与分配，客户端不再提供注册入口。

---

#### POST `/login` - 用户登录

**请求体**:
```json
{
    "username": "string",
    "password": "string"
}
```

**成功响应** (200):
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIs...",
        "user": {
            "id": 1,
            "username": "testuser",
            "role": "user"
        }
    }
}
```

**错误响应**:
- 400: 用户名或密码错误

---

#### GET `/profile` - 获取个人信息

**认证**: 需要 Bearer Token

**成功响应** (200):
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "id": 1,
        "username": "testuser",
        "email": "test@example.com",
        "student_id": "2021001",
        "role": "user",
        "solved_count": 5,
        "submit_count": 20
    }
}
```

---

#### PUT `/profile` - 更新个人信息

**认证**: 需要 Bearer Token

**请求体**:
```json
{
    "email": "new@example.com",
    "student_id": "2021002"
}
```

**说明**:
- `is_public=false` 表示隐藏题：仅管理员或参赛用户可访问。
- 固定起止比赛：开赛后可访问；窗口期比赛：需先开始个人会话后可访问；比赛结束后参赛用户仍可访问。

**成功响应** (200):
```json
{
    "code": 200,
    "message": "更新成功"
}
```

---

#### PUT `/password` - 修改密码

**认证**: 需要 Bearer Token

**请求体**:
```json
{
    "old_password": "old_pass",
    "new_password": "new_pass_123"
}
```

**成功响应** (200):
```json
{
    "code": 200,
    "message": "修改成功"
}
```

**错误响应**:
- 400: 原密码错误 / 新密码不合法

---

### 3.2 题目模块 `/api/v1/problem`

**上传链路（当前默认）**:
- 前端请求超时默认 `180s`，测试点上传接口单独使用 `600s` 超时并上报上传进度。
- Nginx 代理建议设置：`client_max_body_size 200m`、`proxy_send_timeout 600s`、`proxy_read_timeout 600s`。
- 后端 Gin：`MaxMultipartMemory = 256MB`。
- 题面图片上传：单图大小上限 `10MB`，支持 `png/jpg/jpeg/gif/webp/bmp`。

#### GET `/list` - 获取题目列表

**查询参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| `page` | int | 页码，默认 1 |
| `size` | int | 每页数量，默认 20 |
| `difficulty` | string | 难度筛选：easy/medium/hard |
| `tag` | string | 标签筛选 |
| `keyword` | string | 关键词搜索 |
**说明**: 
- 未登录/普通用户仅返回公开题目（`is_public=true`），管理员登录可返回全部题目
- 登录状态下会返回 `has_accepted` 标识用户是否已通过题目

**成功响应** (200):
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "total": 100,
        "page": 1,
        "size": 20,
        "list": [
            {
                "id": 1,
                "title": "两数之和",
                "difficulty": "easy",
                "tags": ["数组", "哈希表"],
                "submit_count": 1000,
                "accepted_count": 500,
                "has_ai_judge": true,
                "has_file_io": true,
                "has_accepted": true
            }
        ]
    }
}
```

---

#### GET `/:id` - 获取题目详情

**路径参数**: `id` - 题目 ID
**说明**: 
- 登录状态下会返回 `has_accepted` 标识用户是否已通过题目
- 当题目为隐藏题时，仅管理员或参赛用户可访问；固定起止比赛要求已开赛，窗口期比赛要求个人会话已开始；赛后参赛用户仍可访问
- 进行中的比赛里，非管理员访问隐藏题时不返回该题标签（`tags` 为空）

**成功响应** (200):
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "id": 1,
        "title": "两数之和",
        "has_accepted": true,
        "description": "给定一个整数数组...",
        "input_format": "第一行输入...",
        "output_format": "输出...",
        "hint": "可选提示信息",
        "samples": [
            {"input": "1 2\n3 4", "output": "3\n7"}
        ],
        "time_limit": 1000,
        "memory_limit": 256,
        "difficulty": "easy",
        "tags": ["数组", "哈希表"],
        "file_io_enabled": false,
        "file_input_name": "",
        "file_output_name": "",
        "ai_judge_config": {
            "enabled": true,
            "required_algorithm": "哈希表",
            "required_language": "",
            "forbidden_features": [],
            "custom_prompt": "",
            "strict_mode": true
        }
    }
}
```

---

#### POST `/` - 创建题目（管理员）

**认证**: 需要 Bearer Token + 管理员权限

**请求体**:
```json
{
    "title": "题目标题",
    "description": "题目描述（支持 Markdown）",
    "input_format": "输入格式",
    "output_format": "输出格式",
    "hint": "可选提示",
    "samples": [
        {"input": "1 2", "output": "3"}
    ],
    "time_limit": 1000,
    "memory_limit": 256,
    "difficulty": "easy",
    "tags": ["数组"],
    "is_public": true,
    "file_io_enabled": true,
    "file_input_name": "data.in",
    "file_output_name": "data.out",
    "ai_judge_config": {
        "enabled": true,
        "required_algorithm": "动态规划",
        "required_language": "C++",
        "forbidden_features": ["STL sort"],
        "custom_prompt": "必须使用自底向上的 DP",
        "strict_mode": true
    }
}
```

---

#### PUT `/:id` - 更新题目（管理员）

**认证**: 需要 Bearer Token + 管理员权限

**请求体**: 同创建

---

#### DELETE `/:id` - 删除题目（管理员）

**认证**: 需要 Bearer Token + 管理员权限

---

#### POST `/:id/image` - 上传题面图片（管理员）

**认证**: 需要 Bearer Token + 管理员权限

**请求类型**: `multipart/form-data`

**表单字段**:
| 字段 | 类型 | 说明 |
|------|------|------|
| `image` | file | 图片文件（`png/jpg/jpeg/gif/webp/bmp`） |

**说明**:
- 图片会保存到题目目录：`{problems_path}/{problem_id}/images/`。
- 成功后返回可直接使用的 Markdown 片段。

**成功响应** (200):
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "url": "/api/v1/problem/1/image/1739769823_ab12cd34ef56.png",
        "markdown": "![example](/api/v1/problem/1/image/1739769823_ab12cd34ef56.png)",
        "filename": "1739769823_ab12cd34ef56.png"
    }
}
```

---

#### GET `/:id/image/:filename` - 获取题面图片

**认证**: 无需认证（按题目可见性访问）

**说明**:
- 图片通过后端读取题目目录并返回文件流。
- 为避免路径穿越，文件名会做安全校验（禁止 `..` 与目录分隔符）。

---

#### POST `/:id/testcase` - 上传测试用例（管理员）

**认证**: 需要 Bearer Token + 管理员权限

**请求类型**: `multipart/form-data`

**表单字段**:
| 字段 | 类型 | 说明 |
|------|------|------|
| `input` | file | 输入文件 |
| `output` | file | 输出文件 |
| `score` | int | 分数，默认 10 |
| `is_sample` | string | 是否为样例，"true"/"false" |

---

#### POST `/:id/testcase/zip` - 批量上传测试用例（管理员）

**认证**: 需要 Bearer Token + 管理员权限

**请求类型**: `multipart/form-data`

**表单字段**:
| 字段 | 类型 | 说明 |
|------|------|------|
| `zip_file` | file | Zip 压缩包（包含成对的 `.in` + `.out/.ans`） |

**说明**:
- 该操作会覆盖原有测试点（先删除再重建）。
- 会按文件名中的数字顺序生成测试点序号并自动分配分值总和 100。
- Zip 文件在服务端保存临时文件后，由后端服务解压并落盘到题目数据目录。

---

#### POST `/:id/rejudge` - 整题重测（管理员）

**认证**: 需要 Bearer Token + 管理员权限

**说明**:
- 将该题目的历史提交批量重置为 `Pending` 后重新入队判题。
- 正在 `Pending/Judging` 的提交不会重复入队。

**成功响应** (200):
```json
{
    "code": 200,
    "message": "整题重测任务已提交",
    "data": {
        "total": 120,
        "queued": 120,
        "failed": 0
    }
}
```

---

#### GET `/:id/testcases` - 获取测试用例列表（管理员）

**认证**: 需要 Bearer Token + 管理员权限

**成功响应** (200):
```json
{
    "code": 200,
    "message": "success",
    "data": [
        {
            "id": 1,
            "problem_id": 1,
            "input_file": "/opt/oj/data/problems/1/1.in",
            "output_file": "/opt/oj/data/problems/1/1.out",
            "score": 10,
            "is_sample": false,
            "order_num": 1
        }
    ]
}
```

---

#### DELETE `/:id/testcases` - 删除所有测试用例（管理员）

**认证**: 需要 Bearer Token + 管理员权限

---

### 3.3 提交模块 `/api/v1/submission`

#### POST `/` - 提交代码

**认证**: 需要 Bearer Token

**限流**: 每分钟最多 10 次

**请求体**:
```json
{
    "problem_id": 1,
    "language": "cpp",      // c, cpp, python, java, go
    "code": "#include <iostream>\nint main() { ... }"
}
```

**成功响应** (200):
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "id": 12345,
        "problem_id": 1,
        "user_id": 1,
        "language": "cpp",
        "status": "Pending",
        "created_at": "2025-01-28T10:00:00Z"
    }
}
```

---

#### GET `/:id` - 获取提交详情

**认证**: 需要 Bearer Token（管理员可查看所有，普通用户仅能查看本人）
**说明**:
- 进行中的 OI 比赛内，普通用户查看本人提交时会被遮罩为 `Submitted`，并隐藏分数、测试点、AI 结果与编译信息；管理员不受影响。

**成功响应** (200):
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "id": 12345,
        "problem_id": 1,
        "user_id": 1,
        "language": "cpp",
        "code": "...",
        "status": "Wrong Answer",
        "time_used": 15,
        "memory_used": 1024,
        "score": 80,
        "testcase_results": [
            {"id": 1, "status": "Accepted", "time": 10, "memory": 512},
            {"id": 2, "status": "Accepted", "time": 15, "memory": 1024},
            {"id": 3, "status": "Wrong Answer", "time": 12, "memory": 800}
        ],
        "ai_judge_result": {
            "enabled": true,
            "passed": false,
            "algorithm_detected": "暴力枚举",
            "language_check": "passed",
            "reason": "题目要求使用动态规划算法，但检测到代码使用了暴力枚举方法。",
            "details": {
                "required": "动态规划",
                "detected": "暴力枚举",
                "confidence": 0.92
            }
        },
        "compile_error": "",
        "final_message": "测试点部分通过，且未使用要求的算法",
        "created_at": "2025-01-28T10:00:00Z"
    }
}
```

---

#### GET `/list` - 获取提交列表

**认证**: 需要 Bearer Token（管理员可查看所有，普通用户仅能查看本人）
**说明**:
- 进行中的 OI 比赛内，普通用户列表中的相关提交会显示为 `Submitted`，`score/time_used/memory_used` 被置为 0；管理员不受影响。

**查询参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| `page` | int | 页码 |
| `size` | int | 每页数量 |
| `problem_id` | uint | 题目 ID 筛选 |
| `user_id` | uint | 用户 ID 筛选（仅管理员生效） |
| `status` | string | 状态筛选 |

---

#### GET `/my` - 获取我的提交

**认证**: 需要 Bearer Token

**查询参数**: 同上（自动筛选当前用户）

---

### 3.4 排行榜 `/api/v1/rank`

#### GET `/` - 获取排行榜

**查询参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| `page` | int | 页码 |
| `size` | int | 每页数量 |

**成功响应** (200):
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "total": 100,
        "page": 1,
        "size": 20,
        "list": [
            {
                "id": 1,
                "username": "user1",
                "solved_count": 50,
                "submit_count": 100
            }
        ]
    }
}
```

---

### 3.5 比赛模块 `/api/v1/contest`

#### GET `/list` - 获取比赛列表

**认证**: 需要 Bearer Token

**查询参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| `page` | int | 页码 |
| `size` | int | 每页数量 |

**成功响应** (200):
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "total": 2,
        "page": 1,
        "size": 20,
        "list": [
            {
                "id": 1,
                "title": "期中赛",
                "type": "oi",
                "timing_mode": "window",
                "duration_minutes": 180,
                "start_at": "2026-03-01T08:00:00Z",
                "end_at": "2026-03-01T11:00:00Z",
                "problem_count": 5
            }
        ]
    }
}
```

#### GET `/:id` - 获取比赛详情

**认证**: 需要 Bearer Token（且在比赛允许名单/分组内）
**说明**:
- `has_accepted` 仅在以下情况展示：管理员、IOI 赛制、或比赛已结束；进行中的 OI 对普通用户不展示通过信息
- `has_submitted` 表示在比赛时间范围内是否提交过该题
- `timing_mode` 支持 `fixed`（固定起止）与 `window`（窗口期 + 个人固定时长）
- 窗口期比赛中，用户需先调用 `POST /contest/:id/start` 启动个人比赛会话
- `my_live_total` / `my_post_total` 分别表示赛时/赛后得分，`my_total` 为两者之和

**成功响应** (200):
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "contest": {
            "id": 1,
            "title": "期中赛",
            "type": "oi",
            "timing_mode": "window",
            "duration_minutes": 180,
            "start_at": "2026-03-01T08:00:00Z",
            "end_at": "2026-03-01T11:00:00Z",
            "problem_ids": [1, 2, 3],
            "allowed_users": [10, 11],
            "allowed_groups": ["ClassA"]
        },
        "problems": [
            {"id": 1, "title": "A+B", "difficulty": "easy", "has_accepted": false, "has_submitted": true}
        ],
        "session": {
            "started": true,
            "can_start": false,
            "in_live": true,
            "start_at": "2026-03-01T08:30:00Z",
            "end_at": "2026-03-01T11:30:00Z",
            "remaining_seconds": 5400
        },
        "my_live_total": 160,
        "my_post_total": 20,
        "my_total": 180
    }
}
```

#### POST `/:id/start` - 开始窗口期个人比赛

**认证**: 需要 Bearer Token

**说明**:
- 仅 `timing_mode=window` 的比赛可调用。
- 用户在比赛窗口期内调用后，系统会创建或复用其个人会话（`start_at/end_at`）。

**成功响应** (200):
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "contest_id": 1,
        "user_id": 10,
        "start_at": "2026-03-01T08:30:00Z",
        "end_at": "2026-03-01T11:30:00Z"
    }
}
```

### 3.6 统计模块 `/api/v1/statistics`

#### GET `/` - 获取系统统计（公开）

**成功响应** (200):
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "problems": 10,
        "users": 30,
        "submissions": 120
    }
}
```

### 3.7 管理模块 `/api/v1/admin`

#### POST `/users` - 创建用户（管理员分配账号）

**认证**: 需要 Bearer Token + 管理员权限

**请求体**:
```json
{
    "username": "string",     // 3-20 字符，必填
    "email": "string",        // 邮箱格式，可选
    "password": "string",     // 6-20 字符，必填
    "student_id": "string",   // 可选
    "role": "user",           // user 或 admin，可选
    "group": "ClassA"         // 分组，可选
}
```

**成功响应** (200):
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "id": 1,
        "username": "testuser",
        "email": "test@example.com",
        "student_id": "",
        "role": "user",
        "solved_count": 0,
        "submit_count": 0
    }
}
```

#### POST `/users/batch` - 批量导入用户

**认证**: 需要 Bearer Token + 管理员权限

**请求体**:
```json
{
    "users": [
        {
            "username": "student01",
            "password": "pass123",
            "email": "",
            "student_id": "2025001",
            "group": "ClassA",
            "role": "user"
        }
    ]
}
```

**成功响应** (200):
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "total": 10,
        "created": 8,
        "failed": 2,
        "errors": [
            {"index": 3, "username": "student04", "error": "用户名已存在"}
        ]
    }
}
```

#### PUT `/users/:id` - 更新用户信息

**认证**: 需要 Bearer Token + 管理员权限

**请求体**:
```json
{
    "username": "string",     // 可选
    "email": "string",        // 可选（可置空）
    "student_id": "string",   // 可选（可置空）
    "group": "ClassA",        // 可选（可置空）
    "role": "user",           // 可选
    "password": "string"      // 可选（重置密码）
}
```

#### POST `/contests` - 创建比赛

**认证**: 需要 Bearer Token + 管理员权限

**请求体**:
```json
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

**说明**:
- `allowed_users` 与 `allowed_groups` 至少填写一个，否则普通用户无法看到/参加比赛。
- 当 `timing_mode=window` 时，`duration_minutes` 必须大于 0。

#### PUT `/contests/:id` - 更新比赛

**认证**: 需要 Bearer Token + 管理员权限

**请求体**: 同创建比赛

#### DELETE `/contests/:id` - 删除比赛

**认证**: 需要 Bearer Token + 管理员权限

#### POST `/contests/:id/refresh` - 刷新比赛统计（管理员）

**认证**: 需要 Bearer Token + 管理员权限

**说明**: 赛后手动触发题目与用户统计同步。

#### GET `/contests/:id/leaderboard` - 比赛排行榜（管理员）

**认证**: 需要 Bearer Token + 管理员权限
**查询参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| `board_mode` | string | 榜单模式：`combined`（默认，赛时|赛后）、`live`（赛时）、`post`（赛后） |

**说明**: OI/IOI 赛制均按每题最后一次提交得分汇总

**成功响应** (200):
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "contest": {
            "id": 1,
            "title": "期中赛",
            "type": "oi"
        },
        "board_mode": "combined",
        "problem_ids": [1, 2, 3],
        "entries": [
            {
                "user_id": 10,
                "username": "student01",
                "group": "ClassA",
                "total": 240,
                "scores": [80, 80, 80],
                "live_total": 220,
                "post_total": 20,
                "live_scores": [80, 80, 60],
                "post_scores": [0, 0, 20]
            }
        ]
    }
}
```

#### GET `/contests/:id/export` - 导出比赛成绩（管理员）

**认证**: 需要 Bearer Token + 管理员权限
**查询参数**: 同排行榜接口（`board_mode`）

**响应**: `text/csv` 文件下载

#### GET `/users` - 获取用户列表

**认证**: 需要 Bearer Token + 管理员权限

---

#### PUT `/users/:id/role` - 设置用户角色

**认证**: 需要 Bearer Token + 管理员权限

**请求体**:
```json
{
    "role": "admin"   // "user" 或 "admin"
}
```

---

#### GET `/settings/ai` - 获取 AI 设置

**认证**: 需要 Bearer Token + 管理员权限

**成功响应** (200):
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "enabled": true,
        "provider": "deepseek",
        "api_key": "********",    // 隐藏显示
        "api_url": "https://api.deepseek.com/v1/chat/completions",
        "model": "deepseek-chat",
        "timeout": 60
    }
}
```

---

#### PUT `/settings/ai` - 更新 AI 设置

**认证**: 需要 Bearer Token + 管理员权限

**请求体**:
```json
{
    "enabled": true,
    "provider": "deepseek",
    "api_key": "sk-xxx",          // 为空或 "********" 时不更新
    "api_url": "https://api.deepseek.com/v1/chat/completions",
    "model": "deepseek-chat",
    "timeout": 60
}
```

---

#### POST `/settings/ai/test` - 测试 AI 连接

**认证**: 需要 Bearer Token + 管理员权限

**成功响应** (200):
```json
{
    "code": 200,
    "message": "配置有效",
    "data": {
        "provider": "deepseek",
        "model": "deepseek-chat",
        "api_url": "https://api.deepseek.com/v1/chat/completions"
    }
}
```

---

## 4. 数据模型

### 4.1 数据库表结构

#### users 表
| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER | 主键，自增 |
| username | VARCHAR(50) | 用户名，唯一 |
| email | VARCHAR(100) | 邮箱，可选 |
| password_hash | VARCHAR(255) | 密码哈希 |
| student_id | VARCHAR(50) | 学号 |
| role | VARCHAR(20) | 角色：user/admin |
| group | VARCHAR(50) | 分组 |
| solved_count | INTEGER | 解题数 |
| accepted_count | INTEGER | 通过提交总数 |
| submit_count | INTEGER | 提交数 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

#### problems 表
| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER | 主键，自增 |
| title | VARCHAR(200) | 标题 |
| description | TEXT | 描述（Markdown） |
| input_format | TEXT | 输入格式 |
| output_format | TEXT | 输出格式 |
| hint | TEXT | 提示（可选） |
| samples | TEXT | 样例（JSON） |
| time_limit | INTEGER | 时间限制（ms） |
| memory_limit | INTEGER | 内存限制（MB） |
| difficulty | VARCHAR(20) | 难度 |
| tags | TEXT | 标签（JSON） |
| ai_judge_config | TEXT | AI 判题配置（JSON） |
| file_io_enabled | BOOLEAN | 是否启用文件 IO |
| file_input_name | VARCHAR(100) | 输入文件名（如 `data.in`） |
| file_output_name | VARCHAR(100) | 输出文件名（如 `data.out`） |
| is_public | BOOLEAN | 是否公开 |
| created_by | INTEGER | 创建者 ID |
| submit_count | INTEGER | 提交数 |
| accepted_count | INTEGER | 通过数 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

#### testcases 表
| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER | 主键，自增 |
| problem_id | INTEGER | 题目 ID |
| input_file | VARCHAR(255) | 输入文件路径 |
| output_file | VARCHAR(255) | 输出文件路径 |
| score | INTEGER | 分数 |
| is_sample | BOOLEAN | 是否为样例 |
| order_num | INTEGER | 排序序号 |

#### submissions 表
| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER | 主键，自增 |
| problem_id | INTEGER | 题目 ID |
| user_id | INTEGER | 用户 ID |
| language | VARCHAR(20) | 编程语言 |
| code | TEXT | 源代码 |
| status | VARCHAR(30) | 状态 |
| time_used | INTEGER | 用时（ms） |
| memory_used | INTEGER | 内存（KB） |
| score | INTEGER | 得分 |
| testcase_results | TEXT | 测试点结果（JSON） |
| ai_judge_result | TEXT | AI 判题结果（JSON） |
| compile_error | TEXT | 编译错误信息 |
| final_message | TEXT | 最终判定说明 |
| created_at | DATETIME | 提交时间 |

#### contests 表
| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER | 主键，自增 |
| title | VARCHAR(200) | 比赛名称 |
| description | TEXT | 比赛描述 |
| type | VARCHAR(10) | 赛制：oi/ioi |
| timing_mode | VARCHAR(20) | 计时模式：fixed/window |
| duration_minutes | INTEGER | 窗口期个人比赛时长（分钟） |
| start_at | DATETIME | 开始时间 |
| end_at | DATETIME | 结束时间 |
| problem_ids | TEXT | 题目 ID 列表（JSON） |
| allowed_users | TEXT | 允许用户 ID 列表（JSON） |
| allowed_groups | TEXT | 允许分组列表（JSON） |
| is_stats_synced | BOOLEAN | 赛后统计是否已同步 |
| created_by | INTEGER | 创建者 ID |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

#### contest_participations 表
| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER | 主键，自增 |
| contest_id | INTEGER | 比赛 ID |
| user_id | INTEGER | 用户 ID |
| start_at | DATETIME | 个人开始时间 |
| end_at | DATETIME | 个人结束时间 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

#### settings 表
| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER | 主键，自增 |
| key | VARCHAR(100) | 设置键名，唯一 |
| value | TEXT | 设置值 |
| updated_at | DATETIME | 更新时间 |

**常用设置键**:
| 键名 | 说明 |
|------|------|
| `ai_enabled` | AI 判题是否启用 |
| `ai_provider` | AI 提供商（deepseek/openai） |
| `ai_api_key` | API Key |
| `ai_api_url` | API 端点地址 |
| `ai_model` | 模型名称 |
| `ai_timeout` | 超时时间（秒） |

---

## 5. 判题系统

### 5.1 判题流程

```
1. 用户提交代码
       ↓
2. submission_handler.Submit()
   - 验证参数
   - 创建 Submission 记录（状态: Pending）
   - 保存代码文件
   - 调用 judge.SubmitToQueue()
       ↓
3. judge.SubmitToQueue()
   - 获取题目信息
   - 获取测试用例
   - 创建 JudgeTask
   - 加入判题队列
       ↓
4. queue.Worker 取出任务
   - 调用 judger.Handle()
       ↓
5. judger.Handle()
   - 更新状态为 Judging
   - 调用 runTestcases() 执行传统评测
   - 如果启用 AI 判题，调用 aiClient.AnalyzeCode()
   - 综合结果，更新 Submission
   - 清理临时文件
```

**文件操作题目说明**：
当题目启用 `file_io_enabled` 时，评测会将测试输入写入指定输入文件，并从指定输出文件读取结果进行对比。

### 5.2 判题队列 (`judge/queue/queue.go`)

```go
type JudgeTask struct {
    Submission *Submission   // 提交记录
    Problem    *Problem      // 题目信息
    Testcases  []Testcase    // 测试用例
}

type JudgeQueue struct {
    tasks    chan *JudgeTask
    handlers []func(*JudgeTask)
}
```

| 方法 | 说明 |
|------|------|
| `Init(bufferSize int)` | 初始化队列 |
| `GetQueue() *JudgeQueue` | 获取队列实例 |
| `Push(task *JudgeTask)` | 添加任务 |
| `RegisterHandler(handler func(*JudgeTask))` | 注册处理器 |
| `Start(workers int)` | 启动 worker |
| `Stop()` | 停止队列 |

### 5.3 沙箱执行 (`judge/sandbox/sandbox.go`)

#### 语言配置

```go
var languageConfigs = map[string]LanguageConfig{
    "c": {
        SourceFile:  "main.c",
        CompileCmd:  []string{"gcc", "-o", "main", "main.c", "-O2", "-Wall", "-lm", "-std=c11"},
        ExecuteCmd:  []string{"./main"},
        NeedCompile: true,
    },
    "cpp": {
        SourceFile:  "main.cpp",
        CompileCmd:  []string{"g++", "-o", "main", "main.cpp", "-O2", "-Wall", "-std=c++17"},
        ExecuteCmd:  []string{"./main"},
        NeedCompile: true,
    },
    "python": {
        SourceFile:  "main.py",
        ExecuteCmd:  []string{"python3", "main.py"},
        NeedCompile: false,
    },
    "java": {
        SourceFile:  "Main.java",
        CompileCmd:  []string{"javac", "Main.java"},
        ExecuteCmd:  []string{"java", "Main"},
        NeedCompile: true,
    },
    "go": {
        SourceFile:  "main.go",
        CompileCmd:  []string{"go", "build", "-o", "main", "main.go"},
        ExecuteCmd:  []string{"./main"},
        NeedCompile: true,
    },
}
```

#### 执行结果

```go
type ExecuteResult struct {
    Status   string  // 状态
    Time     int     // 运行时间 (ms)
    Memory   int     // 内存使用 (KB)
    Output   string  // 标准输出
    Error    string  // 错误信息
    ExitCode int     // 退出码
}
```

#### 关键函数

| 函数 | 说明 |
|------|------|
| `Execute(workDir, language, code, input string, timeLimit, memoryLimit int) (*ExecuteResult, error)` | 执行代码 |
| `CompareOutput(expected, actual string) bool` | 比较输出（忽略空白差异） |
| `GetWorkDir(submissionID uint) string` | 获取工作目录 |
| `CleanWorkDir(workDir string)` | 清理工作目录 |

### 5.4 判题主逻辑 (`judge/judger.go`)

| 函数 | 说明 |
|------|------|
| `Start(cfg *Config)` | 启动判题服务 |
| `Handle(task *JudgeTask)` | 处理单个判题任务 |
| `runTestcases(submission, problem, testcases)` | 运行所有测试点 |
| `calculateTraditionalStatus(results)` | 计算传统评测状态 |
| `calculateScore(results, allPassed)` | 计算得分 |
| `SubmitToQueue(submission *Submission)` | 提交到队列（供 handler 调用） |

---

## 6. AI 判题系统

### 6.1 配置方式

AI 判题的配置存储在数据库 `settings` 表中，通过管理后台动态配置：

1. 登录管理员账号
2. 进入 **管理后台** → **系统设置**
3. 配置 AI 判题参数：
   - 启用/禁用 AI 判题
   - API Key
   - API 地址
   - 模型名称
   - 超时时间

### 6.2 DeepSeek 客户端 (`judge/ai/deepseek.go`)

```go
type DeepSeekClient struct{}

func NewDeepSeekClient() *DeepSeekClient
```

| 方法 | 说明 |
|------|------|
| `AnalyzeCode(problem *Problem, code, language string) (*AIJudgeResult, error)` | 分析代码 |
| `getSettings() *AISettings` | 从 SettingService 获取当前配置 |

**注意**: 客户端不再在创建时传入配置，而是每次调用时从 `SettingService` 动态获取配置。

### 6.3 API 请求格式

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
            "content": "你是一个代码分析专家..."
        },
        {
            "role": "user",
            "content": "分析代码...(Prompt)"
        }
    ],
    "response_format": {"type": "json_object"},
    "temperature": 0.1
}
```

### 6.4 AI 返回格式

```json
{
    "algorithm_analysis": {
        "detected_algorithms": ["动态规划", "记忆化搜索"],
        "primary_algorithm": "动态规划",
        "confidence": 0.95,
        "evidence": "代码中使用了 dp 数组进行状态转移..."
    },
    "language_features": {
        "language": "C++",
        "used_features": ["vector", "algorithm"],
        "forbidden_features_used": []
    },
    "requirement_check": {
        "algorithm_match": true,
        "language_match": true,
        "all_requirements_met": true
    },
    "summary": "代码使用动态规划算法，符合要求"
}
```

### 6.5 判题结果整合逻辑

```
IF 传统评测未通过:
    返回传统评测结果
ELSE IF AI 判题启用:
    IF AI 判定通过:
        最终结果 = Accepted
    ELSE:
        IF 严格模式:
            最终结果 = Wrong Answer
            final_message = "测试点全部通过，但未满足算法/语言要求"
        ELSE:
            最终结果 = Accepted
            final_message = "测试点通过，AI 提示: ..."
ELSE:
    最终结果 = 传统评测结果

# 评分修正（AI 未满足要求时）
IF AI 判定未通过:
    score = min(score, 50)
```

---

## 7. 前端结构

### 7.1 API 封装 (`src/api/`)

| 文件 | 说明 |
|------|------|
| `request.js` | Axios 实例配置（拦截器、错误处理） |
| `user.js` | 用户相关 API |
| `problem.js` | 题目相关 API |
| `submission.js` | 提交相关 API |
| `rank.js` | 排行榜 API |
| `contest.js` | 比赛 API |
| `statistics.js` | 统计 API |
| `admin.js` | 管理员 API |

### 7.2 状态管理 (`src/stores/`)

#### user.js

| 状态/方法 | 说明 |
|-----------|------|
| `token` | JWT Token |
| `user` | 用户信息 |
| `isLoggedIn` | 是否已登录 |
| `isAdmin` | 是否管理员 |
| `login(credentials)` | 登录 |
| `fetchProfile()` | 获取个人信息 |
| `logout()` | 退出登录 |

### 7.3 路由 (`src/router/index.js`)

| 路由 | 组件 | 权限 |
|------|------|------|
| `/` | Home | 公开 |
| `/problems` | ProblemList | 公开 |
| `/problem/:id` | ProblemDetail | 公开 |
| `/submissions` | SubmissionList | 需登录 |
| `/submission/:id` | SubmissionDetail | 需登录 |
| `/contests` | ContestList | 需登录 |
| `/contest/:id` | ContestDetail | 需登录 |
| `/rank` | Rank | 公开 |
| `/login` | Login | 公开 |
| `/profile` | Profile | 需登录 |
| `/admin/*` | AdminLayout | 需管理员 |
| `/admin/contests` | ContestManage | 需管理员 |
| `/admin/contest/create` | ContestEdit | 需管理员 |
| `/admin/contest/:id/edit` | ContestEdit | 需管理员 |
| `/admin/settings` | Settings | 需管理员 |

### 7.4 主要组件

| 组件 | 路径 | 说明 |
|------|------|------|
| `Navbar` | `components/common/Navbar.vue` | 导航栏 |
| `Footer` | `components/common/Footer.vue` | 页脚 |
| `DifficultyBadge` | `components/problem/DifficultyBadge.vue` | 难度标签 |
| `StatusBadge` | `components/problem/StatusBadge.vue` | 状态标签 |
| `AIJudgeResult` | `components/submission/AIJudgeResult.vue` | AI 判题结果展示 |
| `TestcaseResults` | `components/submission/TestcaseResults.vue` | 测试点结果展示 |
| `MarkdownPreview` | `components/common/MarkdownPreview.vue` | Markdown 预览（支持 LaTeX） |
| `Settings` | `views/admin/Settings.vue` | 系统设置页面 |

### 7.5 当前 UI 设计要点

- 全局采用 Swiss 风格变量体系（`src/styles/global.scss`），统一颜色、字号、边框与容器宽度。
- 导航栏品牌为 `USTC OJ`，使用轻量半透明吸顶布局。
- 列表页（题目、提交、比赛、排行榜）统一表格样式与居中布局。
- 题目详情页采用 splitpanes 分栏：左侧题面，右侧 Monaco 编辑器；支持 `fontSize` 和 `tabSize` 调整。
- 提交详情与比赛详情使用指标仪表盘风格，突出状态、分数、时间与内存信息。
- 管理后台 `ProblemEdit` 支持 Markdown 双栏编辑预览、题面图片上传并按目标字段插入、单文件/Zip 上传进度、整题重测。
- 管理后台 `ContestEdit` 的比赛描述支持与题目管理一致的双栏 Markdown 编辑/预览。
- 比赛详情页在窗口期模式下支持“开始比赛”会话状态展示；管理员排行榜支持 `赛时|赛后 / 赛时 / 赛后` 切换与对应导出。

---

## 8. 调试指南

### 8.1 后端调试

#### 启动命令

```bash
cd backend
go run ./cmd/server -config ./configs/config.yaml
```

#### 调试日志

程序会输出以下日志：
- `[Queue]` - 判题队列日志
- `[Worker-N]` - 判题 Worker 日志
- `[Judger]` - 判题逻辑日志

#### 常见问题

1. **数据库错误**
   - 检查 `config.yaml` 中的 `database.path`
   - 确保目录存在且有写权限

2. **判题失败**
   - 检查编译器是否安装（gcc, g++, python3, javac, go）
   - 检查 `paths.problems` 和 `paths.submissions` 目录

3. **AI 判题不工作**
   - 登录管理后台，检查 **系统设置** 中是否配置了 API Key
   - 检查 AI 判题是否已启用
   - 检查网络是否能访问 AI API

#### 手动测试 API

```bash
# 管理员登录
curl -X POST http://localhost:8080/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 管理员创建用户（需要管理员 Token）
curl -X POST http://localhost:8080/api/v1/admin/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin_token>" \
  -d '{"username":"test","email":"test@test.com","password":"123456","student_id":"2025001"}'

# 普通用户登录
curl -X POST http://localhost:8080/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}'

# 获取题目列表
curl http://localhost:8080/api/v1/problem/list

# 提交代码（需要 Token）
curl -X POST http://localhost:8080/api/v1/submission \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"problem_id":1,"language":"cpp","code":"..."}'

# 整题重测（管理员）
curl -X POST http://localhost:8080/api/v1/problem/1/rejudge \
  -H "Authorization: Bearer <admin_token>"
```

### 8.2 前端调试

#### 启动命令

```bash
cd frontend
npm run dev
```

#### 调试工具

- Vue DevTools - 查看组件状态
- 浏览器 Network 面板 - 查看 API 请求
- Console - 查看错误日志

#### 常见问题

1. **API 请求失败**
   - 检查后端是否运行在 8080 端口
   - 检查 `vite.config.js` 中的 proxy 配置

2. **登录状态丢失**
   - 检查 localStorage 中的 token
   - 检查 Token 是否过期

### 8.3 判题调试

#### 查看判题日志

```bash
# 实时查看后端日志
go run ./cmd/server 2>&1 | grep -E "\[Queue\]|\[Worker\]|\[Judger\]"
```

#### 手动执行代码测试

```bash
# 创建测试目录
mkdir -p ./data/sandbox/test
cd ./data/sandbox/test

# 写入代码
echo '#include <iostream>
int main() {
    int a, b;
    std::cin >> a >> b;
    std::cout << a + b << std::endl;
    return 0;
}' > main.cpp

# 编译
g++ -o main main.cpp

# 运行
echo "1 2" | ./main
```

#### 检查测试数据

```bash
# 查看题目测试数据
ls -la ./data/problems/1/

# 查看输入文件
cat ./data/problems/1/1.in

# 查看输出文件
cat ./data/problems/1/1.out
```

### 8.4 数据库调试

```bash
# 使用 sqlite3 查看数据
sqlite3 ./data/oj.db

# 常用 SQL
.tables                           # 查看所有表
.schema users                     # 查看表结构
SELECT * FROM users;              # 查看用户
SELECT * FROM problems;           # 查看题目
SELECT * FROM submissions WHERE status='Pending';  # 查看待判题
```

---

## 附录：错误码对照表

| HTTP 状态码 | 业务码 | 说明 |
|------------|--------|------|
| 200 | 200 | 成功 |
| 400 | 400 | 参数错误 |
| 401 | 401 | 未登录/Token 无效 |
| 403 | 403 | 无权限 |
| 404 | 404 | 资源不存在 |
| 429 | 429 | 请求过于频繁 |
| 500 | 500 | 服务器错误 |

---

*文档版本: v1.1*
*更新日期: 2026-02-11*
