// Copyright (c) 2020 KHS Films
//
// This file is a part of mtproto package.
// See https://github.com/lonesta/mtproto/blob/master/LICENSE for details

package tl

import "fmt"

type ErrRegisteredObjectNotFound struct {
	Crc  uint32
	Data []byte
}

func (e ErrRegisteredObjectNotFound) Error() string {
	return fmt.Sprintf("object with provided crc not registered: 0x%x", e.Crc)
}

type ErrorPartialWrite struct {
	Has  int
	Want int
}

func (e *ErrorPartialWrite) Error() string {
	return fmt.Sprintf("write failed: writed only %v bytes, expected %v", e.Has, e.Want)
}
