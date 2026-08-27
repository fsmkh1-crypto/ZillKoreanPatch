from __future__ import annotations

import importlib.util
from pathlib import Path
from types import SimpleNamespace
import unittest
from unittest import mock


MODULE_PATH = Path(__file__).with_name("promote-staging.py")
SPEC = importlib.util.spec_from_file_location("korean_promote_staging", MODULE_PATH)
assert SPEC and SPEC.loader
promote = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(promote)


class PromoteStagingTests(unittest.TestCase):
    def test_build_rows_uses_latest_path_commit_and_dedupes_paths(self) -> None:
        path = "work/korean-results/staging-a.jsonl"
        latest = "a" * 40

        def fake_run_git(args: list[str]) -> SimpleNamespace:
            if args == ["cat-file", "-e", f"origin/korean-manual-staging:{path}"]:
                return SimpleNamespace(returncode=0, stdout="", stderr="")
            if args == ["log", "-1", "--format=%H", "origin/korean-manual-staging", "--", path]:
                return SimpleNamespace(returncode=0, stdout=latest + "\n", stderr="")
            if args == ["cat-file", "-e", f"{latest}:{path}"]:
                return SimpleNamespace(returncode=0, stdout="", stderr="")
            self.fail(f"unexpected git args: {args}")

        with mock.patch.object(promote, "run_git", side_effect=fake_run_git):
            rows = promote.build_rows("origin/korean-manual-staging", [path, path])

        self.assertEqual(
            rows,
            [{"recover_from": {"commit": latest, "path": path}}],
        )
        self.assertEqual(
            promote.render(rows),
            '{"recover_from":{"commit":"' + latest + '","path":"' + path + '"}}\n',
        )

    def test_rejects_path_outside_results_directory(self) -> None:
        for path in (
            "work/korean-staging/a.jsonl",
            "work/korean-results/../a.jsonl",
            "/work/korean-results/a.jsonl",
            "work/korean-results/a.toml",
        ):
            with self.subTest(path=path):
                with self.assertRaises(SystemExit):
                    promote.validate_staging_path(path)

    def test_fails_when_path_does_not_exist_at_staging_ref(self) -> None:
        path = "work/korean-results/missing.jsonl"
        missing = SimpleNamespace(returncode=128, stdout="", stderr="missing")
        with mock.patch.object(promote, "run_git", return_value=missing):
            with self.assertRaisesRegex(SystemExit, "staging source does not exist"):
                promote.resolve_latest_source_commit("origin/korean-manual-staging", path)

    def test_fails_when_latest_commit_does_not_contain_path(self) -> None:
        path = "work/korean-results/staging-a.jsonl"
        latest = "b" * 40
        responses = [
            SimpleNamespace(returncode=0, stdout="", stderr=""),
            SimpleNamespace(returncode=0, stdout=latest + "\n", stderr=""),
            SimpleNamespace(returncode=128, stdout="", stderr="missing from commit"),
        ]
        with mock.patch.object(promote, "run_git", side_effect=responses):
            with self.assertRaisesRegex(SystemExit, "resolved commit does not contain staging source"):
                promote.resolve_latest_source_commit("origin/korean-manual-staging", path)


if __name__ == "__main__":
    unittest.main()
