#!/usr/bin/env python3
"""Collect system pressure metrics from Linux host.

This script is intended to run on the OJ server during load tests.
"""

from __future__ import annotations

import argparse
import csv
import os
import time
from datetime import datetime, timezone
from pathlib import Path
from typing import Dict, Optional, Tuple


def now_iso() -> str:
    return datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z")


def read_proc_stat() -> Dict[str, int]:
    with open("/proc/stat", "r", encoding="utf-8") as f:
        line = f.readline().strip()
    parts = line.split()
    # cpu user nice system idle iowait irq softirq steal ...
    vals = [int(x) for x in parts[1:9]]
    user, nice, system, idle, iowait, irq, softirq, steal = vals
    total = sum(vals)
    busy = user + nice + system + irq + softirq + steal
    return {
        "total": total,
        "busy": busy,
        "idle": idle,
        "iowait": iowait,
    }


def read_meminfo_mb() -> Dict[str, float]:
    info: Dict[str, int] = {}
    with open("/proc/meminfo", "r", encoding="utf-8") as f:
        for line in f:
            if ":" not in line:
                continue
            key, right = line.split(":", 1)
            fields = right.strip().split()
            if not fields:
                continue
            try:
                info[key] = int(fields[0])
            except ValueError:
                continue

    mem_total = info.get("MemTotal", 0)
    mem_avail = info.get("MemAvailable", info.get("MemFree", 0))
    mem_used = max(0, mem_total - mem_avail)

    swap_total = info.get("SwapTotal", 0)
    swap_free = info.get("SwapFree", 0)
    swap_used = max(0, swap_total - swap_free)

    def kb_to_mb(x: int) -> float:
        return x / 1024.0

    return {
        "mem_total_mb": kb_to_mb(mem_total),
        "mem_available_mb": kb_to_mb(mem_avail),
        "mem_used_mb": kb_to_mb(mem_used),
        "mem_used_percent": (mem_used / mem_total * 100.0) if mem_total else 0.0,
        "swap_total_mb": kb_to_mb(swap_total),
        "swap_used_mb": kb_to_mb(swap_used),
        "swap_used_percent": (swap_used / swap_total * 100.0) if swap_total else 0.0,
    }


def read_process_ticks(pid: int) -> Optional[int]:
    stat_path = f"/proc/{pid}/stat"
    if not os.path.exists(stat_path):
        return None
    with open(stat_path, "r", encoding="utf-8") as f:
        raw = f.read().strip()
    # Safe split around ") " because comm can contain spaces.
    rparen = raw.rfind(")")
    if rparen < 0:
        return None
    rest = raw[rparen + 2 :]
    fields = rest.split()
    # fields index after pid/comm/state: utime idx 11, stime idx 12 in rest
    if len(fields) < 13:
        return None
    utime = int(fields[11])
    stime = int(fields[12])
    return utime + stime


def read_process_mem_mb(pid: int) -> Tuple[Optional[float], Optional[float]]:
    status_path = f"/proc/{pid}/status"
    if not os.path.exists(status_path):
        return None, None
    vm_rss_kb = None
    vm_size_kb = None
    with open(status_path, "r", encoding="utf-8") as f:
        for line in f:
            if line.startswith("VmRSS:"):
                parts = line.split()
                if len(parts) >= 2:
                    vm_rss_kb = float(parts[1])
            elif line.startswith("VmSize:"):
                parts = line.split()
                if len(parts) >= 2:
                    vm_size_kb = float(parts[1])
    return (vm_rss_kb / 1024.0 if vm_rss_kb is not None else None, vm_size_kb / 1024.0 if vm_size_kb is not None else None)


def parse_args() -> argparse.Namespace:
    p = argparse.ArgumentParser(description="Collect server CPU/memory/load metrics")
    p.add_argument("--interval-sec", type=float, default=1.0)
    p.add_argument("--duration-sec", type=int, default=7200)
    p.add_argument("--pid", type=int, default=0, help="Optional OJ backend PID")
    p.add_argument("--output", default="stress-testing/output/server_metrics.csv")
    return p.parse_args()


def main() -> None:
    args = parse_args()
    if args.interval_sec <= 0:
        raise SystemExit("interval-sec must be > 0")
    if args.duration_sec <= 0:
        raise SystemExit("duration-sec must be > 0")

    out_path = Path(args.output)
    out_path.parent.mkdir(parents=True, exist_ok=True)

    cpu_count = os.cpu_count() or 1
    ticks_per_sec = os.sysconf(os.sysconf_names["SC_CLK_TCK"])

    fields = [
        "timestamp",
        "load1",
        "load5",
        "load15",
        "cpu_usage_percent",
        "cpu_iowait_percent",
        "mem_total_mb",
        "mem_available_mb",
        "mem_used_mb",
        "mem_used_percent",
        "swap_total_mb",
        "swap_used_mb",
        "swap_used_percent",
        "pid",
        "pid_cpu_percent",
        "pid_rss_mb",
        "pid_vms_mb",
    ]

    start = time.time()
    end = start + args.duration_sec

    prev_cpu = read_proc_stat()
    prev_proc_ticks = read_process_ticks(args.pid) if args.pid else None
    prev_time = time.time()

    print(f"Collecting metrics to {out_path} for {args.duration_sec}s (interval={args.interval_sec}s)")
    with out_path.open("w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=fields)
        writer.writeheader()

        while time.time() < end:
            time.sleep(args.interval_sec)
            now = time.time()
            dt = max(1e-6, now - prev_time)

            cpu_now = read_proc_stat()
            total_delta = max(1, cpu_now["total"] - prev_cpu["total"])
            busy_delta = max(0, cpu_now["busy"] - prev_cpu["busy"])
            iowait_delta = max(0, cpu_now["iowait"] - prev_cpu["iowait"])

            cpu_usage = busy_delta / total_delta * 100.0
            cpu_iowait = iowait_delta / total_delta * 100.0

            mem = read_meminfo_mb()
            load1, load5, load15 = os.getloadavg()

            pid_cpu = ""
            pid_rss = ""
            pid_vms = ""
            if args.pid:
                proc_ticks = read_process_ticks(args.pid)
                rss_mb, vms_mb = read_process_mem_mb(args.pid)
                if proc_ticks is not None and prev_proc_ticks is not None:
                    proc_delta = max(0, proc_ticks - prev_proc_ticks)
                    total_cpu_ticks = dt * ticks_per_sec * cpu_count
                    if total_cpu_ticks > 0:
                        pid_cpu = f"{proc_delta / total_cpu_ticks * 100.0 * cpu_count:.2f}"
                if rss_mb is not None:
                    pid_rss = f"{rss_mb:.2f}"
                if vms_mb is not None:
                    pid_vms = f"{vms_mb:.2f}"
                prev_proc_ticks = proc_ticks

            row = {
                "timestamp": now_iso(),
                "load1": f"{load1:.3f}",
                "load5": f"{load5:.3f}",
                "load15": f"{load15:.3f}",
                "cpu_usage_percent": f"{cpu_usage:.2f}",
                "cpu_iowait_percent": f"{cpu_iowait:.2f}",
                "mem_total_mb": f"{mem['mem_total_mb']:.2f}",
                "mem_available_mb": f"{mem['mem_available_mb']:.2f}",
                "mem_used_mb": f"{mem['mem_used_mb']:.2f}",
                "mem_used_percent": f"{mem['mem_used_percent']:.2f}",
                "swap_total_mb": f"{mem['swap_total_mb']:.2f}",
                "swap_used_mb": f"{mem['swap_used_mb']:.2f}",
                "swap_used_percent": f"{mem['swap_used_percent']:.2f}",
                "pid": args.pid if args.pid else "",
                "pid_cpu_percent": pid_cpu,
                "pid_rss_mb": pid_rss,
                "pid_vms_mb": pid_vms,
            }
            writer.writerow(row)
            f.flush()

            prev_cpu = cpu_now
            prev_time = now

    print("Metrics collection completed.")


if __name__ == "__main__":
    main()
