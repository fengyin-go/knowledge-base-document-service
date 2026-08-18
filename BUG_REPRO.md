# BUG_REPRO

## Bug 是什么

删除文档后，关联的版本、评论和收藏没有被完整清理，后续列表仍能看到脏数据。

## 如何触发

运行 service 包里的级联删除测试，先创建文档、评论和收藏，再删除文档并检查关联列表。

## 错误信息

`service_test.go:217: comments should be empty`
