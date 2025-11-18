#!/usr/bin/env python

import sys
import argparse

from lark import Lark, exceptions as Except

#----- begin

# command line parser
parser = argparse.ArgumentParser()
parser.add_argument("--ebnf", required=True, help="EBNF grammar")
parser.add_argument("--test", help="Test file")
args = parser.parse_args()

# ebnf parser
with open(args.ebnf, "r") as fh:
    content = fh.read()

parser = Lark(content, parser="earley", start="start", debug=False)


# open the test file
if args.test:
    lines = 0
    count = 0
    with open(args.test, "r") as fh:
        for line in fh:
            line = line.rstrip("\n\r").lstrip(" ")
            if len(line) > 0:
                lines += 1
                try:
                    tree = parser.parse(line)
                    print(f"\033[90m--- {line}\033[0m")
                    print(tree.pretty())
                    count += 1
                except (Except.UnexpectedCharacters, Except.UnexpectedEOF):
                    print(f"\033[91m--- {line}\033[0m")

    print(f"Coverage: {count} out of {lines} ({100.0 *count/lines:0.02f} %)")
