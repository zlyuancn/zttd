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
    "sync"
    "sync/atomic"
)

type TcpTransferDevice struct {
    mx                     *sync.Mutex
    bindaddr               string
    targetaddr             string
    listener               net.Listener
    isrun                  Status
    channels               *list.List
    create_channel_handler CreateChannelHandler
    data_handler           DataTransferHandler
    close_callback         CloseCallback
}

// 创建一个Tcp数据转移设备(绑定地址, 目标地址)
func NewTcpTransferDevice(bindaddr string, targetaddr string) (*TcpTransferDevice, error) {
    listener, err := net.Listen("tcp", bindaddr)
    if err != nil {
        return nil, err
    }

    return &TcpTransferDevice{
        mx:         new(sync.Mutex),
        bindaddr:   bindaddr,
        targetaddr: targetaddr,
        listener:   listener,
        isrun:      isClosedStatus,
        channels:   list.New(),
    }, nil
}

// 启动
func (m *TcpTransferDevice) Run() {
    if !atomic.CompareAndSwapInt32((*int32)(&m.isrun), int32(isClosedStatus), int32(isRunStatus)) {
        return
    }

    listener := m.listener
    for atomic.LoadInt32((*int32)(&m.isrun)) == int32(isRunStatus) {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }
        go m.createChannel(conn)
    }
}

// 关闭(会关闭所有已创建的通道)
func (m *TcpTransferDevice) Close() {
    if !atomic.CompareAndSwapInt32((*int32)(&m.isrun), int32(isRunStatus), int32(isCloseingStatus)) {
        return
    }

    m.mx.Lock()
    channels := m.channels
    m.channels = list.New()

    wg := sync.WaitGroup{}
    wg.Add(channels.Len())

    e := channels.Front()
    for {
        if e == nil {
            break
        }

        go func(c *Channel) {
            c.Close()
            wg.Done()
        }(e.Value.(*Channel))

        e = e.Next()
    }

    wg.Wait()

    m.mx.Unlock()

    m.isrun = isClosedStatus
}

// 设置创建通道Hook函数, 只能在Run之前设置
func (m *TcpTransferDevice) SetCreateChannelHandler(fn CreateChannelHandler) {
    if atomic.LoadInt32((*int32)(&m.isrun)) == int32(isClosedStatus) {
        m.create_channel_handler = fn
    }
}

// 设置数据传输Hook函数, 只能在Run之前设置
func (m *TcpTransferDevice) SetDataTransferHandler(fn DataTransferHandler) {
    if atomic.LoadInt32((*int32)(&m.isrun)) == int32(isClosedStatus) {
        m.data_handler = fn
    }
}

// 设置关闭回调函数, 只能在Run之前设置
func (m *TcpTransferDevice) SetCloseCallBack(fn CloseCallback) {
    if atomic.LoadInt32((*int32)(&m.isrun)) == int32(isClosedStatus) {
        m.close_callback = fn
    }
}

func (m *TcpTransferDevice) _CreateChannelHandler(conn net.Conn) (allow bool) {
    if atomic.LoadInt32((*int32)(&m.isrun)) == int32(isRunStatus) {
        if m.create_channel_handler != nil {
            return m.create_channel_handler(conn)
        }
        return true
    }
    return false
}
func (m *TcpTransferDevice) _DataHandler(c *Channel, conn net.Conn, data []byte) []byte {
    if atomic.LoadInt32((*int32)(&m.isrun)) == int32(isRunStatus) {
        if m.data_handler != nil {
            return m.data_handler(c, conn, data)
        }
        return data
    }
    return nil
}
func (m *TcpTransferDevice) _CloseCallBack(c *Channel) {
    status := Status(atomic.LoadInt32((*int32)(&m.isrun)))
    if status == isRunStatus || status == isClosedStatus {
        m.mx.Lock()
        m.channels.Remove(c.le)
        m.mx.Unlock()
    }

    if status == isRunStatus || status == isCloseingStatus {
        if m.close_callback != nil {
            m.close_callback(c)
        }
    }
}

func (m *TcpTransferDevice) createChannel(cconn net.Conn) {
    if !m._CreateChannelHandler(cconn) {
        _ = cconn.Close()
        return
    }

    sconn, err := net.Dial("tcp", m.targetaddr)
    if err != nil {
        _ = cconn.Close()
        return
    }

    c := NewChannel(cconn, sconn)

    m.mx.Lock()
    c.le = m.channels.PushBack(c)
    c.SetDataTransferHandler(m._DataHandler)
    c.SetCloseCallBack(m._CloseCallBack)
    m.mx.Unlock()

    c.Run()
}

// 获取绑定地址
func (m *TcpTransferDevice) BindAddr() string {
    return m.bindaddr
}

// 获取目标地址
func (m *TcpTransferDevice) TargetAddr() string {
    return m.targetaddr
}

// 获取设备状态
func (m *TcpTransferDevice) Status() Status {
    return Status(atomic.LoadInt32((*int32)(&m.isrun)))
}

// 获取通道数量
func (m *TcpTransferDevice) ChannelCount() int {
    m.mx.Lock()
    defer m.mx.Unlock()
    return m.channels.Len()
}
