# BUG_REPRO

## Bug 是什么

发布文档后接口返回成功，但文档停在 `reviewing` 状态，搜索、热门文档和统计都没有把它当成已发布文档。

## 如何触发

运行 service 包里覆盖文档生命周期、搜索和统计的测试，发布后的状态和下游查询结果会同时失败。

## 错误信息

`service_test.go:58: publish: <nil> status=reviewing`
