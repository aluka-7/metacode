# 项目简介
metadata 用于储存各种元信息

code错误码一般被用来进行异常传递，且需要具有携带`message`文案信息的能力。

在这里，错误码被设计成`Codes`接口，声明如下[代码位置](https://git.forchange.cn/framework/metacode/code.go)：

```go
// 错误代码接口,其中包含代码和消息.
type Codes interface {
    // 有时错误返回字符串形式的代码
    // 注意：请勿在监控器报告中使用“error”,即使它现在也可以使用
    Error() string
    // 获取错误代码.
    Code() int
    // 获取错误信息.
    Message() string
    // 获取错误详细信息,可能为nil.
    Details() []interface{}
}

// Code是int型错误代码规范.
type Code int
```
可以看到该接口一共有四个方法，且`type Code int`结构体实现了该接口。

### 注册message

一个`Code`错误码可以对应一个`message`，默认实现会从全局变量`_messages`中获取，业务可以将自定义`Code`对应的`message`通过调用`Register`方法的方式传递进去，如：

```go
cms := map[int]string{
    0: "很好很强大！",
    -304: "啥都没变啊~",
    -404: "啥都没有啊~",
}
metacode.Register(cms)

fmt.Println(metacode.OK.Message()) // 输出：很好很强大！
```

注意：`map[int]string`类型并不是绝对，比如有业务要支持多语言的场景就可以扩展为类似`map[int]LangStruct`的结构，因为全局变量`_messages`是`atomic.Value`类型，只需要修改对应的`Message`方法实现即可。

### Details

`Details`接口为`gRrc`预留，`gRrc`传递异常会将服务端的错误码pb序列化之后赋值给`Details`，客户端拿到之后反序列化得到，具体可阅读`status`的实现：
1. `code`包内的`Status`结构体实现了`Codes`接口[代码位置](https://git.forchange.cn/framework/metacode/status.go)
2. `grpc/status`包内包装了`metacode.Status`和`grpc.Status`进行互相转换的方法[代码位置](https://git.forchange.cn/base/grpc/status/status.go)
3. `grpc`的`client`和`server`则使用转换方法将`gRrc`底层返回的`error`最终转换为`metacode.Status` [代码位置](https://git.forchange.cn/base/grpc/client.go)


## 转换为code
错误码转换有以下两种情况：
1. 因为框架传递错误是靠`code`错误码，比如mvc框架返回的`code`字段默认就是数字，那么客户端接收到如`{"code":-404}`的话，可以使用`mc := metacode.Int(-404)`或`mc := metacode.String("-404")`来进行转换。
2. 在项目中`dao`层返回一个错误码，往往返回参数类型建议为`error`而不是`metacode.Codes`，因为`error`更通用，那么上层`service`就可以使用`mc := metacode.Cause(err)`进行转换。

## 判断

错误码判断是否相等：
1. `code`与`code`判断使用：`metacode.Equal(mc1, mc2)`
2. `code`与`error`判断使用：`metacode.EqualError(mc, err)`

## 使用工具生成

使用proto协议定义错误码，格式如下：

```proto
// user.proto
syntax = "proto3";

package code;

enum UserErrCode { 
  UserUndefined = 0; // 因protobuf协议限制必须存在！！！无意义的0，工具生成代码时会忽略该参数
  UserNotLogin = 123; // 正式错误码
}
```

需要注意以下几点：

1. 必须是enum类型，且名字规范必须以"ErrCode"结尾，如：UserErrCode
2. 因为protobuf协议限制，第一个enum值必须为无意义的0


使用`feo tool protoc --code user.proto`进行生成，生成如下代码：

```go
package metacode

import (
    "git.forchange.cn/framework/metacode"
)

var _ metacode.Codes

// UserErrCode
var (
    UserNotLogin = metacode.NewCode(123);
)
```
