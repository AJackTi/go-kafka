#!/usr/bin/env python3
"""Validate the syntax of an example dotenv file without loading secrets."""

from __future__ import annotations

import re
import sys
from pathlib import Path


NAME = re.compile(r"[A-Z][A-Z0-9_]*\Z")


def validate(path: Path) -> None:
    names: set[str] = set()
    for line_number, raw_line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
        line = raw_line.strip()
        if not line or line.startswith("#"):
            continue
        if "=" not in line:
            raise ValueError(f"{path}:{line_number}: missing '='")
        name, _ = line.split("=", 1)
        if not NAME.fullmatch(name):
            raise ValueError(f"{path}:{line_number}: invalid variable name {name!r}")
        if name in names:
            raise ValueError(f"{path}:{line_number}: duplicate variable {name!r}")
        names.add(name)


def main() -> int:
    paths = [Path(value) for value in sys.argv[1:]]
    if not paths:
        print("usage: validate_env.py FILE [FILE ...]", file=sys.stderr)
        return 2
    try:
        for path in paths:
            validate(path)
    except (OSError, ValueError) as error:
        print(error, file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
