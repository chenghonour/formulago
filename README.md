# FormulaGo
[English](README_EN.md) | 中文

## 简介
一个高性能的企业后台管理框架，使用`Hertz`与`Ent`
- 高生产率：短时间即可搭建一个企业管理系统。
- 高性能：使用Go里最强性能的 [Hertz 框架](https://github.com/cloudwego/hertz)。字节跳动已经在上万个服务上部署应用。
- 模块化：参考DDD设计理念，模块设计更简洁、更方便。
- 路由接口：参考[ Google 开放平台](https://github.com/googleapis/googleapis)，使用 Protobuf 定义接口规范。
- 面向接口开发，更好拓展与单元测试。

## 演示网站
- 地址: [https://formulago.com](https://base.hcteam.vip)
- 账号: admin/formulago
- 前端项目地址: [https://formulago-ui.com](https://github.com/czx-ly/formulago-ui)

## 架构图
`看起来像F1赛车吗?`
![Go Backend Clean Architecture](./formulago.png)

## 依赖
- 使用 `Hertz` 作为 HTTP 框架
- 使用 `Protobuf` IDL 定义 `HTTP` 接口
- 使用 `hz` 工具进行代码生成
- 使用 `Ent` 与 `MySQL`(你也可以使用PostgresSQL)

## 内置功能
1. 用户管理：用户是系统操作者，该功能主要完成系统用户配置。
2. 菜单管理：配置系统菜单，操作权限，按钮权限标识等。
3. 角色管理：角色菜单权限分配、设置角色的权限划分。
4. 字典管理：对系统中经常使用的一些较为固定的数据进行维护。
5. 操作日志：系统正常操作日志记录和查询；系统异常信息日志记录和查询。
6. 在线用户：当前系统中活跃用户Token状态监控。
7. 文件管理：文件上传，S3(Aliyun OSS)多种上传方式适配。
8. OAuth2.0登录：支持Google, Github, Wecom 等OAuth2.0认证登录, 可以自己拓展。
9. 常用工具：在pkg包集成常用的工具包。
10. 开发工具：提供便捷的Struct与Protobuf转换工具，Struct to Protobuf，Delete Struct Tag等。

## 接口定义
本项目使用`Protobuf` IDL 定义`HTTP` 接口。对应的admin模块相关接口在 [admin.proto](api/admin/admin.proto) 文件中定义。

## 代码生成

本项目使用 `hz` 生成代码. `hz` 详细使用说明参考 [hz](https://www.cloudwego.io/docs/hertz/tutorials/toolkit/toolkit/).
- hz install.
```bash
go install github.com/cloudwego/hertz/cmd/hz@latest
```
- hz new: 新建一个 Hertz 项目
```bash
hz new -I api -idl api/admin/admin.proto -model_dir api/model -module formulago --unset_omitempty
```
- hz update: 当你的IDL文件更新，使用该指令进行项目代码更新
```bash
hz update -I api -idl api/admin/admin.proto -model_dir api/model --unset_omitempty
```

## 变量绑定与校验

变量绑定与校验使用说明参考 [Binding and Validate](https://www.cloudwego.io/docs/hertz/tutorials/basic-feature/binding-and-validate/).

## Ent

ent - 一个简单但功能强大的 Go 实体框架。

本项目使用 `Ent` 连接与操作 `MySQL`(你也可以使用PostgresSQL) ，详细使用说明参考 [Ent](https://github.com/ent/ent).

#### 快速开始

- 将配置文件中数据库参数换成你自己的： [Database config file](configs/config.yaml).
- 在项目根目录下依次执行以下命令, 将在目录 data/ent/schema/ 生成一个User的实体:
```bash
  cd data
  go run -mod=mod entgo.io/ent/cmd/ent init User
  ```
- 在 User schema 中添加表字段, 运行以下命令将生成Ent的操作代码文件
```bash
  go generate ./ent
  ```
- Ent的更多使用说明，请参考 [Ent Guides](https://entgo.io/).

## 如何运行

### 更新配置文件
- 使用你自己的参考更新 [Prod configuration file](configs/config.yaml) 与 [Dev configuration file](configs/config_dev.yaml)。
- 注意，YAML文件的参数结构要与config.go里的Config结构体一致。
- 当运行时环境变量 "IS_PROD" 等于 true, 将会使用生产环境Prod配置, 否则使用测试环境Dev配置。

### 使用 Docker 启动 MySQL
```bash
cd formulago && docker-compose up
```

### 运行本项目
```
cd formulago
go build -o formulago &&./formulago

# 项目运行后，HTTP Get请求以下路由，将会初始化数据库表数据
# 初始管理员账号: admin/formulago
@router yourHost/api/initDatabase [GET]
enjoy it!
```

### 欢迎贡献代码或提供建议