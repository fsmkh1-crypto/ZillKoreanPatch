import importlib.util
import json
import tempfile
import unittest
from pathlib import Path


MODULE_PATH = Path(__file__).with_name("beta1-terminology-migrate.py")
SPEC = importlib.util.spec_from_file_location("beta1_terminology_migrate", MODULE_PATH)
mod = importlib.util.module_from_spec(SPEC)
assert SPEC.loader is not None
SPEC.loader.exec_module(mod)


class Beta1TerminologyMigrationTests(unittest.TestCase):
    def make_repo(self):
        temp = tempfile.TemporaryDirectory()
        self.addCleanup(temp.cleanup)
        root = Path(temp.name)
        (root / "translations/korean/messages").mkdir(parents=True)
        (root / "translations/terminology").mkdir(parents=True)
        decisions = root / "decisions.json"
        decisions.write_text(json.dumps({"decisions":[
            {"japanese":"ロストール","korean":"로스톨","legacy":["로스토올"]},
            {"japanese":"レムオン","korean":"레무온","legacy":["레몬"]}
        ]}, ensure_ascii=False), encoding="utf-8")
        (root / "translations/terminology/korean-canonical.toml").write_text(
            'format = "zill-korean-canonical"\nversion = 1\n\n'
            '[[entry]]\njapanese = "ロストール"\nkorean = "로스토올"\n',
            encoding="utf-8"
        )
        return root, decisions

    def test_source_anchored_apply_updates_overlay_and_canonical(self):
        root, decisions = self.make_repo()
        overlay = root / "translations/korean/messages/msgsec001.toml"
        overlay.write_text(
            '["1"]\njapanese = "ロストールへ行く<end>"\nkorean = "로스토올로 간다<end>"\n\n'
            '["2"]\njapanese = "別の語<end>"\nkorean = "로스토올은 별명이다<end>"\n',
            encoding="utf-8"
        )
        result = mod.migrate(root, decisions, apply=True)
        text = overlay.read_text(encoding="utf-8")
        self.assertIn('korean = "로스톨로 간다<end>"', text)
        self.assertIn('korean = "로스토올은 별명이다<end>"', text)
        self.assertEqual(result["changed_records"], 1)
        canonical = (root / "translations/terminology/korean-canonical.toml").read_text(encoding="utf-8")
        self.assertIn('japanese = "ロストール"\nkorean = "로스톨"', canonical)
        self.assertIn('japanese = "レムオン"\nkorean = "레무온"', canonical)

    def test_dry_run_does_not_mutate(self):
        root, decisions = self.make_repo()
        overlay = root / "translations/korean/messages/msgsec001.toml"
        original='["1"]\njapanese = "レムオン<end>"\nkorean = "레몬<end>"\n'
        overlay.write_text(original, encoding="utf-8")
        result = mod.migrate(root, decisions, apply=False)
        self.assertEqual(result["changed_records"], 1)
        self.assertEqual(overlay.read_text(encoding="utf-8"), original)

    def test_refuses_unknown_korean_for_source_identity(self):
        root, decisions = self.make_repo()
        overlay = root / "translations/korean/messages/msgsec001.toml"
        overlay.write_text('["1"]\njapanese = "ロストール<end>"\nkorean = "알 수 없는 표기<end>"\n', encoding="utf-8")
        with self.assertRaisesRegex(ValueError, "neither canonical"):
            mod.migrate(root, decisions, apply=False)

    def test_refuses_unapproved_existing_canonical_value(self):
        root, decisions = self.make_repo()
        canonical = root / "translations/terminology/korean-canonical.toml"
        canonical.write_text(canonical.read_text(encoding="utf-8").replace("로스토올", "전혀다른표기"), encoding="utf-8")
        overlay = root / "translations/korean/messages/msgsec001.toml"
        overlay.write_text('["1"]\njapanese = "ロストール<end>"\nkorean = "로스토올<end>"\n', encoding="utf-8")
        with self.assertRaisesRegex(ValueError, "unexpected value"):
            mod.migrate(root, decisions, apply=False)

    def test_preserves_control_topology(self):
        root, decisions = self.make_repo()
        overlay = root / "translations/korean/messages/msgsec001.toml"
        overlay.write_text('["1"]\njapanese = "<value:$28>、ロストールへ<end>"\nkorean = "<value:$28>, 로스토올로<end>"\n', encoding="utf-8")
        mod.migrate(root, decisions, apply=True)
        self.assertIn('<value:$28>, 로스톨로<end>', overlay.read_text(encoding="utf-8"))


if __name__ == "__main__":
    unittest.main()
