package str

import (
	"fmt"
	"testing"
)

func TestCursor(t *testing.T) {
	text1 := []byte("x = 42\n")
	text2 := []byte("x = 32\ny=>(1,\n2,\n3)\n\nz=2")
	fmt.Println("--- text 1 ---")
	fmt.Println(string(text1))
	fmt.Println("--- text 2 ---")
	fmt.Println(string(text2))
	c := NewCursor(text1)
	fmt.Println("--- word ---")
	fmt.Println(c.Word())
	var w string
	for {
		w = c.Word()
		fmt.Println(c.bytepos, w, c.BOL(), c.EOL())
		if c.bytepos == 4 {
			if w != "42" {
				t.Errorf("The word at byte position 4 should be 42 but is: %s\n", w)
			}
		}
		if c.Next() != nil {
			if c.EOL() == true || c.BOL() == true {
				t.Errorf("End of line or beginning of line is true after the last byte!\n")
			}
			break
		}
	}
}
