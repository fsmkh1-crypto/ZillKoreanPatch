#!/usr/bin/env python3
"""Capture the CPU state at a reproducible PPSSPP freeze.

This is intentionally narrow: it pauses the currently running/looping CPU through
Zill's existing PPSSPP debugger bridge, captures all registers, disassembles the
current PC, and dumps a small stack window.  The emulator is intentionally left
stopped so the captured state cannot move before follow-up inspection.
"""

from __future__ import annotations

import argparse
import json
import pathlib
import subprocess
import sys
from typing import Any


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--port", required=True, type=int, help="PPSSPP debugger port")
    parser.add_argument("--host", default="127.0.0.1")
    parser.add_argument("--zill", default="./zill", help="path to zill executable")
    parser.add_argument("--out", default="freeze-snapshot.json")
    parser.add_argument("--disasm-before", type=int, default=16)
    parser.add_argument("--disasm-count", type=int, default=40)
    parser.add_argument("--stack-bytes", type=int, default=512)
    args = parser.parse_args()
    if not 1 <= args.port <= 65535:
        parser.error("--port must be between 1 and 65535")
    if args.disasm_before < 0 or args.disasm_count < 1 or args.stack_bytes < 1:
        parser.error("snapshot sizes must be positive")
    return args


def read_json_line(proc: subprocess.Popen[str]) -> dict[str, Any]:
    assert proc.stdout is not None
    line = proc.stdout.readline()
    if not line:
        stderr = proc.stderr.read() if proc.stderr is not None else ""
        raise RuntimeError(f"PPSSPP debugger bridge closed unexpectedly: {stderr.strip()}")
    return json.loads(line)


def command(proc: subprocess.Popen[str], ident: int, name: str, **fields: Any) -> dict[str, Any]:
    assert proc.stdin is not None
    payload = {"id": ident, "command": name, **fields}
    proc.stdin.write(json.dumps(payload, separators=(",", ":")) + "\n")
    proc.stdin.flush()
    while True:
        row = read_json_line(proc)
        if row.get("id") != ident:
            continue
        if not row.get("ok"):
            error = row.get("error", {})
            raise RuntimeError(error.get("message", f"command {name} failed"))
        return row["result"]


def raw(proc: subprocess.Popen[str], ident: int, event: str, params: dict[str, Any] | None = None) -> dict[str, Any]:
    result = command(proc, ident, "raw", event=event, params=params or {})
    return result["response"]


def register_value(registers: dict[str, Any], name: str) -> int | None:
    for category in registers.get("categories", []):
        names = category.get("registerNames", [])
        values = category.get("uintValues", [])
        for index, register_name in enumerate(names):
            if register_name == name and index < len(values):
                return int(values[index]) & 0xFFFFFFFF
    return None


def main() -> int:
    args = parse_args()
    cmd = [args.zill, "ppsspp-debugger", "--host", args.host, "--port", str(args.port), "--timeout", "10"]
    proc = subprocess.Popen(cmd, stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, bufsize=1)
    try:
        ready = read_json_line(proc)
        if ready.get("event") not in {"ready", "reconnected"}:
            raise RuntimeError(f"unexpected debugger handshake: {ready}")

        before = command(proc, 1, "status")
        cpu_before = before.get("cpu", {})
        if not cpu_before.get("stepping") and not cpu_before.get("paused"):
            pause = command(proc, 2, "pause")
        else:
            pause = {"changed": False, "cpu": cpu_before}

        cpu = raw(proc, 3, "cpu.status")
        registers = raw(proc, 4, "cpu.getAllRegs")
        pc = register_value(registers, "pc")
        sp = register_value(registers, "sp")
        if pc is None:
            pc = int(cpu.get("pc", 0)) & 0xFFFFFFFF
        if pc == 0:
            raise RuntimeError("captured PC is zero; game CPU is not available")

        start = max(0, pc - args.disasm_before * 4)
        disasm = raw(proc, 5, "memory.disasm", {
            "address": start,
            "count": args.disasm_count,
            "displaySymbols": True,
        })

        stack: dict[str, Any] | None = None
        stack_error: str | None = None
        if sp is not None and sp != 0:
            try:
                stack = command(proc, 6, "read_memory", address=sp, size=args.stack_bytes, replacements=True)
            except RuntimeError as exc:
                stack_error = str(exc)

        snapshot = {
            "format": "zill-ppsspp-freeze-snapshot-v1",
            "purpose": "capture actual CPU state after the reproducible Korean-patch freeze",
            "bridge": ready,
            "status_before_capture": before,
            "pause": pause,
            "cpu": cpu,
            "pc": pc,
            "pc_hex": f"0x{pc:08X}",
            "sp": sp,
            "sp_hex": None if sp is None else f"0x{sp:08X}",
            "registers": registers,
            "disassembly": disasm,
            "stack": stack,
            "stack_error": stack_error,
            "left_paused": True,
        }
        path = pathlib.Path(args.out).resolve()
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(json.dumps(snapshot, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
        print(f"FREEZE_SNAPSHOT pc=0x{pc:08X} sp={snapshot['sp_hex']} out={path}")
        print("PPSSPP is intentionally left paused; do not resume it before preserving the snapshot.")
        command(proc, 7, "quit")
        return 0
    except Exception as exc:
        print(f"freeze snapshot failed: {exc}", file=sys.stderr)
        return 1
    finally:
        if proc.stdin is not None:
            proc.stdin.close()
        try:
            proc.wait(timeout=2)
        except subprocess.TimeoutExpired:
            proc.kill()


if __name__ == "__main__":
    raise SystemExit(main())
