import os
import subprocess
import sys
import tempfile
import unittest

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)

import make_assets  # noqa: E402

PNG = os.path.join(HERE, "..", "assets", "playerone.png")


class LogoTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.logo = make_assets.convert_logo(PNG)

    def test_size(self):
        self.assertEqual(len(self.logo), 280 * 205)

    def test_top_left_is_background(self):
        self.assertEqual(self.logo[0], 0x00)

    def test_enough_solid_ink(self):
        nibbles = 2 * len(self.logo)
        solid = sum((b >> 4) == 15 for b in self.logo) + sum((b & 15) == 15 for b in self.logo)
        self.assertGreaterEqual(solid, nibbles * 0.05)


class FontTest(unittest.TestCase):
    def test_size(self):
        self.assertEqual(len(make_assets.font_bytes()), 512)

    def test_space_is_blank(self):
        self.assertEqual(make_assets.font_bytes()[0:8], bytes(8))

    def test_letter_a(self):
        i = 0x41 - 0x20
        self.assertEqual(
            make_assets.font_bytes()[i * 8:i * 8 + 8],
            bytes([0x0C, 0x1E, 0x33, 0x33, 0x3F, 0x33, 0x33, 0x00]),
        )


class CliTest(unittest.TestCase):
    def test_cli_creates_outdir_and_files(self):
        with tempfile.TemporaryDirectory() as tmp:
            out = os.path.join(tmp, "nested", "out")
            subprocess.run(
                [sys.executable, os.path.join(HERE, "make_assets.py"), PNG, out],
                check=True,
            )
            self.assertEqual(os.path.getsize(os.path.join(out, "logo.raw")), 57400)
            self.assertEqual(os.path.getsize(os.path.join(out, "font.raw")), 512)


if __name__ == "__main__":
    unittest.main()
