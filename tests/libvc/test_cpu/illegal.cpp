#include <catch2/catch_all.hpp>

#include "instruction_test.h"
#include "core/csr.h"
#include "vc/exceptions.h"

using namespace vc;

// trap handler address used by the tests
constexpr u32 HANDLER = 0x0000'0200;

// opcodes
constexpr u32 OPC_BRANCH = 0x63;
constexpr u32 OPC_JALR   = 0x67;
constexpr u32 OPC_OP_IMM = 0x13;
constexpr u32 OPC_OP     = 0x33;
constexpr u32 OPC_CUSTOM = 0x0B;
constexpr u32 OPC_SYSTEM = 0x73;

// R-type layout: funct7 | rs2 | rs1 | funct3 | rd | opcode
static constexpr u32 encode(u32 funct7, u32 rs2, u32 rs1, u32 funct3, u32 rd, u32 opcode)
{
    return (funct7 << 25) | (rs2 << 20) | (rs1 << 15) | (funct3 << 12) | (rd << 7) | opcode;
}

static u32 readCsr(InstructionTest& test, u16 id)
{
    return test.mem.read32(MMAP_CSR_REGISTERS + id * sizeof(u32));
}

struct IllegalCase
{
    const char* name;
    u32 instruction;
};

TEST_CASE("Unassigned encodings raise an illegal instruction trap", "[cpu][trap][illegal]") {
    auto c = GENERATE(
        // branches: funct3 2 and 3 are not assigned
        IllegalCase{"branch funct3 2", encode(0, 7, 6, 0b010, 0, OPC_BRANCH)},
        IllegalCase{"branch funct3 3", encode(0, 7, 6, 0b011, 0, OPC_BRANCH)},

        // jalr: only funct3 0
        IllegalCase{"jalr funct3 1", encode(0, 0, 6, 0b001, 5, OPC_JALR)},

        // OP-IMM funct3 001: SLLI / BSETI / BCLRI / BINVI / CLZ group
        IllegalCase{"op-imm 001 funct7 0x01 (shamt bit 5)", encode(0x01, 3, 6, 0b001, 5, OPC_OP_IMM)},
        IllegalCase{"op-imm 001 funct7 0x10", encode(0x10, 3, 6, 0b001, 5, OPC_OP_IMM)},
        IllegalCase{"op-imm 001 funct7 0x30 rs2 3", encode(0x30, 3, 6, 0b001, 5, OPC_OP_IMM)},
        IllegalCase{"op-imm 001 funct7 0x30 rs2 6", encode(0x30, 6, 6, 0b001, 5, OPC_OP_IMM)},
        IllegalCase{"op-imm 001 funct7 0x30 rs2 31", encode(0x30, 31, 6, 0b001, 5, OPC_OP_IMM)},

        // OP-IMM funct3 101: SRLI / SRAI / BEXTI / ORC.B / RORI / REV8
        IllegalCase{"op-imm 101 funct7 0x01 (shamt bit 5)", encode(0x01, 3, 6, 0b101, 5, OPC_OP_IMM)},
        IllegalCase{"op-imm 101 funct7 0x10", encode(0x10, 3, 6, 0b101, 5, OPC_OP_IMM)},
        IllegalCase{"op-imm 101 funct7 0x14 rs2 0 (next to orc.b)", encode(0x14, 0, 6, 0b101, 5, OPC_OP_IMM)},
        IllegalCase{"op-imm 101 funct7 0x14 rs2 6 (next to orc.b)", encode(0x14, 6, 6, 0b101, 5, OPC_OP_IMM)},
        IllegalCase{"op-imm 101 funct7 0x34 rs2 0 (next to rev8)", encode(0x34, 0, 6, 0b101, 5, OPC_OP_IMM)},
        IllegalCase{"op-imm 101 funct7 0x34 rs2 7 (brev8, Zbkb)", encode(0x34, 7, 6, 0b101, 5, OPC_OP_IMM)},

        // OP: one unassigned funct7 per funct3
        IllegalCase{"op 000 funct7 0x04", encode(0x04, 7, 6, 0b000, 5, OPC_OP)},
        IllegalCase{"op 001 funct7 0x20", encode(0x20, 7, 6, 0b001, 5, OPC_OP)},
        IllegalCase{"op 010 funct7 0x20", encode(0x20, 7, 6, 0b010, 5, OPC_OP)},
        IllegalCase{"op 011 funct7 0x10", encode(0x10, 7, 6, 0b011, 5, OPC_OP)},
        IllegalCase{"op 100 funct7 0x07", encode(0x07, 7, 6, 0b100, 5, OPC_OP)},
        IllegalCase{"op 101 funct7 0x04", encode(0x04, 7, 6, 0b101, 5, OPC_OP)},
        IllegalCase{"op 110 funct7 0x04", encode(0x04, 7, 6, 0b110, 5, OPC_OP)},
        IllegalCase{"op 111 funct7 0x30", encode(0x30, 7, 6, 0b111, 5, OPC_OP)},
        IllegalCase{"op 000 funct7 0x7F", encode(0x7F, 7, 6, 0b000, 5, OPC_OP)},

        // custom thread instructions: funct3 3 and 6 are not assigned
        IllegalCase{"custom funct3 3", encode(0, 0, 0, 0b011, 5, OPC_CUSTOM)},
        IllegalCase{"custom funct3 6", encode(0, 0, 0, 0b110, 5, OPC_CUSTOM)},

        // SYSTEM: funct3 4 and unknown funct12 values
        IllegalCase{"system funct3 4", encode(0x18, 0, 6, 0b100, 5, OPC_SYSTEM)},
        IllegalCase{"system funct12 0x7FF", encode(0x3F, 31, 0, 0b000, 0, OPC_SYSTEM)},
        IllegalCase{"system funct12 0x102 (sret)", encode(0x08, 2, 0, 0b000, 0, OPC_SYSTEM)},
        IllegalCase{"ecall with rd != 0", encode(0, 0, 0, 0b000, 5, OPC_SYSTEM)},
        IllegalCase{"ecall with rs1 != 0", encode(0, 0, 6, 0b000, 0, OPC_SYSTEM)},
        IllegalCase{"mret with rd != 0", encode(0x18, 2, 0, 0b000, 5, OPC_SYSTEM)}
    );

    InstructionTest test;
    test.mem.write32(MMAP_CSR_REGISTERS + CSR_MTVEC * sizeof(u32), HANDLER);

    INFO(c.name << " = 0x" << std::hex << c.instruction);

    test.setReg(5, 0xAAAA'5555);    // rd sentinel
    test.setReg(6, 0x1234'5678);
    test.setReg(7, 0x0000'0003);
    test.loadInstruction(c.instruction);
    u32 instr_addr = test.getPC();
    test.execute();

    REQUIRE(readCsr(test, CSR_MCAUSE) == CPU_MAIN_ILLEGAL_INSTRUCTION);
    REQUIRE(readCsr(test, CSR_MEPC) == instr_addr);
    REQUIRE(test.getPC() == HANDLER);
    REQUIRE(test.getReg(5) == 0xAAAA'5555);
}

TEST_CASE("ebreak and wfi still move to the next instruction", "[cpu][illegal]") {
    // what these two should do is not decided yet: they stay no-ops for now
    auto instruction = GENERATE(0x0010'0073u, 0x1050'0073u);

    InstructionTest test;
    test.mem.write32(MMAP_CSR_REGISTERS + CSR_MTVEC * sizeof(u32), HANDLER);
    test.loadInstruction(instruction);
    test.execute();

    REQUIRE(test.deltaPC() == 4);
    REQUIRE(readCsr(test, CSR_MCAUSE) == 0);
}
