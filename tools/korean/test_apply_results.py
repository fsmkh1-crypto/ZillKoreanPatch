from __future__ import annotations

import contextlib
import importlib.util
import io
from pathlib import Path
from types import SimpleNamespace
import tempfile
import unittest
from unittest import mock


MODULE_PATH = Path(__file__).with_name("apply-results.py")
SPEC = importlib.util.spec_from_file_location("korean_apply_results", MODULE_PATH)
assert SPEC and SPEC.loader
apply_results = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(apply_results)


class RecoverySpecTests(unittest.TestCase):
    def test_accepts_full_sha_and_result_packet_path(self) -> None:
        commit = "a" * 40
        got = apply_results.validate_recovery_spec(
            {"commit": commit, "path": "work/korean-results/old.jsonl"},
            Path("manifest.jsonl"),
            1,
        )
        self.assertEqual(got, (commit, "work/korean-results/old.jsonl"))

    def test_rejects_abbreviated_or_nonhex_commit(self) -> None:
        for commit in ("abc123", "g" * 40, "HEAD"):
            with self.subTest(commit=commit):
                with self.assertRaisesRegex(SystemExit, "full 40-character hex SHA"):
                    apply_results.validate_recovery_spec(
                        {"commit": commit, "path": "work/korean-results/old.jsonl"},
                        Path("manifest.jsonl"),
                        1,
                    )

    def test_rejects_path_outside_result_packet_directory(self) -> None:
        commit = "b" * 40
        for path in (
            "translations/korean/messages/msgsec001.toml",
            "work/korean-results/../secrets.jsonl",
            "/work/korean-results/old.jsonl",
            "work/korean-results/old.toml",
        ):
            with self.subTest(path=path):
                with self.assertRaises(SystemExit):
                    apply_results.validate_recovery_spec(
                        {"commit": commit, "path": path},
                        Path("manifest.jsonl"),
                        1,
                    )


class HistoricalRowsTests(unittest.TestCase):
    def test_historical_rows_include_auditable_origin(self) -> None:
        commit = "c" * 40
        packet = '{"section":3,"id":"30000","korean":"안녕<end>"}\n'
        proc = SimpleNamespace(returncode=0, stdout=packet, stderr="")
        with mock.patch.object(apply_results.subprocess, "run", return_value=proc) as run:
            rows = list(
                apply_results.historical_rows(
                    {
                        "recover_from": {
                            "commit": commit,
                            "path": "work/korean-results/old.jsonl",
                        }
                    },
                    Path("manifest.jsonl"),
                    2,
                )
            )
        self.assertEqual(
            rows,
            [
                (
                    3,
                    "30000",
                    "안녕<end>",
                    True,
                    f"{commit}:work/korean-results/old.jsonl:1",
                )
            ],
        )
        run.assert_called_once()
        self.assertEqual(
            run.call_args.args[0],
            ["git", "show", f"{commit}:work/korean-results/old.jsonl"],
        )

    def test_nested_recovery_is_rejected(self) -> None:
        commit = "d" * 40
        nested = (
            '{"recover_from":{"commit":"' + ("e" * 40) + '",'
            '"path":"work/korean-results/older.jsonl"}}\n'
        )
        proc = SimpleNamespace(returncode=0, stdout=nested, stderr="")
        with mock.patch.object(apply_results.subprocess, "run", return_value=proc):
            with self.assertRaisesRegex(SystemExit, "nested recover_from is not allowed"):
                list(
                    apply_results.historical_rows(
                        {
                            "recover_from": {
                                "commit": commit,
                                "path": "work/korean-results/old.jsonl",
                            }
                        },
                        Path("manifest.jsonl"),
                        1,
                    )
                )


class OrdinaryConflictTests(unittest.TestCase):
    def test_reports_all_conflicts_and_writes_nothing(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            packet = root / "batch.jsonl"
            packet.write_text(
                '\n'.join(
                    [
                        '{"section":3,"id":"30000","korean":"새 값 A<end>"}',
                        '{"section":3,"id":"30001","korean":"새 값 B<end>"}',
                    ]
                ) + '\n',
                encoding="utf-8",
            )
            output_dir = root / "korean"
            existing = {
                (3, "30000"): "기존 값 A<end>",
                (3, "30001"): "기존 값 B<end>",
            }
            canonical = {
                "30000": {"japanese": "A<end>"},
                "30001": {"japanese": "B<end>"},
            }
            stderr = io.StringIO()

            with (
                mock.patch.object(apply_results, "KOREAN", output_dir),
                mock.patch.object(apply_results, "load_existing", return_value=(existing, {})),
                mock.patch.object(apply_results, "canonical_for", return_value=canonical),
                mock.patch.object(apply_results.sys, "argv", ["apply-results.py", str(packet)]),
                contextlib.redirect_stderr(stderr),
            ):
                with self.assertRaisesRegex(SystemExit, "2 conflicting ordinary translation rows"):
                    apply_results.main()

            log = stderr.getvalue()
            self.assertIn("conflict 3/30000", log)
            self.assertIn("conflict 3/30001", log)
            self.assertFalse(output_dir.exists(), "fail-closed conflict path must not write overlays")


if __name__ == "__main__":
    unittest.main()
