package gecore

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// Hold the *KeyPointer created with func KeyMapper()
var keyPointers = []*KeyPointer{}

// Only one ExtendedFunctionInterface can be executed
// The method Draw() of ExtendedFunctionInterface is
// Must be called at the end of main loop drawing
var currentExtendedFunction *ExtendedFunctionInterface

// KeyMapper provides an entry point for managing key mapping and command execution.
func KeyMapper() *KeyPointer {
	kp := &KeyPointer{}
	kp.root = &keyMap{}
	kp.current = kp.root
	keyPointers = append(keyPointers, kp)
	return kp
}

type KeyPointer struct {
	root             *keyMap // ルートキーマップ
	current          *keyMap
	extendedFunction *ExtendedFunctionInterface
}

type Key int64

// mod int16, Key int16, rune int32
func MakeKey(mod tcell.ModMask, key tcell.Key, ch rune) Key {
	// verb.PP("MakeKey mod %#v, ch %#v, key %#v", mod, ch, int64(mod)<<(32+16)+int64(key)<<32+int64(ch))
	return Key(int64(mod)<<(32+16) + int64(key)<<32 + int64(ch))
}

// keyMap is a map that associates keys with arbitrary actions.
type keyMap map[Key]any

// func KeyMapper() Release the KeyPointer created with
func (kp *KeyPointer) Release() {
	for index, k := range keyPointers {
		if k == kp {
			keyPointers = append(keyPointers[:index], keyPointers[index+1:]...)
			break
		}
	}
	// kp = nil
	*kp = KeyPointer{}
}

func (kp *KeyPointer) SetExtendedFunction(e *ExtendedFunctionInterface) {
	currentExtendedFunction = e
}

func (kp *KeyPointer) IsExtendedFunctionValid() bool {
	return currentExtendedFunction != nil
}

func (kp *KeyPointer) GetExtendedFunctionInterface() *ExtendedFunctionInterface {
	return currentExtendedFunction
}

// Bind adds an action associated with the specified key.
func (kp *KeyPointer) Bind(keys []string, fn any) {
	a := kp.root // Start from root *keyMap

	for i := 0; i < len(keys); i++ {
		s := keys[i]
		// md, ky, ch, err := cbind.Decode(s)
		md, ky, ch, err := Decode(s)
		if err != nil {
			// verb.PP("%#v", err)
			continue
		}
		k := MakeKey(md, ky, ch)

		if (*a) == nil {
			(*a) = make(keyMap)
		}
		if i == len(keys)-1 {
			(*a)[k] = fn
		} else {
			if (*a)[k] == nil {
				(*a)[k] = make(keyMap)
			}
			// Set pointer to next keymap
			nextKeyMap := (*a)[k].(keyMap)
			a = &nextKeyMap
		}
	}
}

// Reset resets the current keymap to the root keymap.
// Exit from ExtendedFunction
func (kp *KeyPointer) Reset() {
	kp.ResetKeyMapInAllInstance()
	currentExtendedFunction = nil
}

// ResetKeyMapInAllInstance resets the current keymap of all KeyPointer instances created with func KeyMapper() to the root keymap.
func (kp *KeyPointer) ResetKeyMapInAllInstance() {
	for _, k := range keyPointers {
		k.current = k.root
	}
}

// Reset resets the current keymap to the root keymap.
func (kp *KeyPointer) ResetKeyMap() {
	kp.current = kp.root
}

// Execute executes the action associated with the specified key.
// SkipExtendedFunction should be set to true mainly in the following cases
// - Prioritize processing over ExtendedFunctions
// - when calling from inside ExtendedFunctions
func (kp *KeyPointer) Execute(tKey *tcell.EventKey, skipExtendedFunction bool) error {
	k := MakeKey(tKey.Modifiers(), tKey.Key(), tKey.Rune())

	if !skipExtendedFunction && currentExtendedFunction != nil {
		(*currentExtendedFunction).Event(tKey)
		kp.ResetKeyMapInAllInstance()
		return ErrCodeExtendedFunction
	}

	if _, ok := (*kp.current)[k]; !ok {
		kp.ResetKeyMap()
		return ErrCodeKeyBindingNotFount
	}

	switch e := (*kp.current)[k].(type) {
	case func():
		e()
		kp.ResetKeyMapInAllInstance()
		return ErrCodeFunc
	case func() error:
		err := e()
		kp.ResetKeyMap()
		return err
	case *ExtendedFunctionInterface:
		currentExtendedFunction = e
		(*currentExtendedFunction).Event(tKey)
		kp.ResetKeyMapInAllInstance()
		return ErrCodeExtendedFunction
	case keyMap:
		kp.current = &e
		return ErrCodeKeyBound
	default:
		kp.Reset()
		return fmt.Errorf("unknown key binding type %T", e)
	}
}

type ExtendedFunctionInterface interface {
	Draw()
	Event(*tcell.EventKey) *tcell.EventKey
}
