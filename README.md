## 工程目录结构说明

| 目录          | 说明                         |
|-------------|----------------------------|
| ant         | 自己实现的web框架                 |
| api         | 接口参数相关                     |
| bee         | 自己实现的rpc框架                 |
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
