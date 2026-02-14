# OJ 压力测试 Workflow（IOI 考试场景）

本目录用于模拟约 50 人考试场景，记录每次提交反馈时间，并同步采集服务器压力指标。

说明：系统中的“组别”不是独立资源，而是用户字段 `group`。本 workflow 通过批量创建用户并赋值 `group` 来实现建组。

## 1. 前置条件

1. OJ 服务可访问（例如 `http://127.0.0.1:8080`）。
2. 至少有 1 个管理员账号（用于建用户、建比赛）。
3. 已有可用题目 ID（例如 `1,2,3`）。
4. 建议在**独立测试环境**执行，避免影响正式数据。

安装依赖：

```bash
python3 -m venv .venv
source .venv/bin/activate
pip install -r stress-testing/requirements.txt
```

## 2. 创建账号、组别和比赛

使用 `prepare_exam.py` 一次性完成：

```bash
python3 stress-testing/prepare_exam.py \
  --base-url http://127.0.0.1:8080 \
  --admin-username admin \
  --admin-password '你的管理员密码' \
  --user-count 50 \
  --user-prefix examstu \
  --user-password 'Pass123456' \
  --group-count 2 \
  --contest-title 'IOI 模拟压测' \
  --contest-type ioi \
  --timing-mode fixed \
  --start-delay-minutes 3 \
  --contest-duration-minutes 120 \
  --problem-ids 1,2,3
```

脚本输出目录：`stress-testing/output/<run_tag>/`

关键产物：

1. `users.csv`：压测账号和密码。
2. `contest.json`：比赛详情（含 `id`）。
3. `prepare_result.json`：本次准备过程摘要。

## 3. 采集服务器压力指标

在 OJ 服务器上执行（建议单独终端）：

1. 找后端 PID：

```bash
pgrep -f 'oj-system|backend|server' | head -n 1
```

2. 启动采集（示例 2.5 小时）：

```bash
python3 stress-testing/collect_server_metrics.py \
  --interval-sec 1 \
  --duration-sec 9000 \
  --pid <后端PID> \
  --output stress-testing/output/<run_tag>/server_metrics.csv
```

如果压测机和 OJ 服务器不是同一台，请在服务器上单独运行该脚本。

## 4. 运行考试风格压测（低压力、真实节奏）

默认节奏是“前期稀疏，临近结束变密集”：

1. 常规期（前 80%）：每人约 12~30 分钟提交一次。
2. 临近期（80%~95%）：每人约 3~8 分钟提交一次。
3. 冲刺期（最后 5%）：每人约 40~120 秒提交一次。

执行命令：

```bash
python3 stress-testing/run_exam_load.py \
  --base-url http://127.0.0.1:8080 \
  --contest-id <contest_id> \
  --users-file stress-testing/output/<run_tag>/users.csv \
  --run-tag <run_tag> \
  --duration-minutes 120 \
  --normal-interval-sec 720,1800 \
  --rush-interval-sec 180,480 \
  --sprint-interval-sec 40,120 \
  --language-weights cpp:0.7,python:0.2,java:0.05,go:0.05 \
  --poll-interval-sec 1.5 \
  --poll-max-wait-sec 240
```

输出文件：`stress-testing/output/<run_tag>/submission_metrics.csv`

记录字段包括：

1. `submit_ack_ms`：提交接口响应耗时。
2. `queue_wait_ms`：提交成功到进入 `Judging` 的时间。
3. `judge_exec_ms`：`Judging` 到最终状态耗时。
4. `judge_done_ms`：提交成功到最终状态总耗时。
5. `final_status`、`http_status`、`error_message`。

## 5. 生成分析报告

```bash
python3 stress-testing/analyze_results.py \
  --submission-csv stress-testing/output/<run_tag>/submission_metrics.csv \
  --server-csv stress-testing/output/<run_tag>/server_metrics.csv \
  --output stress-testing/output/<run_tag>/report.md \
  --title 'IOI 50人考试压测报告'
```

报告会给出：

1. 提交成功率、完成评测数。
2. `submit_ack_ms / queue_wait_ms / judge_exec_ms / judge_done_ms` 的 P50/P90/P95/P99/Max。
3. 状态分布、每分钟峰值提交量、每人提交量分布。
4. CPU/内存/Swap/Load 等服务器压力统计（若提供服务器 CSV）。

## 6. 如何据此确定评测并发

建议固定同一场景，分别测试 `judge.workers=1/2/3`，每轮都走完整 workflow。

对比重点：

1. P95 `judge_done_ms` 是否显著变差。
2. 提交失败、评测超时、异常状态是否增加。
3. CPU 是否长期 > 85%，内存/Swap 是否持续吃紧。

选择“满足稳定性前提下的最大 workers”。

## 7. 常见问题

1. 登录失败：确认管理员或学生密码正确。
2. 比赛不可见：检查 `start_at` 是否已到，以及 `allowed_groups` 是否包含这些账号组。
3. 429 过多：适当加大各阶段间隔，尤其冲刺期间隔。
4. 没有 `server_metrics.csv`：说明未在服务器侧启动采集脚本。
