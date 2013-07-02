package tcputil

import (
	"sync"
)

const (
	_GATEWAY_MAX_LINKS_ = 256
)

//
// 网关后端
//
type TcpGatewayBackend struct {
	server     *TcpListener
	links      []*TcpConn
	linksMutex sync.RWMutex
	In         chan *TcpGatewayIntput
}

//
// 来自实际客户端的消息，用法跟TcpInput是一样的，区别是多了ClientId
//
type TcpGatewayIntput struct {
	ClientId uint32
	*TcpInput
}

//
// 在指定的地址和端口创建一个网关后端，等待网关前端连接。
// 一个网关后端可以被多个网关前端连接，客户端ID分配算法会保证不同网关前端的客户端ID不冲突。
//
func NewTcpGatewayBackend(addr string, pack int, memPool MemPool) (*TcpGatewayBackend, error) {
	var server, err = Listen(addr, pack, 0, memPool)

	if err != nil {
		return nil, err
	}

	var this = &TcpGatewayBackend{
		server: server,
		links:  make([]*TcpConn, _GATEWAY_MAX_LINKS_),
		In:     make(chan *TcpGatewayIntput),
	}

	go func() {
		for {
			var link = this.server.Accpet()

			if link == nil {
				break
			}

			go func() {
				defer func() {
					link.Close()
				}()

				var linkId = this.addLink(link)

				if linkId < 0 {
					return
				}

				defer func() {
					this.delLink(linkId)
				}()

				for {
					var msg = link.ReadPackage()

					if msg == nil {
						break
					}

					this.In <- &TcpGatewayIntput{msg.ReadUint32(), msg}
				}
			}()
		}

		this.In <- nil
	}()

	return this, nil
}

func (this *TcpGatewayBackend) addLink(link *TcpConn) int {
	this.linksMutex.Lock()
	defer this.linksMutex.Unlock()

	for id := 0; id < len(this.links); id++ {
		if this.links[id] == nil {
			var serverIdMsg = link.NewPackage(4)

			serverIdMsg.WriteUint32(uint32(id) << 24)

			if err := serverIdMsg.Send(); err != nil {
				return -1
			}

			this.links[id] = link

			return id
		}
	}

	return -1
}

func (this *TcpGatewayBackend) delLink(linkId int) {
	this.linksMutex.Lock()
	defer this.linksMutex.Unlock()

	this.links[linkId] = nil
}

func (this *TcpGatewayBackend) getLink(clientId uint32) *TcpConn {
	this.linksMutex.RLock()
	defer this.linksMutex.RUnlock()

	return this.links[int(clientId>>24)]
}

//
// 你懂的。
//
func (this *TcpGatewayBackend) Close() {
	this.linksMutex.Lock()
	defer this.linksMutex.Unlock()

	this.server.Close()

	for _, link := range this.links {
		if link != nil {
			link.Close()
		}
	}
}

//
// 创建一个发送给指定客户端的消息包
//
func (this *TcpGatewayBackend) NewPackage(clientId uint32, size int) *TcpOutput {
	var link = this.getLink(clientId)

	if link == nil {
		return nil
	}

	// [gateway command](1) + [client id](4) + [real package size](pack) + [real package content](return)
	return link.NewPackage(1+4+link.pack+size).WriteUint8(_GATEWAY_COMMAND_NONE_).WriteUint32(clientId).WriteUint(link.pack, uint64(size))
}

//
// 告诉网关前端，移除一个客户端
//
func (this *TcpGatewayBackend) DelClient(clientId uint32) {
	var link = this.getLink(clientId)

	if link == nil {
		return
	}

	// [gateway command](1) + [client id](4)
	link.NewPackage(1 + 4).WriteUint8(_GATEWAY_COMMAND_DEL_CLIENT_).WriteUint32(clientId).Send()
}

//
// 创建一个发送给指定客户端的广播包，广播包的用法跟'TcpOutput'包一样
//
func (this *TcpGatewayBackend) NewBroadcast(clientIds []uint32, size int) *TcpBroadcast {
	var link *TcpConn

	// find first usable link
	for i := 0; i < len(clientIds); i++ {
		link = this.getLink(clientIds[i])

		if link != nil {
			break
		}
	}

	if link == nil {
		return nil
	}

	// [gateway command](1) + [client id list length](2) + [client id list](4 x len) + [real pack size](pack) + [real package content](return)
	var (
		output = link.NewPackage(1 + 2 + 4*len(clientIds) + link.pack + size)
		idNum  = len(clientIds)
	)

	output.WriteUint8(_GATEWAY_COMMAND_BROADCAST_).WriteUint16(uint16(idNum))

	for i := 0; i < idNum; i++ {
		output.WriteUint32(clientIds[i])
	}

	output.WriteUint(link.pack, uint64(size))

	return &TcpBroadcast{this, output}
}

//
// 网关广播实例
//
type TcpBroadcast struct {
	owner *TcpGatewayBackend
	*TcpOutput
}

func (this *TcpBroadcast) Send() error {
	this.owner.linksMutex.RLock()
	defer this.owner.linksMutex.RUnlock()

	var err error

	for _, link := range this.owner.links {
		if link != nil {
			this.TcpOutput.owner = link
			err = this.TcpOutput.Send()
		}
	}

	return err
}
