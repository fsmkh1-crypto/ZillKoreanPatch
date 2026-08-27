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

    def test_unreachable_recovery_source_fails_closed(self) -> None:
        commit = "f" * 40
        proc = SimpleNamespace(returncode=128, stdout="", stderr="fatal: invalid object name")
        with mock.patch.object(apply_results.subprocess, "run", return_value=proc):
            with self.assertRaisesRegex(SystemExit, "cannot read recovery source"):
                list(
                    apply_results.historical_rows(
                        {
                            "recover_from": {
                                "commit": commit,
                                "path": "work/korean-results/missing.jsonl",
                            }
                        },
                        Path("manifest.jsonl"),
                        1,
                    )
                )


class RecoveryConflictTests(unittest.TestCase):
    def _run_recovery_main(
        self,
        packet_texts: dict[str, str],
        manifest_rows: list[dict[str, object]],
        *,
        existing: dict[tuple[int, str], str] | None = None,
        canonical: dict[str, dict[str, object]] | None = None,
    ) -> tuple[str, str, Path]:
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        root = Path(self.tmp.name)
        manifest = root / "manifest.jsonl"
        manifest.write_text(
            "\n".join(__import__("json").dumps(row) for row in manifest_rows) + "\n",
            encoding="utf-8",
        )
        output_dir = root / "korean"
        existing = existing or {}
        canonical = canonical or {"30000": {"japanese": "A<end>"}}

        def fake_git_show(argv: list[str], **_: object) -> SimpleNamespace:
            spec = argv[2]
            try:
                content = packet_texts[spec]
            except KeyError:
                return SimpleNamespace(returncode=128, stdout="", stderr=f"missing {spec}")
            return SimpleNamespace(returncode=0, stdout=content, stderr="")

        stdout = io.StringIO()
        stderr = io.StringIO()
        with (
            mock.patch.object(apply_results, "KOREAN", output_dir),
            mock.patch.object(apply_results, "load_existing", return_value=(existing.copy(), {})),
            mock.patch.object(apply_results, "canonical_for", return_value=canonical),
            mock.patch.object(apply_results.subprocess, "run", side_effect=fake_git_show),
            mock.patch.object(apply_results.sys, "argv", ["apply-results.py", str(manifest)]),
            contextlib.redirect_stdout(stdout),
            contextlib.redirect_stderr(stderr),
        ):
            apply_results.main()
        return stdout.getvalue(), stderr.getvalue(), output_dir

    def test_identical_id_across_two_recovery_pointers_applies_once(self) -> None:
        c1, c2 = "1" * 40, "2" * 40
        p1 = "work/korean-results/a.jsonl"
        p2 = "work/korean-results/b.jsonl"
        row = '{"section":3,"id":"30000","korean":"같음<end>"}\n'
        stdout, stderr, output_dir = self._run_recovery_main(
            {f"{c1}:{p1}": row, f"{c2}:{p2}": row},
            [
                {"recover_from": {"commit": c1, "path": p1}},
                {"recover_from": {"commit": c2, "path": p2}},
            ],
        )
        self.assertIn("applied 1 new translations", stdout)
        self.assertNotIn("recovery-skip", stderr)
        rendered = (output_dir / "msgsec003-part99.toml").read_text(encoding="utf-8")
        self.assertEqual(rendered.count('["30000"]'), 1)
        self.assertIn('korean = "같음<end>"', rendered)

    def test_conflicting_recovery_pointers_preserve_first_input(self) -> None:
        c1, c2 = "3" * 40, "4" * 40
        p1 = "work/korean-results/a.jsonl"
        p2 = "work/korean-results/b.jsonl"
        stdout, stderr, output_dir = self._run_recovery_main(
            {
                f"{c1}:{p1}": '{"section":3,"id":"30000","korean":"첫값<end>"}\n',
                f"{c2}:{p2}": '{"section":3,"id":"30000","korean":"둘째값<end>"}\n',
            },
            [
                {"recover_from": {"commit": c1, "path": p1}},
                {"recover_from": {"commit": c2, "path": p2}},
            ],
        )
        self.assertIn("applied 1 new translations", stdout)
        self.assertIn("1 historical conflicts preserved current corpus", stdout)
        self.assertIn("recovery-skip 3/30000: earlier input preserved", stderr)
        rendered = (output_dir / "msgsec003-part99.toml").read_text(encoding="utf-8")
        self.assertIn('korean = "첫값<end>"', rendered)
        self.assertNotIn("둘째값", rendered)

    def test_recovery_conflict_preserves_current_corpus(self) -> None:
        commit = "5" * 40
        path = "work/korean-results/a.jsonl"
        stdout, stderr, output_dir = self._run_recovery_main(
            {f"{commit}:{path}": '{"section":3,"id":"30000","korean":"옛값<end>"}\n'},
            [{"recover_from": {"commit": commit, "path": path}}],
            existing={(3, "30000"): "현재값<end>"},
        )
        self.assertIn("applied 0 new translations", stdout)
        self.assertIn("1 historical conflicts preserved current corpus", stdout)
        self.assertIn("recovery-skip 3/30000: current corpus preserved", stderr)
        self.assertFalse(output_dir.exists(), "recovery conflict must not rewrite the current corpus")

    def test_recovery_unknown_id_fails_closed(self) -> None:
        commit = "6" * 40
        path = "work/korean-results/a.jsonl"
        with self.assertRaisesRegex(SystemExit, "unknown canonical id 3/39999"):
            self._run_recovery_main(
                {f"{commit}:{path}": '{"section":3,"id":"39999","korean":"값<end>"}\n'},
                [{"recover_from": {"commit": commit, "path": path}}],
            )

    def test_recovery_fixed_token_mismatch_fails_closed(self) -> None:
        commit = "7" * 40
        path = "work/korean-results/a.jsonl"
        with self.assertRaisesRegex(SystemExit, "fixed-control mismatch 3/30000"):
            self._run_recovery_main(
                {f"{commit}:{path}": '{"section":3,"id":"30000","korean":"값<end>"}\n'},
                [{"recover_from": {"commit": commit, "path": path}}],
                canonical={"30000": {"japanese": "A<value:$28><end><end>"}},
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
