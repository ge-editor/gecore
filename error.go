package gecore

import (
	"errors"
	"fmt"
)

// エラーコードを比較可能な定数として定義
type ErrorCode int

// 独自のエラー型
type GeError struct {
	Code    ErrorCode
	Message string
}

// Errorインターフェイスの実装
func (e *GeError) Error() string {
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

// ヘルパー関数: エラーを生成
func NewGeError(code ErrorCode, message string) *GeError {
	return &GeError{
		Code:    code,
		Message: message,
	}
}

// エラーコード判定関数 (複数のエラーを考慮)
func IsErrorCode(err error, code ErrorCode) bool {
	if err == nil {
		return false
	}

	// errors.Join の場合、Unwrap で結合されたエラーリストを探索可能
	// for {
	if unwrapped, ok := err.(interface{ Unwrap() []error }); ok {
		for _, innerErr := range unwrapped.Unwrap() {
			if IsErrorCode(innerErr, code) {
				return true
			}
		}
		return false
	}

	// 現在のエラーが GeError 型で、コードが一致するか確認
	var geErr *GeError
	if errors.As(err, &geErr) && geErr.Code == code {
		return true
	}

	// 現在のエラーが他のエラーである場合、探索終了
	//	break
	// }

	return false
}
