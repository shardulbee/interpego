package vm

import (
	"encoding/binary"
	"fmt"

	"interpego/code"
	"interpego/compiler"
	"interpego/object"
)

const STACK_SIZE = 2048

type VM struct {
	stack        []object.Object
	stackPointer int
	instructions code.Instructions
	constants    []object.Object
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{stack: make([]object.Object, STACK_SIZE), stackPointer: 0, instructions: bytecode.Instructions, constants: bytecode.Constants}
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])
		switch op {
		case code.OpConstant:
			constantAddress := binary.BigEndian.Uint16(vm.instructions[ip+1:])
			err := vm.push(vm.constants[constantAddress])
			if err != nil {
				return err
			}
			ip += 2
		}
	}
	return nil
}

func (vm *VM) StackTop() object.Object {
	if vm.stackPointer == 0 {
		return nil
	}
	return vm.stack[vm.stackPointer-1]
}

func (vm *VM) push(obj object.Object) error {
	if vm.stackPointer >= STACK_SIZE {
		return fmt.Errorf("stack overflow")
	}
	vm.stack[vm.stackPointer] = obj
	vm.stackPointer++
	return nil
}

func (vm *VM) pop() object.Object {
	top := vm.stack[vm.stackPointer-1]
	vm.stackPointer--
	return top
}
