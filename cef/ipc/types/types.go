//----------------------------------------
//
// Copyright © yanghy. All Rights Reserved.
//
// Licensed under Apache License Version 2.0, January 2004
//
// https://www.apache.org/licenses/LICENSE-2.0
//
//----------------------------------------

package types

import (
	"github.com/energye/energy/v2/cef/ipc/target"
	"github.com/energye/energy/v2/consts"
)

type IArrayValue interface {
	Size() uint32
	GetType(index uint32) consts.TCefValueType
	GetBool(index uint32) bool
	GetInt(index uint32) int32
	GetDouble(index uint32) (result float64)
	GetString(index uint32) string
	GetIValue(index uint32) IValue
	GetIBinary(index uint32) IBinaryValue
	GetIObject(index uint32) IObjectValue
	GetIArray(index uint32) IArrayValue
	Free()
}

type IBinaryValue interface {
	GetSize() uint32
	GetData(buffer []byte, dataOffset uint32) uint32
}

type IObjectValue interface {
	Size() uint32
	GetType(key string) consts.TCefValueType
	GetBool(key string) bool
	GetInt(key string) int32
	GetDouble(key string) (result float64)
	GetString(key string) string
	GetIKeys() IV8ValueKeys
	GetIValue(key string) IValue
	GetIBinary(key string) IBinaryValue
	GetIObject(key string) IObjectValue
	GetIArray(key string) IArrayValue
	Free()
}

type IValue interface {
	GetType() consts.TCefValueType
	GetBool() bool
	GetInt() int32
	GetDouble() (result float64)
	GetString() string
	GetIBinary() IBinaryValue
	GetIObject() IObjectValue
	GetIArray() IArrayValue
	Free()
}

type IV8ValueKeys interface {
	Count() int
	Get(index int) string
	Free()
}

type ICefProcessMessageIPC interface {
	Instance() uintptr
}

type IProcessMessage interface {
	EmitRender(messageId int32, eventName string, target target.ITarget, data ...any) bool
}

// OnType listening type
type OnType int8

const (
	OtMain OnType = iota // Only the main process
	OtSub                // Only the sub process
	OtAll                // All process
)

// OnOptions Listening options
type OnOptions struct {
	OnType OnType // Listening type, default main process
}
