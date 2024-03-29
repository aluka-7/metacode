package metacode

import (
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"sync"
)

// All common code
var (
	OK      = add(0) // 正确
	Success = add(1) // 成功

	NotModified        = add(-304) // 木有改动
	TemporaryRedirect  = add(-307) // 撞车跳转
	RequestErr         = add(-400) // 请求错误
	Unauthorized       = add(-401) // 未认证
	AccessDenied       = add(-403) // 访问权限不足
	NothingFound       = add(-404) // 啥都木有
	MethodNotAllowed   = add(-405) // 不支持该方法
	Conflict           = add(-409) // 冲突
	Canceled           = add(-498) // 客户端取消请求
	ServerErr          = add(-500) // 服务器错误
	ServiceUnavailable = add(-503) // 过载保护,服务暂不可用
	Deadline           = add(-504) // 服务调用超时
	LimitExceed        = add(-509) // 超出限制
	ValidateErr        = add(-512) // 服务器请求参数校验出错
)
var (
	_messages sync.Map             // NOTE: stored map[int]string
	_codes    = map[int]struct{}{} // register codes.
)

// Register register code message map.
func Register(l string, cm map[int]string) {
	_messages.Store(l, cm)
}

// NewCode 新建一个新的元数据。
// 注意：代码必须在全局范围内唯一，新代码将检查重复，然后出现恐慌。
func NewCode(c int) Code {
	if c <= 0 {
		panic("business code must greater than zero")
	}
	return add(c)
}
func add(e int) Code {
	if _, ok := _codes[e]; ok {
		panic(fmt.Sprintf("metacode code: %d already exist", e))
	}
	_codes[e] = struct{}{}
	return IntCode(e)
}

// 错误代码接口,其中包含代码和消息.
type Codes interface {
	// Error 有时错误返回字符串形式的代码
	// 注意：请勿在监控器报告中使用“error”,即使它现在也可以使用
	Error() string
	// Code 获取错误代码.
	Code() int
	// Message 获取错误信息.
	// param l string 语言
	Message(l string) string
	// Details 获取错误详细信息,可能为nil.
	Details() []interface{}
}

// Code是int型错误代码规范.
type Code int

func (e Code) Error() string {
	return strconv.FormatInt(int64(e), 10)
}

// Code return error code
func (e Code) Code() int { return int(e) }

// Message return error message
func (e Code) Message(l string) string {
	if m, ok := _messages.Load(l); ok {
		if cm, ok := m.(map[int]string); ok {
			if msg, ok := cm[e.Code()]; ok {
				return msg
			}
		}
	}
	return e.Error()
}

// Details return details.
func (e Code) Details() []interface{} { return nil }

// Int parse code int to error.
func IntCode(i int) Code { return Code(i) }

// String parse code string to error.
func String(e string) Code {
	if e == "" {
		return OK
	}
	// try error string
	i, err := strconv.Atoi(e)
	if err != nil {
		return ServerErr
	}
	return Code(i)
}

// Cause cause from error to code.
func Cause(e error) Codes {
	if e == nil {
		return OK
	}
	ec, ok := errors.Cause(e).(Codes)
	if ok {
		return ec
	}
	return String(e.Error())
}

// Equal equal a and b by code int.
func Equal(a, b Codes) bool {
	if a == nil {
		a = OK
	}
	if b == nil {
		b = OK
	}
	return a.Code() == b.Code()
}

// EqualError equal error
func EqualError(code Codes, err error) bool {
	return Cause(err).Code() == code.Code()
}
