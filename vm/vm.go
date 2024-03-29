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
	fn        *object.CompiledFunction
	ip        int
	stackBase int
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}

func NewFrame(fun *object.CompiledFunction, stackBase int) *Frame {
	return &Frame{fn: fun, ip: 0, stackBase: stackBase}
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
	vm.pushFrame(NewFrame(&object.CompiledFunction{Instructions: bytecode.Instructions}, 0))
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
	vm.pushFrame(NewFrame(&object.CompiledFunction{Instructions: bytecode.Instructions}, 0))
	return vm
}

func (vm *VM) Run() error {
	var ip int
	var instructions code.Instructions
	var op code.Opcode

	for vm.currentFrame().ip < len(vm.currentFrame().Instructions()) {
		ip = vm.currentFrame().ip
		instructions = vm.currentFrame().Instructions()
		op = code.Opcode(instructions[ip])
		switch op {
		case code.OpBang:
			popped := vm.pop()
			if popped.Type() != object.BOOLEAN_TYPE {
				return fmt.Errorf("only boolean objects are supported by bang prefix operator. got=%T (%+v)", popped, popped)
			}
			bool := popped.(*object.Boolean)
			vm.push(nativeBoolToBooleanObject(!bool.Value))
			vm.currentFrame().ip += 1
		case code.OpMinus:
			popped := vm.pop()
			if popped.Type() != object.INTEGER_TYPE {
				return fmt.Errorf("only integer objects are supported by minus prefix operator. got=%T (%+v)", popped, popped)
			}
			int := popped.(*object.Integer)
			vm.push(&object.Integer{Value: -int.Value})
			vm.currentFrame().ip += 1
		case code.OpConstant:
			constantAddress := code.ReadUint16(instructions[ip+1:])

			err := vm.push(vm.constants[constantAddress])
			if err != nil {
				return err
			}
			vm.currentFrame().ip += 3
		case code.OpAdd, code.OpMul, code.OpDiv, code.OpSub:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
			vm.currentFrame().ip += 1
		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}
			vm.currentFrame().ip += 1
		case code.OpPop:
			vm.pop()
			vm.currentFrame().ip += 1
		case code.OpTrue:
			err := vm.push(TRUE)
			if err != nil {
				return err
			}
			vm.currentFrame().ip += 1
		case code.OpFalse:
			err := vm.push(FALSE)
			if err != nil {
				return err
			}
			vm.currentFrame().ip += 1
		case code.OpNull:
			err := vm.push(NULL)
			if err != nil {
				return err
			}
			vm.currentFrame().ip += 1
		case code.OpJumpNotTruthy:
			popped := vm.pop()

			switch popped {
			case FALSE:
				jumpAddress := code.ReadUint16(instructions[ip+1:])
				vm.currentFrame().ip = int(jumpAddress)
			case TRUE:
				vm.currentFrame().ip += 3
			default:
				return fmt.Errorf("conditional expression does not have expected type. expected=ast.Boolean, got=%T (%+v)", popped, popped)
			}
		case code.OpJump:
			jumpAddress := code.ReadUint16(instructions[ip+1:])
			vm.currentFrame().ip = int(jumpAddress)
		case code.OpSetGlobal:
			vm.globals[code.ReadUint16(instructions[ip+1:])] = vm.pop()
			vm.currentFrame().ip += 3
		case code.OpGetGlobal:
			err := vm.push(vm.globals[code.ReadUint16(instructions[ip+1:])])
			if err != nil {
				return err
			}
			vm.currentFrame().ip += 3
		case code.OpSetLocal:
			localsOffset := code.ReadUint8(instructions[ip+1:])

			vm.stack[vm.currentFrame().stackBase+int(localsOffset)] = vm.pop()
			vm.currentFrame().ip += 2
		case code.OpGetLocal:
			localsOffset := instructions[ip+1]
			err := vm.push(vm.stack[vm.currentFrame().stackBase+int(localsOffset)])
			if err != nil {
				return err
			}
			vm.currentFrame().ip += 2
		case code.OpCall:
			numArgs := code.ReadUint8(instructions[ip+1:])
			vm.currentFrame().ip += 2

			// given the following fn(x, y) { let a = 1; x + y + a }(2, 3) the stack looks as follows:
			// [CompiledFunction, 2, 3, null, null, null, null, ...]
			//                           ^------ stackpointer

			// we want it to look like the following when the function starts executing:
			// [2, 3, null, null, null, null, ...]
			//  ^------ stackBase
			//                ^------- stackPointer

			var args []object.Object
			if numArgs > 0 {
				args = make([]object.Object, numArgs)
				for i := int(numArgs) - 1; i >= 0; i-- {
					popped := vm.pop()
					args[i] = popped
				}

			}

			// the stack now looks like
			// [CompiledFunction, null, null, null, null, null, null, ...]
			//                     ^------ stackpointer
			popped := vm.pop()
			fn, ok := popped.(*object.CompiledFunction)
			if !ok {
				return fmt.Errorf("calling non-function! type=%T", popped)
			}

			// the stack is now empty
			// [null, null, null, null, null, null, null, ...]
			//    ^------ stackpointer
			// this should also be the stack base, since stackBase + 0 is the first local (which is the first fn param if it exists)
			newFrame := NewFrame(fn, vm.stackPointer)

			if numArgs > 0 {
				for i := 0; i < int(numArgs); i++ {
					vm.push(args[i])
				}
			}
			// the stack now looks like
			// [2, 3, null, null, null, null, null, ...]
			//  ^------ stackBase
			//          ^------ stackPointer
			vm.stackPointer += fn.NumLocals - int(numArgs)

			// the stack now looks like
			// [2, 3, null, null, null, null, null, ...]
			//  ^------- stackBase
			//               ^------ stackpointer
			vm.pushFrame(newFrame)
		case code.OpReturnValue:
			popped := vm.pop()
			frame := vm.popFrame()
			vm.stackPointer = frame.stackBase

			err := vm.push(popped)
			if err != nil {
				return err
			}
		case code.OpReturn:
			frame := vm.popFrame()
			vm.stackPointer = frame.stackBase
		default:
			return fmt.Errorf("unknown opcode encountered: %d", op)
		}
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
