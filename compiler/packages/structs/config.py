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
# @brief	Configuration dataclass

#----- imports
import os

from dataclasses import dataclass
from argparse import Namespace


#----- class
@dataclass(init=False)
class Config:
    """Global configuration dataclass

    Args:
        infile (str): the input filename
        outfile (str): the output filename (default: a.out)
    """
    infile: str
    outfile: str


    def __init__(self, args: Namespace) -> None:
        """Constructor

        Args:
            args (Namespace): the namespace from the Argparse parser
        """
        # set infile / outfile
        if self._is_exist(args.infile):
            self.infile = args.infile
        else:
            raise FileNotFoundError(f"Error: unable to find '{args.infile}'!")

        if args.outfile:
            self.outfile = args.outfile
        else:
            self.outfile = "a.out"


    def _is_exist(self, filename: str) -> bool:
        """Check if the file exists

        Args:
            filename (str): the filename to check

        Returns:
            bool: True if the file exists, False otherwise
        """
        return os.path.exists(filename)
