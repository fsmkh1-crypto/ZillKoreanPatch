#!/usr/bin/env python3
import argparse
import difflib
import glob
import json
import re
import tomllib
from pathlib import Path

LINE_BREAK = "<line-break>"
SPACE_RE = re.compile(r"\s+")
SECTION_RE = re.compile(r'^["?')
