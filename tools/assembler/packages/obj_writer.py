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
# @brief	Object Writer

#----- imports

from packages import constants as const

from packages.structs import Config, Symbol, SymbolType
from packages.classes import SymbolTable


#----- class
class ObjWriter:
    """Write the binary object to disk"""

    def __init__(self, config: Config) -> None:
        """Constructor"""
        self.config = config


    def write(self) -> None:
        """Write the compiled object to disk"""

        with open(self.config.outfile, "wb") as fh:
            # write the header
            fh.write(b"ON32")

            # go through all the sections
            for _, section in self.config.sections.items():
                # text section
                if section.name == '.text':
                    stype = 1
                    if self.config.arch.cpu.name == "risc-v":
                        sarch = 1
                    elif self.config.arch.cpu.name == "x68fp":
                        sarch = 2
                    else:
                        raise ValueError("Error: unknown architecture type")

                    # compute the closest power of 2 to fit the section size
                    size = (section.size - 1).bit_length()
                    byte = (stype << 7) | (sarch << 5) | size
                    fh.write(byte.to_bytes(1) + b"\x00\x00\x00")

                    # compute the entrypoint location
                    entrypoint = self.config.symbols.get(self.config.entrypoint)
                    if entrypoint.value is None:
                        raise ValueError("Error: entrypoint is not defined.")
                    else:
                        address = section.address + entrypoint.value
                        fh.write(address.to_bytes(4, const.ENDIANESS))

                    # write the data
                    fh.write(section.data)

                    # add the padding
                    padding = pow(2, size) - (section.size)
                    fh.write(b"\x00" * padding)

                if section.name in ['.data', '.bss']:
                    stype = 0
                    if section.name == '.data':
                        skind = 1
                    elif section.name == '.bss':
                        skind = 2
                    else:
                        raise ValueError("Error: unknown data section type")

                    # compute the closest power of 2 to fit the section size
                    size = (section.size - 1).bit_length()
                    byte = (stype << 7) | (skind << 5) | size
                    fh.write(byte.to_bytes(1) + b"\x00\x00\x00")

                    # write the data
                    fh.write(section.data)

                    # add the padding
                    padding = pow(2, size) - (section.size)
                    fh.write(b"\x00" * padding)

            # add the symbol table
            self.write_symbols(fh, self.config.symbols)

    def write_symbols(self, fh, symbols: SymbolTable) -> None:
        """Write the symbols in the file"""
        data = bytearray()
        for k, s in symbols.items:
            # only label are considered
            if s.type != SymbolType.LABEL:
                continue

            # encode the section and the length of the variable name
            section = 0
            if s.section == ".text":
                section = 1
            elif s.section == ".data":
                section = 2
            elif s.section == ".bss":
                section = 3

            length = len(k) & 63

            data.append((section << 6) | (length))

            # add variable name
            if len(s.name) >= 64:
                data.extend(s.name[:64].encode())
            else:
                data.extend(s.name.encode())

            # add the variable value
            data.extend(s.value.to_bytes(4, const.ENDIANESS))  # type: ignore

        # compute the overall size
        size = (len(data) -1).bit_length()

        # write the header
        header = (7 << 5) | size
        fh.write(header.to_bytes(1))
        fh.write(b"\x00\x00\x00")

        # write the data
        fh.write(data)

        # add some padding if needed
        padding = pow(2, size) - len(data)
        fh.write(b"\x00" * padding)

