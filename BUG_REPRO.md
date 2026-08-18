# BUG_REPRO

## Bug 是什么

内存 store 的部分列表方法没有读锁，并发写入和遍历同一个 map 时会触发 data race，严重时直接 panic。

## 如何触发

运行 store 包的并发列表与写入测试，并开启 race 检查。

## 错误信息

`fatal error: concurrent map iteration and map write`
