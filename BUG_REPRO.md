# BUG_REPRO

## Bug 是什么

查询不存在的作者、目录、文档或标签时，store 层返回的 `ErrNotFound` 不能被 handler 层识别，HTTP 接口把缺失资源响应成 500。

## 如何触发

在埋错分支运行 handler 包里覆盖作者接口的测试，缺失作者查询会进入服务错误处理逻辑。

## 错误信息

`server_test.go:73: missing status = 500`
