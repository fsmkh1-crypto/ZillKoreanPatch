#!/usr/bin/env python3
import importlib.util
import unittest
from pathlib import Path

SCRIPT = Path(__file__).with_name("materialize-beta1-reviewed-manifest.py")
SPEC = importlib.util.spec_from_file_location("materialize_beta1_reviewed_manifest", SCRIPT)
MOD = importlib.util.module_from_spec(SPEC)
assert SPEC and SPEC.loader
SPEC.loader.exec_module(MOD)


class MaterializeReviewedManifestTests(unittest.TestCase):
    def setUp(self):
        self.base = "a" * 40
        self.index = {
            "100": {
                "japanese": "甲<end>",
                "before_korean": "기존 <value:$15><end>",
                "paths": [
                    "translations/korean/messages/msgsec001-part98.toml",
                    "translations/korean/messages/msgsec001-part99.toml",
                ],
            }
        }

    def test_materializes_before_paths_and_pinned_english(self):
        manifest = MOD.build_manifest(
            {"records": [{"id": "100", "proposed_korean": "수정 <value:$15><end>"}]},
            self.index,
            self.base,
        )
        self.assertEqual(manifest["source_head"], self.base)
        self.assertEqual(manifest["english_reference_sha"], MOD.PINNED_ENGLISH_SHA)
        self.assertEqual(manifest["records"][0]["before_korean"], "기존 <value:$15><end>")
        self.assertEqual(manifest["records"][0]["paths"], self.index["100"]["paths"])

    def test_rejects_control_topology_change(self):
        with self.assertRaisesRegex(ValueError, "control/substitution topology"):
            MOD.build_manifest(
                {"records": [{"id": "100", "proposed_korean": "수정<end>"}]},
                self.index,
                self.base,
            )

    def test_rejects_duplicate_id(self):
        with self.assertRaisesRegex(ValueError, "duplicate proposal id"):
            MOD.build_manifest(
                {
                    "records": [
                        {"id": "100", "proposed_korean": "A <value:$15><end>"},
                        {"id": "100", "proposed_korean": "B <value:$15><end>"},
                    ]
                },
                self.index,
                self.base,
            )

    def test_rejects_stale_expected_before(self):
        with self.assertRaisesRegex(ValueError, "before-value changed"):
            MOD.build_manifest(
                {
                    "records": [
                        {
                            "id": "100",
                            "expected_before_korean": "옛값 <value:$15><end>",
                            "proposed_korean": "수정 <value:$15><end>",
                        }
                    ]
                },
                self.index,
                self.base,
            )


if __name__ == "__main__":
    unittest.main()
