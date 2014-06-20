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
		if c.GotoNextByte() != nil {
			if c.EOL() == true || c.BOL() == true {
				t.Errorf("End of line or beginning of line is true after the last byte!\n")
			}
			break
		}
	}
}

func TestKeys(t *testing.T) {
	text := []byte("a_1 = 1\nx:=2;\nZing:: 97\n  rocket92 =42\n #foo=bar\n// ignore\n/* // # ignore\n\nignore\nblublu*/\nhi=2\n\n")
	fmt.Println("\n\n", string(text))
	c := NewCursor(text)
	for c.ValidRange() {
		fmt.Println("--- start ---")

		// Check if we are at a comment
		c.RegisterCommentMarker()
		fmt.Println("IN COMMENT", c.InComment())

		// Refactor/rewrite
		if !c.InComment() {
			c.GotoNextKeyword()
			key := c.KeyString()
			fmt.Println(key)
			if !c.InComment() {
				c.GotoNextDelimiter()
				fmt.Println(c.DelimString())
				if !c.InComment() {
					c.GotoNextValue()
					fmt.Println(c.ValueString())
				}
			}
		} else {
			// Continue until we are no longer in a comment
			for c.ValidRange() && c.InComment() {
				c.GotoNextCommentMarker()
				fmt.Println("AT", c.CommentMarkerString())
				fmt.Println("level", c.GetCommentLevel())
				fmt.Println("SLC", c.AtSingleComment())
			}
		}
	}

	// Two variations over keys and values in configuration files
	//oneliner := "^kdv$"
	//multiliner := "^kd(v*)"
}
