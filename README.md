# 域名要求(接口和回调域名必须使用不相同的主域名)
- 回调域名
- 商户后台域名
- 接口层请求域名
- 管理后台域名

# 软件依赖
- golang 1.19.3
- golang iris v12
- mysql
- redis

# 打包部署程序
- notify.go 8099      订单回调
- main.go 8100        代付订单查询，商户报表生成
- gm.go 8101          运营管理后台
- merchant.go 8102    商户后台
- gate.go 8103        商户api接口
- cb.go  8104         上游回调业务

# 配置 具体看配置文件注释
config.yaml

# 目录配置
> cache 系统缓存

> channels 上游渠道

> config 配置

> controllers 控制器

> logs 日志目录

> menu 后台菜单 

> model 后台菜单

> repositories orm

> resource 资源文件，商户对接文档

> scheduler 定时任务

> services 服务

> static (js,css)

> views 视图

# 开发部署
1. go env -w GOOS=windows
2. go run gm.go
3. 直接访问 127.0.0.1:8081/login
 
# 线上部署
1. 本地打包 go env -w GOOS=linux
2. go build gm.go gate.go cb.go main.go notify.go merchant.go
3. 上传到服务器，赋予777权限，chmod 777 ./main
4. 后台常驻运行并输入日志 nohup ./main  >/data/logs/web/main_listen.log 2>&1 & 