## 约定
RPC接口必需满足以下条件：

```go
type ServiceName struct {
	
}

func (*ServiceName) MethodName(request ArgType, response *ReplyType) error {
	
}
```

1. ServiceName表示的类型，必需是导出的（大写开头） 
2. 方法（MethodName）必需是导出的（大写开头） 
3. 方法必需有两个参数，且参数类型是导出的或内置的。 
4. 第一个参数是RPC请求参数，第二个参数是RPC返回值，所以第二个参数必须是指针类型，才能顺利写入 
5. 方法返回值为error类型

