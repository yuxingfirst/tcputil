package tcp

import (
	"errors"
	"sync"
)

const (
	_GATEWAY_COMMAND_NONE_       = 0
	_GATEWAY_COMMAND_ADD_CLIENT_ = 1
	_GATEWAY_COMMAND_DEL_CLIENT_ = 2
	_GATEWAY_COMMAND_BROADCAST_  = 3
)

type tcpGatewayLink struct {
	owner        *TcpGatewayFrontend
	id           uint32
	addr         string
	pack         int
	conn         *TcpConn
	clients      map[uint32]*TcpConn
	clientsMutex sync.RWMutex
	maxClientId  uint32
}

func newTcpGatewayLink(owner *TcpGatewayFrontend, id uint32, addr string, pack int, memPool MemPool) (*tcpGatewayLink, error) {
	var (
		this             *tcpGatewayLink
		conn             *TcpConn
		err              error
		beginClientIdMsg []byte
		beginClientId    uint32
	)

	if conn, err = Connect(addr, pack, 0, memPool); err != nil {
		return nil, err
	}

	if beginClientIdMsg = conn.Read(); beginClientIdMsg == nil {
		return nil, errors.New("wait link id failed")
	}

	beginClientId = getUint32(beginClientIdMsg)

	this = &tcpGatewayLink{
		owner:       owner,
		id:          id,
		addr:        addr,
		pack:        pack,
		conn:        conn,
		clients:     make(map[uint32]*TcpConn),
		maxClientId: beginClientId,
	}

	go func() {
		defer this.Close(true)

		for {
			var msg = this.conn.ReadPackage()

			if msg == nil {
				break
			}

			switch msg.ReadUint8() {
			case _GATEWAY_COMMAND_NONE_:
				var clientId = msg.ReadUint32()

				if client := this.GetClient(clientId); client != nil {
					client.sendRaw(msg.Data)
				}
			case _GATEWAY_COMMAND_DEL_CLIENT_:
				var clientId = msg.ReadUint32()

				if client := this.GetClient(clientId); client != nil {
					client.Close()
					this.DelClient(clientId)
				}
			case _GATEWAY_COMMAND_BROADCAST_:
				var (
					idNum   = int(msg.ReadUint16())
					realMsg = msg.Data[4*idNum:]
				)

				for i := 0; i < idNum; i++ {
					var clientId = msg.ReadUint32()

					if client := this.GetClient(clientId); client != nil {
						client.sendRaw(realMsg)
					}
				}
			}
		}
	}()

	return this, nil
}

func (this *tcpGatewayLink) AddClient(client *TcpConn) uint32 {
	this.clientsMutex.Lock()
	defer this.clientsMutex.Unlock()

	this.maxClientId += 1

	this.clients[this.maxClientId] = client

	return this.maxClientId
}

func (this *tcpGatewayLink) DelClient(clientId uint32) {
	this.clientsMutex.Lock()
	defer this.clientsMutex.Unlock()

	delete(this.clients, clientId)
}

func (this *tcpGatewayLink) GetClient(clientId uint32) *TcpConn {
	this.clientsMutex.RLock()
	defer this.clientsMutex.RUnlock()

	if client, exists := this.clients[clientId]; exists {
		return client
	}

	return nil
}

func (this *tcpGatewayLink) SendDelClient(clientId uint32) {
	this.conn.NewPackage(4).WriteUint32(clientId).Send()
}

func (this *tcpGatewayLink) SendToBackend(msg []byte) error {
	return this.conn.sendRaw(msg)
}

func (this *tcpGatewayLink) Close(removeFromFrontend bool) {
	this.clientsMutex.Lock()
	defer this.clientsMutex.Unlock()

	this.conn.Close()

	for _, client := range this.clients {
		client.Close()
	}

	if removeFromFrontend {
		this.owner.delLink(this.id)
	}
}
