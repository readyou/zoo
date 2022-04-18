

## 说明
临时性使用，没有过多优化。

/opt/data/user.sql中的内容是直接从数据库中导出的。

```sql
select username into outfile '/opt/data/user.sql' from user limit 500000 
```

## 压测步骤
register -> login -> getprofile -> upload -> updateprofile -> refresh_token -> logout

中间几个接口需要依赖token，login压测后会把token写入到`/opt/data/token.txt`和`/opt/data/refresh.txt`中。

后面的接口压测会读取上面的文件来生成请求的值。
