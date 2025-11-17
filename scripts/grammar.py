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

parser = Lark(content, parser="earley", start="start", debug=True)


# open the test file
if args.test:
    with open(args.test, "r") as fh:
        for line in fh:
            line = line.rstrip("\n\r")
            if len(line) > 0:
                try:
                    print(f"--- {line}")
                    tree = parser.parse(line)
                    print(tree.pretty())
                except (Except.UnexpectedCharacters, Except.UnexpectedEOF):
                    pass
