# 说明

入职虾皮做的EntryTask，写一个简单的账号管理系统，要求写一个http server及前端页面进行交互，http server后面调tcp server(RPC)，tcp server中完成业务逻辑。建议自己实现RPC。要求优先完成功能及文档，评分占比大概是功能完成-3，压测-2，文档-1（具体分值不记得了）。

一周时间完成entry-task，时间不够用，前端页面没有做完，文档写得不够完善。

用户登录参考以前公司水滴筹的方案，简单来说就是用户登录后生成一个token，后续前端所有接口请求都带上此token做用户认证。token有过期时间（一般为1天），为了提升用户体验避免用户反复登录，后端还下发了一个refreshToken用来在token过期之后刷新token。refreshToken只在登录成功后和token下发一次，且其他接口请求都不需要它，它的作用就是在token过期后刷新token用的。refreshToken自身也会有过期时间（一般比token过期时间长一点，比如7天），如果refreshToken也过期了，那用户必需再次登录了。

## 工程目录结构说明

| 目录          | 说明                         |
|-------------|----------------------------|
| ant         | 自己实现的web框架，参考gin                 |
| api         | 接口参数相关                     |
| bee         | 自己实现的rpc框架，参考Go自带RPC框架                 |
| config      | 配置相关                       |
| doc         | 文档相关                       |
| http-server | http-server，仅给tcp-server转发 |
| tcp-server  | tcp server                 |
| util        | 工具类                        |

## tcp-server目录结构说明

tcp-server是主逻辑处理的地方，参考DDD实现。

| 目录         | 说明                              |
|------------|---------------------------------|
| app        | 负责响应request，使用domain等下层组件完成逻辑处理 |
| domain     | 领域层，包含系统核心域和逻辑                  |
| infra      | 基础组件（db, redis等）                |
| repository | 负责存储（redis/mysql）相关的逻辑          |

## 服务部署
1. 设置profile环境变量：export ZOO_PROFILE=test
2. 添加配置文件（同事推荐的配置库viper只支持读单一文件）`/opt/data/config/config.test.yml`，文件中间部分与步骤1设置的profile一致
3. 编译http-server：`cd http-server && go build`
4. 执行http-server：`./http-server`
5. 编译tcp-server：`cd tcp-server && go build`
6. 执行tcp-server：`./tcp-server`

## 接口文档

### 1. 用户注册

```bash
curl --location --request POST 'http://zoo.com/api/user/register' \
--header 'Content-Type: application/json' \
--data-raw '{
    "username": "test1",
    "password": "XXXXXXXXXXX"
}'
```

### 2. 用户登录

```bash
curl --location --request POST 'http://zoo.com/api/user/login' \
--header 'Content-Type: application/json' \
--data-raw '{
    "username": "test1",
    "password": "XXXXXXXXXXX"
}'
```

### 3. 用户登出

```bash
curl --location --request POST 'http://zoo.com/api/user/logout' \
--header 'Content-Type: application/json' \
--data-raw '{
    "Token": "X5ExLrgfCSMop2p9nMdd9E0Dz7s6MvHVu/xWahroTRk="
}'
```

### 4. 刷新token

```bash
curl --location --request POST 'http://zoo.com/api/user/token/refresh' \
--header 'Content-Type: application/json' \
--data-raw '{
    "Token": "9VD6pC705oG6EQmfhEa6axLWXKAhciY264Qnk6SHXdQ=",
    "RefreshToken": "SpGw6VNneKjJCRF1aoPdoRZWyn7n56r4RmyVmQu2eTc="
}'
```

### 5. 查看账号信息

```bash
curl --location --request POST 'http://zoo.com/api/user/get' \
--header 'Content-Type: application/json' \
--data-raw '{
   "Token": "9VD6pC705oG6EQmfhEa6axLWXKAhciY264Qnk6SHXdQ="
}'
```

### 6. 更新账号信息

```bash
curl --location --request POST 'http://zoo.com/api/user/update' \
--header 'Content-Type: application/json' \
--data-raw '{
   "Token": "9VD6pC705oG6EQmfhEa6axLWXKAhciY264Qnk6SHXdQ=",
   "Nickname": "nickxxx",
   "Avatar": "img/c0e299f9becf413bb4fbf7aa4662c795.png"
}'
```

### 7. 查看头像图片

```bash
curl --location --request GET 'http://zoo.com/img/c5c225a013ae49e79120a33bf363e237.png'
```


