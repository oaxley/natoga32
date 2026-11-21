package arch

// PseudoInstructions lists mnemonics that are pseudo-instructions
// requiring expansion before encoding
var PseudoInstructions = map[string]bool{
	// No-argument pseudo-instructions
	"nop": true,
	"ret": true,

	// Single-argument pseudo-instructions
	"call": true,
	"j":    true,
	"jr":   true,

	// Two-argument pseudo-instructions
	"li":   true,
	"la":   true,
	"mv":   true,
	"seqz": true,
	"snez": true,
	"sltz": true,
	"sgtz": true,
	"not":  true,
	"neg":  true,

	// Branch pseudo-instructions (1 register + offset)
	"beqz": true,
	"bnez": true,
	"blez": true,
	"bgez": true,
	"bltz": true,
	"bgtz": true,

	// Branch pseudo-instructions (2 registers + offset)
	"bgt":  true,
	"ble":  true,
	"bgtu": true,
	"bleu": true,

	// CSR pseudo-instructions
	"csrr":  true,
	"csrw":  true,
	"csrs":  true,
	"csrc":  true,
	"csrwi": true,
	"csrsi": true,
	"csrci": true,

	// Counters, read with csrrs rd, csr, x0
	"rdcycle":    true,
	"rdcycleh":   true,
	"rdtime":     true,
	"rdtimeh":    true,
	"rdinstret":  true,
	"rdinstreth": true,

	// Others
	"unimp": true,
	"tail":  true,

	// Custom
	"syscall": true,
}

// CounterCSRs gives the CSR read by each rd* pseudo-instruction
var CounterCSRs = map[string]int64{
	"rdcycle": 0xC00, "rdtime": 0xC01, "rdinstret": 0xC02,
	"rdcycleh": 0xC80, "rdtimeh": 0xC81, "rdinstreth": 0xC82,
}

// UnimpWord is the word of unimp: csrrw x0, cycle, x0. cycle is read-only,
// so the CPU raises an illegal-instruction trap.
const UnimpWord = 0xC0001073

// IsPseudoInstruction checks if the given mnemonic is a pseudo-instruction
func IsPseudoInstruction(mnemonic string) bool {
	return PseudoInstructions[mnemonic]
}

// PseudoOperands gives the operand names of each pseudo-instruction: the
// expected count, used in error messages. A name in brackets is optional
// and can only be the last one.
var PseudoOperands = map[string][]string{
	"nop": {}, "ret": {},
	"call": {"symbol"}, "j": {"label"}, "jr": {"rs"},
	"li": {"rd", "imm"}, "la": {"rd", "symbol"},
	"mv": {"rd", "rs"}, "not": {"rd", "rs"}, "neg": {"rd", "rs"},
	"seqz": {"rd", "rs"}, "snez": {"rd", "rs"}, "sltz": {"rd", "rs"}, "sgtz": {"rd", "rs"},
	"beqz": {"rs", "label"}, "bnez": {"rs", "label"}, "blez": {"rs", "label"},
	"bgez": {"rs", "label"}, "bltz": {"rs", "label"}, "bgtz": {"rs", "label"},
	"bgt": {"rs", "rt", "label"}, "ble": {"rs", "rt", "label"},
	"bgtu": {"rs", "rt", "label"}, "bleu": {"rs", "rt", "label"},
	"csrr": {"rd", "csr"},
	"csrw": {"csr", "rs"}, "csrs": {"csr", "rs"}, "csrc": {"csr", "rs"},
	"csrwi": {"csr", "uimm"}, "csrsi": {"csr", "uimm"}, "csrci": {"csr", "uimm"},
	"syscall": {"[N or rs]"},
	"rdcycle": {"rd"}, "rdcycleh": {"rd"}, "rdtime": {"rd"}, "rdtimeh": {"rd"},
	"rdinstret": {"rd"}, "rdinstreth": {"rd"},
	"unimp": {}, "tail": {"symbol"},
}
