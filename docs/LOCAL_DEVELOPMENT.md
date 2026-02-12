# 本地开发环境搭建指南（WSL）

本文档假设你使用的是全新的 WSL Ubuntu 环境。

---

## 一、安装基础依赖

打开 WSL 终端，依次执行以下命令：

### 1.1 更新系统

```bash
sudo apt update && sudo apt upgrade -y
```

### 1.2 安装编译工具（用于判题）

```bash
# 安装 GCC/G++
sudo apt install -y build-essential

# 安装 Python3
sudo apt install -y python3 python3-pip

# 安装 Java（可选，如果需要支持 Java 判题）
sudo apt install -y default-jdk

# 验证安装
gcc --version
g++ --version
python3 --version
```

### 1.3 安装 Go 语言（后端）

```bash
# 下载 Go 1.21
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz

# 解压到 /usr/local
sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz

# 配置环境变量
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc

# 验证安装
go version
# 应该输出: go version go1.21.6 linux/amd64

# 清理下载文件
rm go1.21.6.linux-amd64.tar.gz
```

### 1.4 安装 Node.js（前端）

```bash
# 安装 nvm (Node Version Manager)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash

# 重新加载配置
source ~/.bashrc

# 安装 Node.js 18 LTS
nvm install 18
nvm use 18

# 验证安装
node --version
# 应该输出: v18.x.x

npm --version
# 应该输出: 10.x.x
```

### 1.5 安装 SQLite（数据库）

```bash
sudo apt install -y sqlite3
```

---

## 二、获取项目代码

### 2.1 进入项目目录

由于你的项目在 Windows 的 D 盘，在 WSL 中的路径是：

```bash
cd /mnt/d/课程资料/25秋季/OJ
```

**提示**：路径中有中文可能会有问题，建议创建一个软链接：

```bash
# 创建软链接到 home 目录
ln -s "/mnt/d/课程资料/25秋季/OJ" ~/oj

# 以后可以直接使用
cd ~/oj
```

---

## 三、启动后端

### 3.1 进入后端目录

```bash
cd ~/oj/backend
```

### 3.2 下载 Go 依赖

```bash
go mod tidy
```

这会下载所有需要的 Go 包，第一次可能需要几分钟。

### 3.3 创建环境变量文件（可选）

```bash
# 创建 .env 文件
cat > .env << 'EOF'
JWT_SECRET=my_secret_key_for_development
EOF
```

**注意**：
- `JWT_SECRET`：开发环境随便填一个字符串即可
- **AI 判题的 API Key 不再需要在这里配置**，可以在管理后台动态设置

### 3.4 启动后端

```bash
# 加载环境变量（如果创建了 .env）
export $(cat .env 2>/dev/null | xargs)

# 启动后端服务
go run ./cmd/server
```

**成功启动后你会看到**：
```
配置加载成功
数据库初始化成功
已创建默认管理员账号: admin / admin123
[Judger] 判题服务已启动
[Queue] 启动判题队列，workers=2
[Worker-0] 启动
[Worker-1] 启动
[GIN-debug] Listening and serving HTTP on :8080
```

后端现在运行在 `http://localhost:8080`

**保持这个终端窗口开启，打开一个新的 WSL 终端继续操作。**

---

## 四、启动前端

### 4.1 打开新终端，进入前端目录

```bash
cd ~/oj/frontend
```

### 4.2 安装 npm 依赖

```bash
npm install
```

第一次安装可能需要几分钟。

### 4.3 启动开发服务器

```bash
npm run dev
```

**成功启动后你会看到**：
```
  VITE v5.x.x  ready in xxx ms

  ➜  Local:   http://localhost:3000/
  ➜  Network: use --host to expose
  ➜  press h + enter to show help
```

前端现在运行在 `http://localhost:3000`

---

## 五、访问测试

### 5.1 打开浏览器

在 Windows 浏览器中访问：

```
http://localhost:3000
```

你应该能看到 OJ 系统的首页。

### 5.2 测试登录

使用默认管理员账号登录：
- 用户名：`admin`
- 密码：`admin123`

普通用户账号请在 **管理后台 → 用户管理** 中由管理员创建后分配给学生。

**修改密码**：
登录后进入 **个人中心** → **修改密码**。

### 5.3 配置 AI 判题（可选）

如果你需要使用 AI 智能判题功能：

1. 登录管理员账号后，点击右上角头像 → "管理后台"
2. 点击左侧 **"系统设置"**
3. 在 AI 判题设置中：
   - 开启 "启用 AI 判题"
   - 填入你的 **DeepSeek API Key**（可在 https://platform.deepseek.com/ 获取）
   - 点击 "保存设置"

### 5.4 创建测试题目

1. 点击 "题目管理" → "创建题目"
3. 填写题目信息，例如：

```
标题: A+B Problem
描述: 输入两个整数 a 和 b，输出它们的和。
输入格式: 两个整数 a 和 b，用空格分隔
输出格式: 一个整数，表示 a+b 的值
样例输入: 1 2
样例输出: 3
时间限制: 1000
内存限制: 256
难度: 简单
```

4. 点击 "创建题目"

### 5.5 上传测试数据

创建题目后，在编辑页面：

1. 创建测试输入文件 `1.in`：
```
1 2
```

2. 创建测试输出文件 `1.out`：
```
3
```

3. 在 "测试数据管理" 区域上传这两个文件

补充说明：
- 支持 **单文件上传** 与 **Zip 批量覆盖上传**（Zip 需包含成对 `.in` + `.out/.ans`）。
- 上传过程中可看到进度条；上传完成后可继续选择下一组文件，无需刷新页面。
- 若更新了测试点，可在题目编辑页使用 **整题重测** 将该题历史提交重新入队评测。
- 支持 **题面图片上传**：在题目编辑页可上传 `png/jpg/jpeg/gif/webp/bmp`，并选择插入到题目描述/输入格式/输出格式/提示。
- 上传成功后会返回并自动插入 Markdown 图片语法，后端图片地址为 `/api/v1/problem/:id/image/:filename`。

### 5.6 提交代码测试

1. 进入题目列表，点击刚创建的题目
2. 在右侧代码框输入：

**C++ 代码**：
```cpp
#include <iostream>
using namespace std;

int main() {
    int a, b;
    cin >> a >> b;
    cout << a + b << endl;
    return 0;
}
```

3. 点击 "提交"
4. 页面会跳转到提交详情，等待几秒后刷新查看结果

### 5.7 比赛窗口期模式（可选）

如需验证窗口期 + 个人时长模式：

1. 管理后台创建比赛时将 **计时模式** 设为 `窗口期 + 个人固定时长`
2. 设置窗口期 `开始时间/结束时间`，并配置 `个人比赛时长（分钟）`
3. 普通用户进入比赛详情后点击 **开始比赛**
4. 管理员在比赛详情可切换排行榜模式：`赛时|赛后`、`赛时`、`赛后`

---

## 六、快速启动脚本

为了方便，你可以创建快速启动脚本：

### 6.1 后端启动脚本

```bash
cat > ~/oj/start-backend.sh << 'EOF'
#!/bin/bash
cd ~/oj/backend
export $(cat .env 2>/dev/null | xargs)
go run ./cmd/server
EOF

chmod +x ~/oj/start-backend.sh
```

### 6.2 前端启动脚本

```bash
cat > ~/oj/start-frontend.sh << 'EOF'
#!/bin/bash
cd ~/oj/frontend
npm run dev
EOF

chmod +x ~/oj/start-frontend.sh
```

### 6.3 使用方式

以后启动只需要：

**终端 1（后端）**：
```bash
~/oj/start-backend.sh
```

**终端 2（前端）**：
```bash
~/oj/start-frontend.sh
```

---

## 七、常见问题排查

### 7.1 后端启动失败

**问题**: `go: command not found`
```bash
# 重新加载环境变量
source ~/.bashrc
```

**问题**: 依赖下载失败
```bash
# 设置 Go 代理（国内）
go env -w GOPROXY=https://goproxy.cn,direct
go mod tidy
```

**问题**: 数据库权限错误
```bash
# 确保数据目录可写
mkdir -p ~/oj/backend/data
chmod 755 ~/oj/backend/data
```

### 7.2 前端启动失败

**问题**: `npm: command not found`
```bash
source ~/.bashrc
```

**问题**: 依赖安装失败
```bash
# 清除缓存重试
rm -rf node_modules package-lock.json
npm install
```

### 7.3 判题失败

**问题**: 编译错误 - 找不到编译器
```bash
# 确认编译器已安装
which gcc g++ python3

# 如果没有，安装它们
sudo apt install -y build-essential python3
```

**问题**: 运行超时
- 检查代码是否有死循环
- 检查时间限制设置是否合理

### 7.4 页面打不开

**问题**: `localhost:3000` 无法访问
- 确认前端是否正在运行
- 检查是否有防火墙阻止

**问题**: API 请求失败
- 确认后端是否运行在 8080 端口
- 查看后端终端的错误日志

---

## 八、开发调试技巧

### 8.1 查看后端日志

后端会在终端输出详细日志，包括：
- API 请求日志
- 判题过程日志

### 8.2 查看数据库

```bash
cd ~/oj/backend
sqlite3 ./data/oj.db

# 常用命令
.tables                    -- 查看所有表
SELECT * FROM users;       -- 查看用户
SELECT * FROM problems;    -- 查看题目
SELECT * FROM submissions; -- 查看提交
.quit                      -- 退出
```

### 8.3 手动测试 API

```bash
# 测试后端是否正常
curl http://localhost:8080/health

# 测试登录
curl -X POST http://localhost:8080/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

### 8.4 热重载

- **前端**：已内置热重载，修改代码后自动刷新
- **后端**：需要重启才能生效（Ctrl+C 停止，然后重新 `go run`）

---

## 九、完整命令速查

```bash
# ===== 一次性安装（只需执行一次）=====
sudo apt update && sudo apt upgrade -y
sudo apt install -y build-essential python3 sqlite3

# 安装 Go
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
source ~/.bashrc

# 安装 Node.js
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
source ~/.bashrc
nvm install 18

# ===== 每次开发时执行 =====
# 终端 1: 启动后端
cd /mnt/d/课程资料/25秋季/OJ/backend
export JWT_SECRET=dev_secret
go run ./cmd/server

# 终端 2: 启动前端
cd /mnt/d/课程资料/25秋季/OJ/frontend
npm run dev

# 浏览器访问: http://localhost:3000
# 管理员账号: admin / admin123
# AI API Key 可在管理后台 → 系统设置中配置
```

---

*祝开发顺利！*
