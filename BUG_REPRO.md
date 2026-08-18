# BUG_REPRO

## Bug 是什么

文档创建后，标签在创建、统计、导出链路里会被改成小写，导致按原始标签名筛选不到。

## 如何触发

运行 service 包里 `TestDocumentFilterByTag`，创建 `Go` 标签的文档后再按 `Go` 过滤会得到空结果。

## 错误信息

`service_test.go:354: filter by Go tag: total=0`
