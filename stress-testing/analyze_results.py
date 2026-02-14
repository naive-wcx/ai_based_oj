#!/usr/bin/env python3
"""Analyze stress test outputs and generate a Markdown report."""

from __future__ import annotations

import argparse
import csv
import math
from collections import Counter, defaultdict
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional


def parse_float(x: str) -> Optional[float]:
    if x is None:
        return None
    x = str(x).strip()
    if not x:
        return None
    try:
        return float(x)
    except ValueError:
        return None


def quantile(values: List[float], q: float) -> float:
    if not values:
        return float("nan")
    arr = sorted(values)
    if len(arr) == 1:
        return arr[0]
    pos = (len(arr) - 1) * q
    lo = int(math.floor(pos))
    hi = int(math.ceil(pos))
    if lo == hi:
        return arr[lo]
    return arr[lo] + (arr[hi] - arr[lo]) * (pos - lo)


def fmt(v: float) -> str:
    if math.isnan(v):
        return "-"
    return f"{v:.2f}"


def load_submission_metrics(path: Path) -> Dict[str, any]:
    submit_ack: List[float] = []
    queue_wait: List[float] = []
    judge_exec: List[float] = []
    judge_done: List[float] = []

    status_counter = Counter()
    http_counter = Counter()
    user_submit_counter = Counter()
    minute_counter = Counter()

    total = 0
    submit_ok = 0
    completed = 0

    with path.open("r", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for row in reader:
            total += 1
            username = (row.get("username") or "").strip()
            if username:
                user_submit_counter[username] += 1

            ts = (row.get("submit_start_at") or row.get("timestamp") or "").strip()
            if ts:
                key = ts[:16]  # UTC minute: YYYY-MM-DDTHH:MM
                minute_counter[key] += 1

            http_status = str(row.get("http_status", "")).strip()
            api_code = str(row.get("api_code", "")).strip()
            http_counter[(http_status, api_code)] += 1

            ack = parse_float(row.get("submit_ack_ms", ""))
            if ack is not None:
                submit_ack.append(ack)

            qw = parse_float(row.get("queue_wait_ms", ""))
            if qw is not None:
                queue_wait.append(qw)

            je = parse_float(row.get("judge_exec_ms", ""))
            if je is not None:
                judge_exec.append(je)

            jd = parse_float(row.get("judge_done_ms", ""))
            if jd is not None:
                judge_done.append(jd)
                completed += 1

            final_status = (row.get("final_status") or "").strip()
            if final_status:
                status_counter[final_status] += 1

            if http_status == "200" and api_code == "200":
                submit_ok += 1

    return {
        "total": total,
        "submit_ok": submit_ok,
        "completed": completed,
        "submit_ack": submit_ack,
        "queue_wait": queue_wait,
        "judge_exec": judge_exec,
        "judge_done": judge_done,
        "status_counter": status_counter,
        "http_counter": http_counter,
        "user_submit_counter": user_submit_counter,
        "minute_counter": minute_counter,
    }


def load_server_metrics(path: Path) -> Dict[str, List[float]]:
    cols = defaultdict(list)
    keys = [
        "cpu_usage_percent",
        "cpu_iowait_percent",
        "mem_used_percent",
        "swap_used_percent",
        "load1",
        "pid_cpu_percent",
        "pid_rss_mb",
        "pid_vms_mb",
    ]
    with path.open("r", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for row in reader:
            for k in keys:
                v = parse_float(row.get(k, ""))
                if v is not None:
                    cols[k].append(v)
    return cols


def stat_line(values: List[float], unit: str = "") -> str:
    if not values:
        return "-"
    p50 = quantile(values, 0.50)
    p90 = quantile(values, 0.90)
    p95 = quantile(values, 0.95)
    p99 = quantile(values, 0.99)
    m = max(values)
    suffix = f" {unit}" if unit else ""
    return f"P50={fmt(p50)}{suffix}, P90={fmt(p90)}{suffix}, P95={fmt(p95)}{suffix}, P99={fmt(p99)}{suffix}, Max={fmt(m)}{suffix}"


def parse_args() -> argparse.Namespace:
    p = argparse.ArgumentParser(description="Analyze stress test output CSV files")
    p.add_argument("--submission-csv", required=True)
    p.add_argument("--server-csv", default="")
    p.add_argument("--output", default="stress-testing/output/report.md")
    p.add_argument("--title", default="OJ Stress Test Report")
    return p.parse_args()


def main() -> None:
    args = parse_args()
    submission_csv = Path(args.submission_csv)
    if not submission_csv.exists():
        raise SystemExit(f"submission csv not found: {submission_csv}")

    sub = load_submission_metrics(submission_csv)

    report_lines: List[str] = []
    report_lines.append(f"# {args.title}")
    report_lines.append("")
    report_lines.append(f"- Generated at: {datetime.utcnow().isoformat(timespec='seconds')}Z")
    report_lines.append(f"- Submission CSV: `{submission_csv}`")
    if args.server_csv:
        report_lines.append(f"- Server CSV: `{args.server_csv}`")
    report_lines.append("")

    report_lines.append("## 1. Submission Overview")
    report_lines.append("")
    report_lines.append(f"- Total submissions attempted: **{sub['total']}**")
    report_lines.append(f"- Submit API success count: **{sub['submit_ok']}**")
    report_lines.append(f"- Judged/finished count: **{sub['completed']}**")
    if sub["total"] > 0:
        ok_rate = sub["submit_ok"] / sub["total"] * 100.0
        report_lines.append(f"- Submit success rate: **{ok_rate:.2f}%**")
    report_lines.append("")

    report_lines.append("## 2. Latency Metrics")
    report_lines.append("")
    report_lines.append(f"- Submit ACK latency (ms): {stat_line(sub['submit_ack'], 'ms')}")
    report_lines.append(f"- Queue wait latency (ms): {stat_line(sub['queue_wait'], 'ms')}")
    report_lines.append(f"- Judge execution latency (ms): {stat_line(sub['judge_exec'], 'ms')}")
    report_lines.append(f"- Judge done latency (ms): {stat_line(sub['judge_done'], 'ms')}")
    report_lines.append("")

    report_lines.append("## 3. Status Distribution")
    report_lines.append("")
    if sub["status_counter"]:
        for k, v in sub["status_counter"].most_common():
            report_lines.append(f"- {k}: {v}")
    else:
        report_lines.append("- No final status captured")
    report_lines.append("")

    report_lines.append("## 4. Submit API Result Distribution")
    report_lines.append("")
    for (http_status, api_code), cnt in sorted(sub["http_counter"].items(), key=lambda x: (-x[1], x[0])):
        report_lines.append(f"- HTTP {http_status}, API code {api_code}: {cnt}")
    report_lines.append("")

    report_lines.append("## 5. Load Pattern")
    report_lines.append("")
    if sub["minute_counter"]:
        peak_minute, peak_count = max(sub["minute_counter"].items(), key=lambda x: x[1])
        avg_per_min = sum(sub["minute_counter"].values()) / max(1, len(sub["minute_counter"]))
        report_lines.append(f"- Peak submissions/minute: **{peak_count}** at `{peak_minute}`")
        report_lines.append(f"- Average submissions/minute (active minutes): **{avg_per_min:.2f}**")
    else:
        report_lines.append("- No minute-level data")

    if sub["user_submit_counter"]:
        per_user = list(sub["user_submit_counter"].values())
        report_lines.append(f"- Per-user submissions avg: **{sum(per_user)/len(per_user):.2f}**")
        report_lines.append(f"- Per-user submissions min/max: **{min(per_user)} / {max(per_user)}**")
    report_lines.append("")

    if args.server_csv:
        server_csv = Path(args.server_csv)
        if server_csv.exists():
            sm = load_server_metrics(server_csv)
            report_lines.append("## 6. Server Pressure")
            report_lines.append("")
            report_lines.append(f"- CPU usage: {stat_line(sm['cpu_usage_percent'], '%')}")
            report_lines.append(f"- CPU iowait: {stat_line(sm['cpu_iowait_percent'], '%')}")
            report_lines.append(f"- Memory used: {stat_line(sm['mem_used_percent'], '%')}")
            report_lines.append(f"- Swap used: {stat_line(sm['swap_used_percent'], '%')}")
            report_lines.append(f"- Load1: {stat_line(sm['load1'])}")
            if sm["pid_cpu_percent"]:
                report_lines.append(f"- Backend PID CPU: {stat_line(sm['pid_cpu_percent'], '%')}")
            if sm["pid_rss_mb"]:
                report_lines.append(f"- Backend PID RSS: {stat_line(sm['pid_rss_mb'], 'MB')}")
            if sm["pid_vms_mb"]:
                report_lines.append(f"- Backend PID VMS: {stat_line(sm['pid_vms_mb'], 'MB')}")
            report_lines.append("")
        else:
            report_lines.append("## 6. Server Pressure")
            report_lines.append("")
            report_lines.append(f"- Server CSV not found: `{server_csv}`")
            report_lines.append("")

    report_lines.append("## 7. Suggested Interpretation")
    report_lines.append("")
    report_lines.append("- Compare this report across different `judge.workers` values (e.g. 1/2/3).")
    report_lines.append("- Keep the highest worker count that still maintains stable P95 `judge_done_ms` and low error growth.")
    report_lines.append("- If CPU > 85% for long periods, memory/swap pressure increases, or timeout/error rate climbs, reduce workers.")

    out_path = Path(args.output)
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text("\n".join(report_lines), encoding="utf-8")
    print(f"Report written to: {out_path}")


if __name__ == "__main__":
    main()
