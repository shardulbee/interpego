package code

import "testing"

func TestMake(t *testing.T) {
	tests := []struct {
		op       Opcode
		operands []int
		expected []byte
	}{
		{OpConstant, []int{65534}, []byte{byte(OpConstant), 255, 254}},
		{OpAdd, []int{}, []byte{byte(OpAdd)}},
	}
	for i, tt := range tests {
		instruction := Make(tt.op, tt.operands...)
		if len(instruction) != len(tt.expected) {
			t.Fatalf("test %d: instruction has wrong length. expected=%d, got=%d", i, len(tt.expected), len(instruction))
		}
		for idx, b := range tt.expected {
			if b != instruction[idx] {
				t.Errorf("test=%d - wrong byte at pos %d. want=%d, got=%d", i, idx, b, instruction[idx])
			}
		}
	}
}

func TestInstructionsString(t *testing.T) {
	instructions := []Instructions{
		Make(OpAdd),
		Make(OpConstant, 2),
		Make(OpConstant, 65535),
		Make(OpPop),
	}
	expected := `0000 OpAdd
0001 OpConstant 2
0004 OpConstant 65535
0005 OpPop
`
	concatted := Instructions{}
	for _, ins := range instructions {
		concatted = append(concatted, ins...)
	}
	if concatted.String() != expected {
		t.Errorf("instructions wrongly formatted.\nwant=%q\ngot=%q",
			expected, concatted.String())
	}
}
