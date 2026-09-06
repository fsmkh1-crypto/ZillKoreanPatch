import importlib.util
import json
import tempfile
import unittest
from pathlib import Path


MODULE_PATH = Path(__file__).with_name("apply-reviewed-copyedit.py")
SPEC = importlib.util.spec_from_file_location("apply_reviewed_copyedit", MODULE_PATH)
apply_reviewed_copyedit = importlib.util.module_from_spec(SPEC)
assert SPEC.loader is not None
SPEC.loader.exec_module(apply_reviewed_copyedit)


class ReviewedCopyeditApplicatorTests(unittest.TestCase):
    def make_repo(self, before="문장이다 다음이다<end>", after="문장이다. 다음이다<end>"):
        temp = tempfile.TemporaryDirectory()
        root = Path(temp.name)
        overlay = root / "translations/korean/messages/msgsec001.toml"
        overlay.parent.mkdir(parents=True)
        overlay.write_text(
            '# SPDX-License-Identifier: CC-BY-SA-4.0\n\n'
            '["10002"]\n'
            'japanese = "日本語<end>"\n'
            f'korean = {json.dumps(before, ensure_ascii=False)}\n',
            encoding="utf-8",
        )
        manifest = root / "manifest.json"
        manifest.write_text(
            json.dumps(
                {
                    "korean_file": "translations/korean/messages/msgsec001.toml",
                    "records": [
                        {
                            "id": 10002,
                            "japanese": "日本語<end>",
                            "before_korean": before,
                            "proposed_korean": after,
                            "decision": "PROPOSED_CHANGE",
                        }
                    ],
                },
                ensure_ascii=False,
            ),
            encoding="utf-8",
        )
        self.addCleanup(temp.cleanup)
        return root, manifest, overlay

    def test_applies_exact_reviewed_value_and_is_idempotent(self):
        root, manifest, overlay = self.make_repo()
        result = apply_reviewed_copyedit.apply_manifest(root, manifest)
        self.assertEqual(result["changed_records"], 1)
        self.assertIn('korean = "문장이다. 다음이다<end>"', overlay.read_text(encoding="utf-8"))
        second = apply_reviewed_copyedit.apply_manifest(root, manifest)
        self.assertEqual(second["changed_records"], 0)
        self.assertEqual(second["already_applied_records"], 1)
        verified = apply_reviewed_copyedit.apply_manifest(root, manifest, verify=True)
        self.assertEqual(verified["status"], "VERIFIED_APPLIED")

    def test_refuses_unreviewed_current_value(self):
        root, manifest, overlay = self.make_repo()
        overlay.write_text(
            overlay.read_text(encoding="utf-8").replace("문장이다 다음이다<end>", "다른 값<end>"),
            encoding="utf-8",
        )
        with self.assertRaisesRegex(ValueError, "refusing blind overwrite"):
            apply_reviewed_copyedit.apply_manifest(root, manifest)

    def test_refuses_changed_japanese_reference(self):
        root, manifest, overlay = self.make_repo()
        overlay.write_text(
            overlay.read_text(encoding="utf-8").replace("日本語<end>", "別の日本語<end>"),
            encoding="utf-8",
        )
        with self.assertRaisesRegex(ValueError, "Japanese/source reference changed"):
            apply_reviewed_copyedit.apply_manifest(root, manifest)

    def test_refuses_control_topology_change_in_manifest(self):
        before = "<value:$15>여 말한다<end>"
        after = "<value:$28>여 말한다<end>"
        root, manifest, _ = self.make_repo(before=before, after=after)
        with self.assertRaisesRegex(ValueError, "runtime control/substitution topology"):
            apply_reviewed_copyedit.apply_manifest(root, manifest)


if __name__ == "__main__":
    unittest.main()
