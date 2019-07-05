/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2019/6/28
   Description :
-------------------------------------------------
*/

package zttd

import "net"

type Status int32

const (
    // 关闭状态
    isClosedStatus   Status = 0
    // 运行状态
    isRunStatus      Status = 1
    // 关闭中, 这个状态是TcpTransfer独有, Channel不会有这个状态
    isCloseingStatus Status = 2
)

// 通道每次传输数据大小
const ChannelBuffSize = 64 * 1024

// 创建通道时调用函数(谁想连接)是否允许
type CreateChannelHandler func(net.Conn) (allow bool)

// 数据传输时 (通道, 谁发的数据, 数据内容) 返回一个数据用于传输, 空数据不会传输
type DataTransferHandler func(*Channel, net.Conn, []byte) []byte

// 关闭回调函数 (通道)
type CloseCallback func(*Channel)
