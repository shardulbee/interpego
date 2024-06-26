package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Opcode constants define the set of operations that can be executed by the
// virtual machine.
const (
	OpConstant Opcode = iota
	OpAdd
	OpPop
	OpSub
	OpMul
	OpDiv
	OpTrue
	OpFalse
	OpEqual
	OpNotEqual
	OpGreaterThan
	OpMinus
	OpBang
	OpJumpNotTruthy
	OpJump
	OpNull
	OpSetGlobal
	OpGetGlobal
	OpReturn
	OpReturnValue
	OpCall
	OpSetLocal
	OpGetLocal
)

type (
	Instructions []byte
	Opcode       byte
)

func (ins Instructions) String() string {
	var out bytes.Buffer

	i := 0
	for i < len(ins) {
		// at the start of the loop, we should always be at an opcode
		op := ins[i]
		def, err := Lookup(op)
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		operands, offset := ReadOperands(ins[i+1:], def)
		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))
		i += offset + 1
	}

	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n",
			len(operands), operandCount)
	}
	switch operandCount {
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	case 0:
		return fmt.Sprintf("%s", def.Name)
	default:
		return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
	}
}

func ReadOperands(ins Instructions, def *Definition) ([]int, int) {
	// we expect to call this such that the first element of ins is an operand
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for _, w := range def.OperandWidths {
		switch w {
		case 2:
			operands[offset] = int(binary.BigEndian.Uint16(ins[offset:]))
		case 1:
			operands[offset] = int(ins[offset])
		}
		offset += w
	}

	return operands, offset
}

// Definition provides the metadata for a bytecode instruction, including
// its opcode and the widths of its operands.
type Definition struct {
	Name          string
	OperandWidths []int // OperandWidths is a slice of integers representing the number of bytes each operand occupies.
}

var definitions = map[Opcode]*Definition{
	OpConstant:      {Name: "OpConstant", OperandWidths: []int{2}},
	OpMinus:         {Name: "OpMinus", OperandWidths: []int{}},
	OpBang:          {Name: "OpBang", OperandWidths: []int{}},
	OpAdd:           {Name: "OpAdd", OperandWidths: []int{}},
	OpPop:           {Name: "OpPop", OperandWidths: []int{}},
	OpSub:           {Name: "OpSub", OperandWidths: []int{}},
	OpMul:           {Name: "OpMul", OperandWidths: []int{}},
	OpDiv:           {Name: "OpDiv", OperandWidths: []int{}},
	OpTrue:          {Name: "OpTrue", OperandWidths: []int{}},
	OpFalse:         {Name: "OpFalse", OperandWidths: []int{}},
	OpNull:          {Name: "OpNull", OperandWidths: []int{}},
	OpEqual:         {Name: "OpEqual", OperandWidths: []int{}},
	OpNotEqual:      {Name: "OpNotEqual", OperandWidths: []int{}},
	OpGreaterThan:   {Name: "OpGreaterThan", OperandWidths: []int{}},
	OpJumpNotTruthy: {Name: "OpJumpNotTruthy", OperandWidths: []int{2}},
	OpJump:          {Name: "OpJump", OperandWidths: []int{2}},
	OpSetGlobal:     {Name: "OpSetGlobal", OperandWidths: []int{2}},
	OpGetGlobal:     {Name: "OpGetGlobal", OperandWidths: []int{2}},
	OpReturnValue:   {Name: "OpReturnValue", OperandWidths: []int{}},
	OpReturn:        {Name: "OpReturn", OperandWidths: []int{}},
	OpCall:          {Name: "OpCall", OperandWidths: []int{1}},
	OpSetLocal:      {Name: "OpSetLocal", OperandWidths: []int{1}},
	OpGetLocal:      {Name: "OpGetLocal", OperandWidths: []int{1}},
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return def, nil
}

// Make creates a new instruction sequence from an opcode and a set of operands.
// It is currently a stub and returns an empty byte slice.
func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		panic(fmt.Sprintf("unable create instruction for the following opcode: %d", op))
		// return []byte{}
	}

	if len(operands) != len(def.OperandWidths) {
		return []byte{}
	}

	instructionLength := 1
	for _, operandWidth := range def.OperandWidths {
		instructionLength += operandWidth
	}

	instructions := make([]byte, instructionLength)
	instructions[0] = byte(op)

	offset := 1
	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			// Converts the operand 'o' to a 2-byte sequence and stores it in 'instructions' starting at the 'offset' index.
			binary.BigEndian.PutUint16(instructions[offset:], uint16(o))
		case 1:
			instructions[offset] = byte(o)
		}
		offset += width
	}

	return instructions
}

func ReadUint16(bytes []byte) uint16 {
	return binary.BigEndian.Uint16(bytes)
}

func ReadUint8(bytes []byte) uint8 {
	return uint8(bytes[0])
}
