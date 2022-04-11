
## 功能点介绍


### Engine

### Context
方法 | 功能描述
-----|-----
PostForm(key) string | 获取POST Form类型请求中的参数
Query(key) string | 获取GET请求的参数
Status(code) | 设置HTTP响应状态
SetHeader(key, value) | 设置header
String(code, format, values) | 响应string类型数据
JSON(code, obj) | 响应JSON类型数据
Data(code, data[]byte) | 响应二进制数据
HTML(code, html) | 响应html类型数据


### Router

