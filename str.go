package str

import (
	"unicode"
	"errors"
)

type Cursor struct {
	bytepos int // change to int64 later
	data []byte
	// consider caching len(data) here
}

func NewCursor(data []byte) *Cursor {
	return &Cursor{0, data}
}

// Return the word that starts at the cursor, or empty string
func (c *Cursor) Word() string {
	if c.bytepos >= len(c.data) {
		return ""
	}
	if unicode.IsSpace(rune(c.data[c.bytepos])) {
		return ""
	}
	if c.bytepos > 0 {
		if !unicode.IsSpace(rune(c.data[c.bytepos - 1])) {
			// Previous letter is not a blank, not a new word
			return ""
		}
	}
	endpos := -1
	for i := c.bytepos; i < len(c.data); i++ {
		if unicode.IsSpace(rune(c.data[i])) {
			endpos = i
			break
		}
	}
	if endpos == -1 {
		endpos = len(c.data)
	}
	return string(c.data[c.bytepos:endpos])
}

// Check if the cursor is at a valid byte position
func (c *Cursor) Valid() bool {
	return (c.bytepos >= 0) && (c.bytepos < len(c.data))
}

// Check if the cursor is at the given byte value
func (c *Cursor) At(b byte) bool {
	return c.Valid() && c.data[c.bytepos] == b
}

// Check if the cursor is right after the given byte value
func (c *Cursor) After(b byte) bool {
	return c.Valid() && (c.bytepos > 0) && (c.data[c.bytepos - 1] == b)
}

// Check if the cursor is at the beginning of a line
func (c *Cursor) BOL() bool {
	return (c.bytepos == 0) || c.After('\n')
}

// Check if the cursor is at the end of a line
func (c *Cursor) EOL() bool {
	return (c.bytepos == len(c.data)-1) || c.At('\n')
}

// Move to the next byte position
func (c *Cursor) Next() error {
	c.bytepos++
	if c.bytepos >= len(c.data) {
		return errors.New("END")
	}
	return nil
}

// Move to the next beginning of a word
func (c *Cursor) NextWord() error {
	if c.bytepos >= len(c.data) {
		return errors.New("END")
	}
	c.bytepos++
	// As long as the cursor is not at the start of the word,
	// progress to the next byte position until the end of the file
	for (c.Word() == "") {
		if err := c.NextWord(); err != nil {
			return err
		}
	}
	return nil
}
