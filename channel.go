/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2019/6/27
   Description :
-------------------------------------------------
*/

package zttd

import (
    "container/list"
    "net"
    "sync/atomic"
)

type Channel struct {
    a              net.Conn
    b              net.Conn
    data_handler   DataTransferHandler
    close_callback CloseCallback
    isrun          Status
    le             *list.Element
}

// 创建一个通道
func NewChannel(a net.Conn, b net.Conn) *Channel {
    return &Channel{
        a: a,
        b: b,
    }
}

// 设置数据传输Hook函数, 只能在Run之前设置
func (m *Channel) SetDataTransferHandler(fn DataTransferHandler) {
    if atomic.LoadInt32((*int32)(&m.isrun)) == int32(isClosedStatus) {
        m.data_handler = fn
    }
}

// 设置关闭回调函数, 只能在Run之前设置
func (m *Channel) SetCloseCallBack(fn CloseCallback) {
    if atomic.LoadInt32((*int32)(&m.isrun)) == int32(isClosedStatus) {
        m.close_callback = fn
    }
}

// 获取a端conn, 一般是接入端
func (m *Channel) A() net.Conn {
    return m.a
}

// 获取b端conn, 一般是目标端
func (m *Channel) B() net.Conn {
    return m.b
}

// 开始数据传输
func (m *Channel) Run() {
    if atomic.CompareAndSwapInt32((*int32)(&m.isrun), int32(isClosedStatus), int32(isRunStatus)) {
        go m.transfer(m.a, m.b)
        go m.transfer(m.b, m.a)
    }
}

// 关闭通道
func (m *Channel) Close() {
    m._CloseHandler()
}

func (m *Channel) _DataHandler(conn net.Conn, data []byte) []byte {
    if atomic.LoadInt32((*int32)(&m.isrun)) == int32(isRunStatus) {
        if m.data_handler != nil {
            return m.data_handler(m, conn, data)
        }
        return data
    }
    return nil
}

func (m *Channel) _CloseHandler() {
    if atomic.CompareAndSwapInt32((*int32)(&m.isrun), int32(isRunStatus), int32(isClosedStatus)) {
        _ = m.a.Close()
        _ = m.b.Close()
        if m.close_callback != nil {
            m.close_callback(m)
        }
    }
}

func (m *Channel) transfer(src net.Conn, target net.Conn) {
    for {
        buff := make([]byte, ChannelBuffSize)
        length, err := src.Read(buff)
        if err != nil {
            m._CloseHandler()
            return
        }

        data := m._DataHandler(src, buff[:length])
        if len(data) != 0 {
            _, _ = target.Write(data)
        }
    }
}

// 获取通道状态
func (m *Channel) Status() Status {
    return Status(atomic.LoadInt32((*int32)(&m.isrun)))
}
