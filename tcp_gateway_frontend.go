package tcp

import (
	"sync"
)

//
// 网关前端
//
type TcpGatewayFrontend struct {
	server     *TcpListener
	pack       int
	memPool    MemPool
	links      map[uint32]*tcpGatewayLink
	linksMutex sync.RWMutex
}

//
// 在指定地址和端口创建一个网关前端，连接到指定的网关后端，并等待客户端接入。
// 新接入的客户端首先需要发送一个uint32类型的后端ID，选择客户端实际所要连接的后端。
//
func NewTcpGatewayFrontend(addr string, pack int, memPool MemPool, backends map[uint32]string) (*TcpGatewayFrontend, error) {
	server, err := Listen(addr, pack, pack+4, memPool)

	if err != nil {
		return nil, err
	}

	var this = &TcpGatewayFrontend{
		server:  server,
		pack:    pack,
		memPool: memPool,
		links:   make(map[uint32]*tcpGatewayLink),
	}

	this.UpdateBackends(backends)

	go func() {
		for {
			var client = this.server.Accpet()

			if client == nil {
				break
			}

			go func() {
				defer func() {
					client.Close()
				}()

				var link, clientId = this.clientInit(client)

				if link == nil || clientId == 0 {
					return
				}

				defer func() {
					link.DelClient(clientId)
					link.SendDelClient(clientId)
				}()

				for {
					var msg = client.Read()

					if msg == nil {
						break
					}

					setUint(msg, pack, len(msg)-pack)

					setUint32(msg[pack:], clientId)

					link.SendToBackend(msg)
				}
			}()
		}
	}()

	return this, nil
}

func (this *TcpGatewayFrontend) clientInit(client *TcpConn) (link *tcpGatewayLink, clientId uint32) {
	var (
		serverIdMsg []byte
		serverId    uint32
	)

	if serverIdMsg = client.Read(); len(serverIdMsg) != this.pack+4+4 {
		return
	}

	serverId = getUint32(serverIdMsg[this.pack+4:])

	if link = this.getLink(serverId); link == nil {
		return
	}

	if clientId = link.AddClient(client); clientId == 0 {
		return
	}

	return
}

func (this *TcpGatewayFrontend) addLink(id uint32, link *tcpGatewayLink) {
	this.linksMutex.Lock()
	defer this.linksMutex.Unlock()

	this.links[id] = link
}

func (this *TcpGatewayFrontend) delLink(id uint32) {
	this.linksMutex.Lock()
	defer this.linksMutex.Unlock()

	delete(this.links, id)
}

func (this *TcpGatewayFrontend) getLink(id uint32) *tcpGatewayLink {
	this.linksMutex.RLock()
	defer this.linksMutex.RUnlock()

	if link, exists := this.links[id]; exists {
		return link
	}

	return nil
}

func (this *TcpGatewayFrontend) removeOldLinks(backends map[uint32]string) {
	this.linksMutex.Lock()
	defer this.linksMutex.Unlock()

	for id, link := range this.links {
		if _, exists := backends[id]; !exists || link.addr != backends[id] {
			link.Close(false)
		}
	}
}

//
// 更新网关后端，移除地址有变化或者已经不在新配置里的久连接，创建久配置中没有的连接。
//
func (this *TcpGatewayFrontend) UpdateBackends(backends map[uint32]string) {
	this.removeOldLinks(backends)

	for id, addr := range backends {
		var link = this.getLink(id)

		if link != nil {
			continue
		}

		link, _ = newTcpGatewayLink(this, id, addr, this.pack, this.memPool)

		if link != nil {
			this.addLink(id, link)
		}
	}
}

//
// 你懂的。
//
func (this *TcpGatewayFrontend) Close() {
	this.linksMutex.Lock()
	defer this.linksMutex.Unlock()

	this.server.Close()

	for _, link := range this.links {
		link.Close(false)
	}
}
