# zttd
###### 一个Tcp数据转移设备

## 获得zttd
`go get -u github.com/zlyuancn/zttd`

## 使用zttd

```go
package main

import "github.com/zlyuancn/zttd"

func main() {
    // 创建一个Tcp数据转移设备, 将3333端口的数据转发到4444端口
    a, err := zttd.NewTcpTransferDevice(":3333", "localhost:4444")
    if err != nil {
        panic(err)
    }
    a.Run()
}
```
