#!/usr/bin/env python3
"""Prepare IOI exam entities for stress testing.

This script does the following with admin credentials:
1. Batch-create users and assign groups.
2. Create one IOI contest (fixed/window).
3. Write generated credentials and metadata to output files.
"""

from __future__ import annotations

import argparse
import csv
import json
from dataclasses import dataclass
from datetime import datetime, timedelta, timezone
from pathlib import Path
from typing import Any, Dict, List

try:
    import requests
except ModuleNotFoundError as exc:  # pragma: no cover
    raise SystemExit(
        "Missing dependency: requests. Run `pip install -r stress-testing/requirements.txt` first."
    ) from exc


@dataclass
class ApiClient:
    base_url: str

    def __post_init__(self) -> None:
        url = self.base_url.rstrip("/")
        if not url.endswith("/api/v1"):
            url = f"{url}/api/v1"
        self.base_url = url
        self.session = requests.Session()

    def request(self, method: str, path: str, token: str | None = None, json_body: Any = None) -> Dict[str, Any]:
        url = f"{self.base_url}{path}"
        headers = {"Content-Type": "application/json"}
        if token:
            headers["Authorization"] = f"Bearer {token}"
        resp = self.session.request(method=method, url=url, headers=headers, json=json_body, timeout=30)

        try:
            data = resp.json()
        except ValueError:
            data = {"message": resp.text.strip() or "non-json response"}

        if resp.status_code != 200:
            msg = data.get("message") if isinstance(data, dict) else str(data)
            raise RuntimeError(f"HTTP {resp.status_code} {path}: {msg}")
        if not isinstance(data, dict) or data.get("code") != 200:
            msg = data.get("message") if isinstance(data, dict) else str(data)
            raise RuntimeError(f"API {path} failed: {msg}")

        return data


def iso_utc(dt: datetime) -> str:
    return dt.astimezone(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z")


def make_run_tag() -> str:
    return datetime.now(timezone.utc).strftime("lt%Y%m%d%H%M%S")


def chunked(items: List[Dict[str, Any]], n: int) -> List[List[Dict[str, Any]]]:
    return [items[i : i + n] for i in range(0, len(items), n)]


def list_all_users(api: ApiClient, admin_token: str, page_size: int = 100) -> List[Dict[str, Any]]:
    users: List[Dict[str, Any]] = []
    page = 1
    total = None
    while True:
        res = api.request("GET", f"/admin/users?page={page}&size={page_size}", token=admin_token)
        payload = res.get("data") or {}
        users.extend(payload.get("list") or [])
        if total is None:
            total = int(payload.get("total", 0))
        if len(users) >= total:
            break
        page += 1
    return users


def parse_args() -> argparse.Namespace:
    p = argparse.ArgumentParser(description="Prepare users/groups/contest for IOI stress test")
    p.add_argument("--base-url", default="http://127.0.0.1:8080", help="OJ base URL")
    p.add_argument("--admin-username", required=True)
    p.add_argument("--admin-password", required=True)

    p.add_argument("--user-count", type=int, default=50)
    p.add_argument("--user-prefix", default="loadstu")
    p.add_argument("--user-password", default="Pass123456")
    p.add_argument("--run-tag", default="")

    p.add_argument("--group-count", type=int, default=2)
    p.add_argument("--group-prefix", default="")

    p.add_argument("--contest-title", default="IOI Stress Test")
    p.add_argument("--contest-description", default="Realistic exam-style load testing contest")
    p.add_argument("--contest-type", default="ioi", choices=["ioi", "oi"])
    p.add_argument("--timing-mode", default="fixed", choices=["fixed", "window"])
    p.add_argument("--start-delay-minutes", type=int, default=3)
    p.add_argument("--contest-duration-minutes", type=int, default=120)
    p.add_argument("--window-duration-minutes", type=int, default=120)
    p.add_argument("--problem-ids", required=True, help="Comma-separated problem ids, e.g. 1,2,3")

    p.add_argument("--output-dir", default="stress-testing/output")
    return p.parse_args()


def main() -> None:
    args = parse_args()
    if args.user_count <= 0:
        raise SystemExit("--user-count must be > 0")
    if args.group_count <= 0:
        raise SystemExit("--group-count must be > 0")

    problem_ids = [int(x.strip()) for x in args.problem_ids.split(",") if x.strip()]
    if not problem_ids:
        raise SystemExit("--problem-ids cannot be empty")

    run_tag = args.run_tag.strip() or make_run_tag()
    group_prefix = args.group_prefix.strip() or f"{run_tag}_G"
    group_names = [f"{group_prefix}{i+1}" for i in range(args.group_count)]

    output_dir = Path(args.output_dir) / run_tag
    output_dir.mkdir(parents=True, exist_ok=True)

    api = ApiClient(args.base_url)

    print("[1/4] Admin login...")
    login = api.request(
        "POST",
        "/user/login",
        json_body={"username": args.admin_username, "password": args.admin_password},
    )
    admin_token = login["data"]["token"]

    print("[2/4] Creating users in batches...")
    users_payload: List[Dict[str, Any]] = []
    generated_users: List[Dict[str, Any]] = []
    for idx in range(1, args.user_count + 1):
        username = f"{args.user_prefix}_{run_tag}_{idx:03d}"
        group_name = group_names[(idx - 1) % len(group_names)]
        user_obj = {
            "username": username,
            "password": args.user_password,
            "email": "",
            "student_id": f"{run_tag.upper()}{idx:04d}",
            "group": group_name,
        }
        users_payload.append(user_obj)
        generated_users.append(user_obj)

    total_created = 0
    total_failed = 0
    create_errors: List[Dict[str, Any]] = []
    for batch in chunked(users_payload, 50):
        res = api.request("POST", "/admin/users/batch", token=admin_token, json_body={"users": batch})
        data = res.get("data") or {}
        total_created += int(data.get("created", 0))
        total_failed += int(data.get("failed", 0))
        create_errors.extend(data.get("errors") or [])

    all_users = list_all_users(api, admin_token)
    generated_names = {u["username"] for u in generated_users}
    created_users = [u for u in all_users if u.get("username") in generated_names]
    created_users.sort(key=lambda x: x.get("username", ""))

    if len(created_users) != args.user_count:
        print(
            f"[WARN] expected {args.user_count} users, found {len(created_users)}. "
            "If duplicate usernames existed, check errors in prepare_result.json"
        )

    print("[3/4] Creating contest...")
    now = datetime.now(timezone.utc)
    start_at = now + timedelta(minutes=args.start_delay_minutes)
    end_at = start_at + timedelta(minutes=args.contest_duration_minutes)

    payload = {
        "title": f"{args.contest_title} ({run_tag})",
        "description": args.contest_description,
        "type": args.contest_type,
        "timing_mode": args.timing_mode,
        "duration_minutes": args.window_duration_minutes if args.timing_mode == "window" else 0,
        "start_at": iso_utc(start_at),
        "end_at": iso_utc(end_at),
        "problem_ids": problem_ids,
        "allowed_users": [],
        "allowed_groups": group_names,
    }
    contest_resp = api.request("POST", "/admin/contests", token=admin_token, json_body=payload)
    contest_data = contest_resp.get("data") or {}
    contest_id = contest_data.get("id")

    print("[4/4] Writing artifacts...")
    users_csv = output_dir / "users.csv"
    with users_csv.open("w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=["user_id", "username", "password", "group"])
        writer.writeheader()
        by_name = {u.get("username"): u for u in created_users}
        for item in generated_users:
            row = by_name.get(item["username"], {})
            writer.writerow(
                {
                    "user_id": row.get("id", ""),
                    "username": item["username"],
                    "password": args.user_password,
                    "group": item["group"],
                }
            )

    contest_json = output_dir / "contest.json"
    contest_json.write_text(json.dumps(contest_data, ensure_ascii=False, indent=2), encoding="utf-8")

    result = {
        "run_tag": run_tag,
        "base_url": api.base_url,
        "contest_id": contest_id,
        "contest_start_at": iso_utc(start_at),
        "contest_end_at": iso_utc(end_at),
        "group_names": group_names,
        "problem_ids": problem_ids,
        "user_count_target": args.user_count,
        "user_count_created": total_created,
        "user_count_failed": total_failed,
        "create_errors": create_errors,
        "files": {
            "users_csv": str(users_csv),
            "contest_json": str(contest_json),
        },
    }
    result_json = output_dir / "prepare_result.json"
    result_json.write_text(json.dumps(result, ensure_ascii=False, indent=2), encoding="utf-8")

    print("Done.")
    print(f"run_tag          : {run_tag}")
    print(f"contest_id       : {contest_id}")
    print(f"users_csv        : {users_csv}")
    print(f"prepare_result   : {result_json}")


if __name__ == "__main__":
    main()
