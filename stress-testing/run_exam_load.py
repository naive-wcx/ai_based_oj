#!/usr/bin/env python3
"""Run realistic IOI exam load simulation.

- Simulates many students with long intervals in normal phase.
- In final phase, submission interval shortens to mimic end-of-exam bursts.
- Records per-submission latency and outcome to CSV.
"""

from __future__ import annotations

import argparse
import asyncio
import csv
import json
import random
import time
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

try:
    import aiohttp
except ModuleNotFoundError as exc:  # pragma: no cover
    raise SystemExit(
        "Missing dependency: aiohttp. Run `pip install -r stress-testing/requirements.txt` first."
    ) from exc


DEFAULT_TEMPLATES = {
    "cpp": r'''#include <bits/stdc++.h>
using namespace std;
int main() {
    ios::sync_with_stdio(false);
    cin.tie(nullptr);
    long long x = 0;
    if (!(cin >> x)) {
        cout << 0 << "\n";
        return 0;
    }
    cout << x << "\n";
    return 0;
}
''',
    "c": r'''#include <stdio.h>
int main(void) {
    long long x = 0;
    if (scanf("%lld", &x) != 1) {
        printf("0\n");
        return 0;
    }
    printf("%lld\n", x);
    return 0;
}
''',
    "python": r'''import sys
raw = sys.stdin.read().strip().split()
print(raw[0] if raw else 0)
''',
    "java": r'''import java.io.*;
import java.util.*;

public class Main {
    public static void main(String[] args) throws Exception {
        Scanner sc = new Scanner(System.in);
        if (sc.hasNextLong()) {
            System.out.println(sc.nextLong());
        } else {
            System.out.println(0);
        }
    }
}
''',
    "go": r'''package main

import "fmt"

func main() {
    var x int64
    if _, err := fmt.Scan(&x); err != nil {
        fmt.Println(0)
        return
    }
    fmt.Println(x)
}
''',
}


@dataclass
class UserCred:
    username: str
    password: str
    group: str


class ApiError(Exception):
    pass


def now_iso() -> str:
    return datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z")


def normalize_base_url(url: str) -> str:
    url = url.rstrip("/")
    if not url.endswith("/api/v1"):
        url = f"{url}/api/v1"
    return url


def parse_pair(text: str, flag: str) -> Tuple[float, float]:
    parts = [p.strip() for p in text.split(",") if p.strip()]
    if len(parts) != 2:
        raise ValueError(f"{flag} should be 'min,max'")
    lo = float(parts[0])
    hi = float(parts[1])
    if lo <= 0 or hi <= 0 or lo > hi:
        raise ValueError(f"{flag} invalid range")
    return lo, hi


def parse_weights(text: str) -> Dict[str, float]:
    items = [x.strip() for x in text.split(",") if x.strip()]
    weights: Dict[str, float] = {}
    for item in items:
        if ":" not in item:
            raise ValueError(f"Invalid weight item: {item}")
        lang, raw = item.split(":", 1)
        lang = lang.strip()
        val = float(raw.strip())
        if val < 0:
            raise ValueError(f"Negative weight: {item}")
        weights[lang] = val
    total = sum(weights.values())
    if total <= 0:
        raise ValueError("Language weights total must be > 0")
    return {k: v / total for k, v in weights.items()}


def choose_language(weights: Dict[str, float], rng: random.Random) -> str:
    x = rng.random()
    acc = 0.0
    last = "cpp"
    for lang, w in weights.items():
        acc += w
        last = lang
        if x <= acc:
            return lang
    return last


def choose_interval(elapsed_ratio: float, cfg: argparse.Namespace, rng: random.Random) -> float:
    if elapsed_ratio < cfg.normal_until:
        return rng.uniform(cfg.normal_min_sec, cfg.normal_max_sec)
    if elapsed_ratio < cfg.rush_until:
        return rng.uniform(cfg.rush_min_sec, cfg.rush_max_sec)
    return rng.uniform(cfg.sprint_min_sec, cfg.sprint_max_sec)


async def api_request(
    session: aiohttp.ClientSession,
    method: str,
    url: str,
    token: Optional[str] = None,
    json_body: Any = None,
    timeout: float = 30.0,
) -> Tuple[int, Optional[Dict[str, Any]], str]:
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    try:
        async with session.request(method, url, headers=headers, json=json_body, timeout=timeout) as resp:
            status = resp.status
            text = await resp.text()
            data = None
            err = ""
            if text:
                try:
                    data = json.loads(text)
                except json.JSONDecodeError:
                    err = f"non-json response: {text[:200]}"
            if status != 200:
                msg = err
                if data and isinstance(data, dict):
                    msg = data.get("message") or msg
                return status, data, msg or f"HTTP {status}"
            return status, data, ""
    except asyncio.TimeoutError:
        return 599, None, "request timeout"
    except Exception as exc:
        return 598, None, str(exc)


async def login(session: aiohttp.ClientSession, base_url: str, username: str, password: str) -> str:
    status, data, err = await api_request(
        session,
        "POST",
        f"{base_url}/user/login",
        json_body={"username": username, "password": password},
    )
    if status != 200 or not data or data.get("code") != 200:
        raise ApiError(f"login failed for {username}: {err or (data or {}).get('message', 'unknown')}")
    token = ((data.get("data") or {}).get("token"))
    if not token:
        raise ApiError(f"login failed for {username}: token missing")
    return token


async def wait_for_contest_access(
    session: aiohttp.ClientSession,
    base_url: str,
    contest_id: int,
    token: str,
    timeout_sec: int,
) -> Dict[str, Any]:
    deadline = time.time() + timeout_sec
    last_msg = ""
    while time.time() < deadline:
        status, data, err = await api_request(
            session,
            "GET",
            f"{base_url}/contest/{contest_id}",
            token=token,
        )
        if status == 200 and data and data.get("code") == 200:
            return data
        last_msg = err or ((data or {}).get("message") if isinstance(data, dict) else "") or "contest not ready"
        await asyncio.sleep(5)
    raise ApiError(f"contest not accessible in time: {last_msg}")


async def poll_submission_done(
    session: aiohttp.ClientSession,
    base_url: str,
    token: str,
    submission_id: int,
    poll_interval_sec: float,
    max_wait_sec: int,
) -> Dict[str, Any]:
    start = time.time()
    first_judging_at = None
    last_status = ""
    last_score = None
    last_msg = ""

    while time.time() - start <= max_wait_sec:
        status, data, err = await api_request(
            session,
            "GET",
            f"{base_url}/submission/{submission_id}",
            token=token,
        )
        if status != 200 or not data or data.get("code") != 200:
            last_msg = err or ((data or {}).get("message") if isinstance(data, dict) else "poll failed")
            await asyncio.sleep(poll_interval_sec)
            continue

        body = data.get("data") or {}
        s = body.get("status", "")
        last_status = s
        last_score = body.get("score")
        if s == "Judging" and first_judging_at is None:
            first_judging_at = time.time()

        if s not in ("Pending", "Judging"):
            return {
                "done": True,
                "first_judging_at": first_judging_at,
                "finished_at": time.time(),
                "final_status": s,
                "final_score": last_score,
                "error": "",
            }

        await asyncio.sleep(poll_interval_sec)

    return {
        "done": False,
        "first_judging_at": first_judging_at,
        "finished_at": time.time(),
        "final_status": last_status or "Timeout",
        "final_score": last_score,
        "error": last_msg or "poll timeout",
    }


async def run_user(
    user: UserCred,
    idx: int,
    cfg: argparse.Namespace,
    base_url: str,
    writer: csv.DictWriter,
    output_file,
    write_lock: asyncio.Lock,
    counters: Dict[str, int],
    counters_lock: asyncio.Lock,
) -> None:
    rng = random.Random(cfg.seed + idx)
    conn = aiohttp.TCPConnector(limit=cfg.http_conn_limit, ssl=False)
    timeout = aiohttp.ClientTimeout(total=60)
    async with aiohttp.ClientSession(connector=conn, timeout=timeout) as session:
        try:
            token = await login(session, base_url, user.username, user.password)
        except Exception as exc:
            async with counters_lock:
                counters["login_failed"] += 1
            print(f"[ERR] {user.username}: {exc}")
            return

        try:
            contest_data = await wait_for_contest_access(
                session,
                base_url,
                cfg.contest_id,
                token,
                timeout_sec=cfg.contest_ready_timeout_sec,
            )
        except Exception as exc:
            async with counters_lock:
                counters["contest_access_failed"] += 1
            print(f"[ERR] {user.username}: {exc}")
            return

        problems = (contest_data.get("data") or {}).get("problems") or []
        problem_ids = [int(p.get("id")) for p in problems if p.get("id") is not None]
        if not problem_ids:
            async with counters_lock:
                counters["contest_access_failed"] += 1
            print(f"[ERR] {user.username}: contest {cfg.contest_id} has no visible problems")
            return

        run_start = cfg.run_start_epoch
        run_end = cfg.run_end_epoch
        next_submit = run_start + rng.uniform(0, cfg.initial_spread_sec)
        submitted = 0

        while True:
            now = time.time()
            if now >= run_end:
                break
            sleep_sec = max(0.0, next_submit - now)
            if sleep_sec > 0:
                await asyncio.sleep(sleep_sec)
            if time.time() >= run_end:
                break
            if cfg.max_submissions_per_user > 0 and submitted >= cfg.max_submissions_per_user:
                break

            elapsed_ratio = (time.time() - run_start) / max(1.0, (run_end - run_start))
            lang = choose_language(cfg.lang_weights, rng)
            code = DEFAULT_TEMPLATES.get(lang) or DEFAULT_TEMPLATES["cpp"]
            problem_id = rng.choice(problem_ids)

            submit_start = time.time()
            submit_start_iso = now_iso()
            status, data, err = await api_request(
                session,
                "POST",
                f"{base_url}/submission",
                token=token,
                json_body={"problem_id": problem_id, "language": lang, "code": code},
            )
            submit_ack_at = time.time()
            submit_ack_iso = now_iso()
            submit_ack_ms = (submit_ack_at - submit_start) * 1000.0

            api_code = ""
            submission_id = ""
            final_status = ""
            final_score = ""
            queue_wait_ms = ""
            judge_exec_ms = ""
            judge_done_ms = ""
            finished_iso = ""
            first_judging_iso = ""
            error_message = ""

            if data and isinstance(data, dict):
                api_code = data.get("code", "")

            if status == 200 and data and data.get("code") == 200:
                submission = data.get("data") or {}
                sid = submission.get("id")
                if sid is not None:
                    submission_id = str(sid)
                    poll_res = await poll_submission_done(
                        session,
                        base_url,
                        token,
                        int(sid),
                        poll_interval_sec=cfg.poll_interval_sec,
                        max_wait_sec=cfg.poll_max_wait_sec,
                    )
                    finished_at = poll_res["finished_at"]
                    finished_iso = datetime.fromtimestamp(finished_at, tz=timezone.utc).replace(microsecond=0).isoformat().replace(
                        "+00:00", "Z"
                    )
                    first_judging_at = poll_res.get("first_judging_at")
                    if first_judging_at:
                        first_judging_iso = (
                            datetime.fromtimestamp(first_judging_at, tz=timezone.utc)
                            .replace(microsecond=0)
                            .isoformat()
                            .replace("+00:00", "Z")
                        )
                        queue_wait_ms = f"{(first_judging_at - submit_ack_at) * 1000.0:.2f}"
                        judge_exec_ms = f"{(finished_at - first_judging_at) * 1000.0:.2f}"
                    judge_done_ms = f"{(finished_at - submit_ack_at) * 1000.0:.2f}"
                    final_status = poll_res.get("final_status", "")
                    fs = poll_res.get("final_score")
                    final_score = "" if fs is None else str(fs)
                    error_message = poll_res.get("error", "")
                else:
                    error_message = "submission id missing"
            else:
                error_message = err or ((data or {}).get("message") if isinstance(data, dict) else "submit failed")

            row = {
                "run_tag": cfg.run_tag,
                "timestamp": submit_start_iso,
                "username": user.username,
                "group": user.group,
                "problem_id": problem_id,
                "language": lang,
                "submission_id": submission_id,
                "submit_start_at": submit_start_iso,
                "submit_ack_at": submit_ack_iso,
                "submit_ack_ms": f"{submit_ack_ms:.2f}",
                "first_judging_at": first_judging_iso,
                "finished_at": finished_iso,
                "queue_wait_ms": queue_wait_ms,
                "judge_exec_ms": judge_exec_ms,
                "judge_done_ms": judge_done_ms,
                "http_status": status,
                "api_code": api_code,
                "final_status": final_status,
                "final_score": final_score,
                "error_message": error_message,
            }

            async with write_lock:
                writer.writerow(row)
                output_file.flush()

            async with counters_lock:
                counters["submitted"] += 1
                if status == 200 and str(api_code) == "200":
                    counters["submit_ok"] += 1
                else:
                    counters["submit_failed"] += 1

            submitted += 1
            interval = choose_interval(elapsed_ratio, cfg, rng)
            next_submit = time.time() + interval


async def progress_reporter(counters: Dict[str, int], lock: asyncio.Lock, interval_sec: int, run_end_epoch: float) -> None:
    while time.time() < run_end_epoch:
        await asyncio.sleep(interval_sec)
        async with lock:
            msg = (
                f"progress submitted={counters['submitted']} ok={counters['submit_ok']} "
                f"submit_failed={counters['submit_failed']} "
                f"login_failed={counters['login_failed']} contest_failed={counters['contest_access_failed']}"
            )
        print(f"[{now_iso()}] {msg}")


def read_users(path: Path) -> List[UserCred]:
    users: List[UserCred] = []
    with path.open("r", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for row in reader:
            username = (row.get("username") or "").strip()
            password = (row.get("password") or "").strip()
            group = (row.get("group") or "").strip()
            if not username or not password:
                continue
            users.append(UserCred(username=username, password=password, group=group))
    return users


def parse_args() -> argparse.Namespace:
    p = argparse.ArgumentParser(description="Run realistic exam-style OJ load simulation")
    p.add_argument("--base-url", default="http://127.0.0.1:8080")
    p.add_argument("--contest-id", type=int, required=True)
    p.add_argument("--users-file", required=True, help="CSV from prepare_exam.py")
    p.add_argument("--run-tag", default="")

    p.add_argument("--duration-minutes", type=float, default=120)
    p.add_argument("--initial-spread-sec", type=float, default=180)
    p.add_argument("--max-submissions-per-user", type=int, default=0, help="0 means unlimited in run window")

    p.add_argument("--normal-until", type=float, default=0.80)
    p.add_argument("--rush-until", type=float, default=0.95)
    p.add_argument("--normal-interval-sec", default="720,1800", help="normal phase per-user interval range")
    p.add_argument("--rush-interval-sec", default="180,480", help="near-end phase per-user interval range")
    p.add_argument("--sprint-interval-sec", default="40,120", help="final sprint per-user interval range")

    p.add_argument("--language-weights", default="cpp:0.7,python:0.2,java:0.05,go:0.05")

    p.add_argument("--poll-interval-sec", type=float, default=1.5)
    p.add_argument("--poll-max-wait-sec", type=int, default=240)
    p.add_argument("--contest-ready-timeout-sec", type=int, default=1800)

    p.add_argument("--http-conn-limit", type=int, default=4)
    p.add_argument("--seed", type=int, default=42)

    p.add_argument("--output-dir", default="stress-testing/output")
    return p.parse_args()


def main() -> None:
    args = parse_args()
    if not 0 < args.normal_until < args.rush_until <= 1:
        raise SystemExit("Require 0 < normal-until < rush-until <= 1")
    if args.duration_minutes <= 0:
        raise SystemExit("duration-minutes must be > 0")

    args.normal_min_sec, args.normal_max_sec = parse_pair(args.normal_interval_sec, "--normal-interval-sec")
    args.rush_min_sec, args.rush_max_sec = parse_pair(args.rush_interval_sec, "--rush-interval-sec")
    args.sprint_min_sec, args.sprint_max_sec = parse_pair(args.sprint_interval_sec, "--sprint-interval-sec")
    args.lang_weights = parse_weights(args.language_weights)

    base_url = normalize_base_url(args.base_url)
    users = read_users(Path(args.users_file))
    if not users:
        raise SystemExit("No users found in users-file")

    run_tag = args.run_tag.strip() or datetime.now(timezone.utc).strftime("load%Y%m%d%H%M%S")
    args.run_tag = run_tag
    output_dir = Path(args.output_dir) / run_tag
    output_dir.mkdir(parents=True, exist_ok=True)

    args.run_start_epoch = time.time()
    args.run_end_epoch = args.run_start_epoch + args.duration_minutes * 60.0

    csv_path = output_dir / "submission_metrics.csv"
    summary_path = output_dir / "run_config.json"

    config_obj = {
        "run_tag": run_tag,
        "base_url": base_url,
        "contest_id": args.contest_id,
        "users_file": str(Path(args.users_file)),
        "user_count": len(users),
        "duration_minutes": args.duration_minutes,
        "phase": {
            "normal_until": args.normal_until,
            "rush_until": args.rush_until,
            "normal_interval_sec": [args.normal_min_sec, args.normal_max_sec],
            "rush_interval_sec": [args.rush_min_sec, args.rush_max_sec],
            "sprint_interval_sec": [args.sprint_min_sec, args.sprint_max_sec],
        },
        "language_weights": args.lang_weights,
        "poll_interval_sec": args.poll_interval_sec,
        "poll_max_wait_sec": args.poll_max_wait_sec,
        "seed": args.seed,
        "started_at": now_iso(),
    }
    summary_path.write_text(json.dumps(config_obj, ensure_ascii=False, indent=2), encoding="utf-8")

    headers = [
        "run_tag",
        "timestamp",
        "username",
        "group",
        "problem_id",
        "language",
        "submission_id",
        "submit_start_at",
        "submit_ack_at",
        "submit_ack_ms",
        "first_judging_at",
        "finished_at",
        "queue_wait_ms",
        "judge_exec_ms",
        "judge_done_ms",
        "http_status",
        "api_code",
        "final_status",
        "final_score",
        "error_message",
    ]

    counters = {
        "submitted": 0,
        "submit_ok": 0,
        "submit_failed": 0,
        "login_failed": 0,
        "contest_access_failed": 0,
    }

    print(f"Run tag      : {run_tag}")
    print(f"Users        : {len(users)}")
    print(f"Contest ID   : {args.contest_id}")
    print(f"Duration(min): {args.duration_minutes}")
    print(f"Output CSV   : {csv_path}")

    async def runner() -> None:
        write_lock = asyncio.Lock()
        counters_lock = asyncio.Lock()

        with csv_path.open("w", newline="", encoding="utf-8") as f:
            writer = csv.DictWriter(f, fieldnames=headers)
            writer.writeheader()

            tasks = [
                asyncio.create_task(
                    run_user(user, i, args, base_url, writer, f, write_lock, counters, counters_lock)
                )
                for i, user in enumerate(users)
            ]
            tasks.append(asyncio.create_task(progress_reporter(counters, counters_lock, 60, args.run_end_epoch)))

            await asyncio.gather(*tasks, return_exceptions=False)

    asyncio.run(runner())

    config_obj["finished_at"] = now_iso()
    config_obj["counters"] = counters
    summary_path.write_text(json.dumps(config_obj, ensure_ascii=False, indent=2), encoding="utf-8")

    print("Done.")
    print(json.dumps(counters, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
