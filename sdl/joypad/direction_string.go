// Code generated by "stringer -type=Direction"; DO NOT EDIT.

package joypad

import "fmt"

const _Direction_name = "NoneUpUpRightRightDownRightDownDownLeftLeftUpLeft"

var _Direction_index = [...]uint8{0, 4, 6, 13, 18, 27, 31, 39, 43, 49}

func (i Direction) String() string {
	if i < 0 || i >= Direction(len(_Direction_index)-1) {
		return fmt.Sprintf("Direction(%d)", i)
	}
	return _Direction_name[_Direction_index[i]:_Direction_index[i+1]]
}