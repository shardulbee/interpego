package vm

import (
	"fmt"

	"interpego/code"
	"interpego/compiler"
	"interpego/object"
)

var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	NULL  = &object.Null{}
)

const (
	STACK_SIZE   = 2048
	GLOBALS_SIZE = 65536
	FRAMES_SIZE  = 1024
)

type Frame struct {
	fn *object.CompiledFunction
	ip int
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}

func NewFrame(fun *object.CompiledFunction) *Frame {
	return &Frame{fn: fun, ip: 0}
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIdx]
}

func (vm *VM) pushFrame(newFrame *Frame) {
	vm.framesIdx++
	vm.frames[vm.framesIdx] = newFrame
}

func (vm *VM) popFrame() *Frame {
	popped := vm.frames[vm.framesIdx]
	vm.framesIdx--
	return popped
}

type VM struct {
	stack        []object.Object
	stackPointer int
	frames       []*Frame
	framesIdx    int
	constants    []object.Object
	globals      []object.Object
}

func New(bytecode *compiler.Bytecode) *VM {
	vm := &VM{
		stack:        make([]object.Object, STACK_SIZE),
		stackPointer: 0,
		frames:       make([]*Frame, FRAMES_SIZE),
		framesIdx:    -1,
		constants:    bytecode.Constants,
		globals:      make([]object.Object, GLOBALS_SIZE),
	}
	vm.pushFrame(NewFrame(&object.CompiledFunction{Instructions: bytecode.Instructions}))
	return vm
}

func NewWithGlobals(globals []object.Object, bytecode *compiler.Bytecode) *VM {
	vm := &VM{
		stack:        make([]object.Object, STACK_SIZE),
		stackPointer: 0,
		frames:       make([]*Frame, FRAMES_SIZE),
		framesIdx:    -1,
		constants:    bytecode.Constants,
		globals:      globals,
	}
	vm.pushFrame(NewFrame(&object.CompiledFunction{Instructions: bytecode.Instructions}))
	return vm
}

func (vm *VM) Run() error {
	var ip int
	var op code.Opcode
	instructions := vm.currentFrame().Instructions()

	for vm.currentFrame().ip < len(instructions) {
		ip = vm.currentFrame().ip
		op = code.Opcode(instructions[ip])
		switch op {
		case code.OpBang:
			popped := vm.pop()
			if popped.Type() != object.BOOLEAN_TYPE {
				return fmt.Errorf("only boolean objects are supported by bang prefix operator. got=%T (%+v)", popped, popped)
			}
			bool := popped.(*object.Boolean)
			vm.push(nativeBoolToBooleanObject(!bool.Value))
		case code.OpMinus:
			popped := vm.pop()
			if popped.Type() != object.INTEGER_TYPE {
				return fmt.Errorf("only integer objects are supported by minus prefix operator. got=%T (%+v)", popped, popped)
			}
			int := popped.(*object.Integer)
			vm.push(&object.Integer{Value: -int.Value})
		case code.OpConstant:
			constantAddress := code.ReadUint16(instructions[ip+1:])
			err := vm.push(vm.constants[constantAddress])
			if err != nil {
				return err
			}
			vm.currentFrame().ip += 2
		case code.OpAdd, code.OpMul, code.OpDiv, code.OpSub:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}
		case code.OpPop:
			vm.pop()
		case code.OpTrue:
			err := vm.push(TRUE)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(FALSE)
			if err != nil {
				return err
			}
		case code.OpNull:
			err := vm.push(NULL)
			if err != nil {
				return err
			}
		case code.OpJumpNotTruthy:
			popped := vm.pop()

			switch popped {
			case FALSE:
				jumpAddress := code.ReadUint16(instructions[ip+1:])
				vm.currentFrame().ip = int(jumpAddress - 1)
			case TRUE:
				vm.currentFrame().ip += 2
			default:
				return fmt.Errorf("conditional expression does not have expected type. expected=ast.Boolean, got=%T (%+v)", popped, popped)
			}
		case code.OpJump:
			jumpAddress := code.ReadUint16(instructions[ip+1:])
			vm.currentFrame().ip = int(jumpAddress - 1)
		case code.OpSetGlobal:
			globalsIdx := code.ReadUint16(instructions[ip+1:])
			vm.globals[globalsIdx] = vm.pop()
			vm.currentFrame().ip += 2
		case code.OpGetGlobal:
			globalsIdx := code.ReadUint16(instructions[ip+1:])
			err := vm.push(vm.globals[globalsIdx])
			if err != nil {
				return err
			}
			vm.currentFrame().ip += 2
		case code.OpCall:
			popped := vm.pop()
			fn, ok := popped.(*object.CompiledFunction)
			if !ok {
				return fmt.Errorf("calling non-function! type=%T", popped)
			}
			newFrame := NewFrame(fn)
			vm.pushFrame(newFrame)
			vm.Run()
			vm.popFrame()
		default:
			return fmt.Errorf("unknown opcode encountered: %d", op)
		}
		vm.currentFrame().ip++
	}
	return nil
}

func (vm *VM) LastPoppedStackElement() object.Object {
	return vm.stack[vm.stackPointer]
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
	if vm.stackPointer == 0 {
		return nil
	}
	top := vm.stack[vm.stackPointer-1]
	vm.stackPointer--
	return top
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if right.Type() == object.INTEGER_TYPE && left.Type() == object.INTEGER_TYPE {
		return vm.executeIntegerBinaryOperation(op, left.(*object.Integer), right.(*object.Integer))
	}

	return fmt.Errorf("unsupported types for binary operation: %s %s", left.Type(), right.Type())
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if right.Type() == object.INTEGER_TYPE && left.Type() == object.INTEGER_TYPE {
		return vm.executeIntegerComparison(op, left.(*object.Integer), right.(*object.Integer))
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(right == left))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(right != left))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s(", op, left.Type(), right.Type())
	}
}

func (vm *VM) executeIntegerBinaryOperation(op code.Opcode, left *object.Integer, right *object.Integer) error {
	var result object.Integer
	switch op {
	case code.OpAdd:
		result = object.Integer{Value: left.Value + right.Value}
	case code.OpDiv:
		result = object.Integer{Value: left.Value / right.Value}
	case code.OpMul:
		result = object.Integer{Value: left.Value * right.Value}
	case code.OpSub:
		result = object.Integer{Value: left.Value - right.Value}
	default:
		return fmt.Errorf("unknown integer operator: %d (%T)", op, op)
	}

	vm.push(&result)
	return nil
}

func (vm *VM) executeIntegerComparison(op code.Opcode, left *object.Integer, right *object.Integer) error {
	var result *object.Boolean
	switch op {
	case code.OpEqual:
		result = nativeBoolToBooleanObject(left.Value == right.Value)
	case code.OpNotEqual:
		result = nativeBoolToBooleanObject(left.Value != right.Value)
	case code.OpGreaterThan:
		result = nativeBoolToBooleanObject(left.Value > right.Value)
	default:
		return fmt.Errorf("unknown integer comparison operator: %d (%T)", op, op)
	}

	vm.push(result)
	return nil
}

func nativeBoolToBooleanObject(val bool) *object.Boolean {
	if val {
		return TRUE
	}
	return FALSE
}
