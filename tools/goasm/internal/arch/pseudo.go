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

	// Custom
	"syscall": true,
}

// IsPseudoInstruction checks if the given mnemonic is a pseudo-instruction
func IsPseudoInstruction(mnemonic string) bool {
	return PseudoInstructions[mnemonic]
}
