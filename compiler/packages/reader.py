# -*- coding: utf-8 -*-
# vim: filetype=python
#
# This source file is subject to the MIT License
# that is bundled with this package in the file LICENSE.txt.
# It is also available through the Internet at this address:
# https://opensource.org/license/mit
#
# @author	Sebastien LEGRAND
# @license	MIT License
#
# @brief	File reader class

#----- imports
from __future__ import annotations
from typing import Any, Dict, List

import os


#----- class
class FileReader:
    """Read a file byte by byte"""
    def __init__(self, filename: str) -> None:
        """Constructor"""
        self._fh = open(filename, "r")
        self._count = 0

        # compute the size of the file
        self._fh.seek(0, os.SEEK_END)
        self._size = self._fh.tell()
        self._fh.seek(0, os.SEEK_SET)

    @property
    def size(self) -> int:
        """Returns the size of the file"""
        return self._size

    @property
    def count(self) -> int:
        """Returns the current number of bytes read"""
        return self._count

    def next(self) -> str:
        """Returns the next character in the file"""
        c = self._fh.read(1)
        if not c:
            self._count = self._count + 1
        return c

    def peek(self) -> str:
        """Returns the next character, but don't move the pointer"""
        c = self._fh.read(1)
        self._fh.seek(-1, 1)
        return c

    def rewind(self, count: int = 1) -> None:
        """Rewind the pointer from the current position"""
        self._fh.seek(-count, 1)
