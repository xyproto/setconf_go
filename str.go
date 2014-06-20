package str

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

type Cursor struct {
	bytepos       int // change to int64 later
	data          []byte
	commentlevel  int  // keep track of nested comments
	singlecomment bool // keep track of comments on single lines
}

func NewCursor(data []byte) *Cursor {
	return &Cursor{0, data, 0, false}
}

// Not a letter that typically is part of a key, delimiter or value in a configuration file
func ConfigLetter(b byte) bool {
	// Not blanks or weird characters. Could check for a-zA-Z0-9_ instead, but that's more restrictive.
	return !(unicode.IsSpace(rune(b)) || strings.Contains("`'()[]{}\"", string(b)))
}

// Letters that are likely to be part of a comment marker
func CommentLetter(b byte) bool {
	return !unicode.IsSpace(rune(b)) && strings.Contains("/*#\n", string(b))
}

// Letters that are likely to be part of a key
func KeyLetter(b byte) bool {
	return (ConfigLetter(b) && !DelimLetter(b)) && !EOL_marker(b)
}

// Letters that are likely to be part of a delimiter. A key value delimiter may be just a blank or \t.
func DelimLetter(b byte) bool {
	return strings.Contains("=<>: \t", string(b)) && !EOL_marker(b)
}

// Letters that are likely to be part of a value that only spans one line
func ValueLetter(b byte) bool {
	return (!unicode.IsSpace(rune(b)) && !DelimLetter(b)) && !strings.Contains(";\n", string(b))
}

// Generate a check function that just checks that the byte is not the end byte
func GenerateMultilineValueLetterCheck(endmark byte) func(byte) bool {
	return func(b byte) bool {
		return b != endmark
	}
}

// End of line marker
func EOL_marker(b byte) bool {
	return b == '\n'
}

// Return the word that starts at the cursor, or an empty string
func (c *Cursor) GetWord(validLetter func(b byte) bool) string {
	if c.bytepos >= len(c.data) {
		return ""
	}
	if !validLetter(c.data[c.bytepos]) {
		return ""
	}
	if c.bytepos > 0 {
		if validLetter(c.data[c.bytepos-1]) {
			// Previous letter is not a blank, not a new word
			return ""
		}
	}
	endpos := -1
	for i := c.bytepos; i < len(c.data); i++ {
		if !validLetter(c.data[i]) {
			endpos = i
			break
		}
	}
	if endpos == -1 {
		endpos = len(c.data)
	}
	return string(c.data[c.bytepos:endpos])
}

// Return the config string (not space, not weird characters) that starts at the cursor, or an empty string
func (c *Cursor) Word() string {
	return c.GetWord(ConfigLetter)
}

// Return the key that starts at the cursor, or an empty string
func (c *Cursor) KeyString() string {
	return c.GetWord(KeyLetter)
}

// Return the delimiter that starts at the cursor, or an empty string.
// Trims the string if it's more than just one space/letter (or an empty string)
func (c *Cursor) DelimString() string {
	delim := c.GetWord(DelimLetter)
	if len(delim) < 2 {
		return delim
	}
	return strings.TrimSpace(delim)
}

// Return the value that starts at the cursor, or an empty string
func (c *Cursor) ValueString() string {
	return c.GetWord(ValueLetter)
}

// Return the multiline value that starts at the cursor and ends at the endmark, or an empty string
func (c *Cursor) MultiValueString(endmark byte) string {
	return c.GetWord(GenerateMultilineValueLetterCheck(endmark))
}

// Return the EOL marker that starts at the cursor, or an empty string
func (c *Cursor) EOLString() string {
	return c.GetWord(EOL_marker)
}

// Return how deeply we are nested into comments
func (c *Cursor) GetCommentLevel() int {
	return c.commentlevel
}

// Check if we are in a single line comment
func (c *Cursor) AtSingleComment() bool {
	return c.singlecomment
}

// Check if we have already registered that we are in a comment
func (c *Cursor) InComment() bool {
	return c.singlecomment || (c.commentlevel > 0)
}

// Check if we are in a comment. Changes the comment state for the cursor if change is true.
func (c *Cursor) RegisterCommentMarker() bool {
	switch c.CommentMarkerString() {
	case "/*":
		if !c.singlecomment {
			fmt.Println("Found start of multiline comment:", c.CommentMarkerString())
			c.commentlevel++
		}
	case "*/":
		if !c.singlecomment {
			c.commentlevel--
		}
	case "#", "//":
		if c.commentlevel == 0 {
			c.singlecomment = true
		}
	case "\n":
		if c.commentlevel == 0 {
			if c.singlecomment {
				c.singlecomment = false
			}
		}
	}
	return c.singlecomment || (c.commentlevel > 0)
}

func (c *Cursor) CommentMarkerString() string {
	return c.GetWord(CommentLetter)
}

// Check if the cursor is at a valid byte position
func (c *Cursor) ValidRange() bool {
	return (c.bytepos >= 0) && (c.bytepos < len(c.data))
}

// Check if the cursor is at the given byte value
func (c *Cursor) At(b byte) bool {
	return c.ValidRange() && c.data[c.bytepos] == b
}

// Get the string for the byte at the current position, or a blank string
func (c *Cursor) Get() string {
	if c.ValidRange() {
		return string(c.data[c.bytepos])
	}
	return ""
}

// Check if the cursor is right after the given byte value
func (c *Cursor) After(b byte) bool {
	return c.ValidRange() && (c.bytepos > 0) && (c.data[c.bytepos-1] == b)
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
func (c *Cursor) GotoNextByte() error {
	c.bytepos++
	if c.bytepos >= len(c.data) {
		return errors.New("END")
	}
	// Change the state for comments if we are at a comment
	c.RegisterCommentMarker()
	return nil
}

// Move to the next beginning of a word of the given type/validation function. Also keep track of comments.
func (c *Cursor) GotoNextWord(wordType func() string) error {
	if c.bytepos >= len(c.data) {
		return errors.New("END")
	}
	c.bytepos++
	// As long as the cursor is not at the start of the word,
	// progress to the next byte position until the end of the file
	for wordType() == "" {
		if err := c.GotoNextWord(wordType); err != nil {
			return err
		}
	}
	// Change the state for comments if we are at a comment
	return nil
}

// Move to the next keyword
func (c *Cursor) GotoNextKeyword() error {
	err := c.GotoNextWord(c.KeyString)
	c.RegisterCommentMarker()
	return err
}

// Move to the next delimiter
func (c *Cursor) GotoNextDelimiter() error {
	err := c.GotoNextWord(c.DelimString)
	c.RegisterCommentMarker()
	return err
}

// Move to the next value
func (c *Cursor) GotoNextValue() error {
	err := c.GotoNextWord(c.ValueString)
	c.RegisterCommentMarker()
	return err
}

// Move to the next comment marker
func (c *Cursor) GotoNextCommentMarker() error {
	err := c.GotoNextWord(c.CommentMarkerString)
	c.RegisterCommentMarker()
	return err
}

func (c *Cursor) GotoEOL() error {
	err := c.GotoNextWord(c.EOLString)
	c.RegisterCommentMarker()
	return err
}
