package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

const (
	ProtocolID = "/ray/1.0.0"
	ServiceTag = "ray-service"
)

var remotePeer string

var localRay string
var localProxy string

// handleStream 处理传入的 libp2p 流
func handleStream(stream network.Stream) {
	defer stream.Close()

	// 从流中读取数据并转发到本地 Ray 服务
	localConn, err := net.Dial("tcp", localRay) // 假设本地 Ray 服务在端口 6379
	if err != nil {
		log.Println("无法连接到本地 Ray 服务:", err)
		return
	}
	defer localConn.Close()

	go func() {
		_, err := io.Copy(localConn, stream)
		if err != nil {
			log.Println("从流到本地连接复制数据时出错:", err)
		}
	}()

	_, err = io.Copy(stream, localConn)
	if err != nil {
		log.Println("从本地连接到流复制数据时出错:", err)
	}
}

// startLocalListener 启动本地监听器，拦截 Ray 的流量并通过 libp2p 传输
func startLocalListener(host host.Host, remotePeerID peer.ID) {
	listener, err := net.Listen("tcp", localProxy) // 代理监听端口
	if err != nil {
		log.Fatal("无法启动本地监听器:", err)
	}
	defer listener.Close()
	fmt.Println("本地代理Ray端口:", localProxy)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("接受连接时出错:", err)
			continue
		}

		go func() {
			defer conn.Close()

			// 与远程节点建立 libp2p 流
			stream, err := host.NewStream(context.Background(), remotePeerID, ProtocolID)
			if err != nil {
				log.Println("无法建立新的流:", err)
				return
			}
			defer stream.Close()

			go func() {
				_, err := io.Copy(stream, conn)
				if err != nil {
					log.Println("从本地连接到流复制数据时出错:", err)
				}
			}()

			_, err = io.Copy(conn, stream)
			if err != nil {
				log.Println("从流到本地连接复制数据时出错:", err)
			}
		}()
	}
}

func main() {
	flag.StringVar(&remotePeer, "peer", "", "远程节点,/ip4/ip/tcp/port/id")
	flag.StringVar(&localRay, "localray", "127.0.0.1:6379", "本地ray服务,127.0.0.1:6379")
	flag.StringVar(&localProxy, "localproxy", "127.0.0.1:6380", "本地代理服务,127.0.0.1:6380")

	flag.Parse()
	ctx := context.Background()

	// 创建 libp2p 主机
	host, err := libp2p.New()
	if err != nil {
		log.Fatal("无法创建 libp2p 主机:", err)
	}
	defer host.Close()

	// 设置流处理器
	host.SetStreamHandler(ProtocolID, handleStream)

	// 创建 DHT 服务
	dht, err := dht.New(ctx, host)
	if err != nil {
		log.Fatal("无法创建 DHT 服务:", err)
	}
	// 获取并打印节点的所有监听地址
	addrs := host.Addrs()
	fmt.Println("节点监听地址:")
	for _, addr := range addrs {
		fmt.Printf(" - %s/p2p/%s\n", addr, host.ID().String())
	}
	// 等待 DHT 引导完成
	if err = dht.Bootstrap(ctx); err != nil {
		log.Fatal("DHT 引导失败:", err)
	}
	// 目标节点
	if remotePeer != "" {
		peerAddr, err := multiaddr.NewMultiaddr(remotePeer)
		if err != nil {
			log.Fatal(err)
		}
		peerinfo, err := peer.AddrInfoFromP2pAddr(peerAddr)
		if err != nil {
			log.Fatal(err)
		}
		if err := host.Connect(ctx, *peerinfo); err != nil {
			log.Println("连接到目标节点失败:", err)
		} else {
			log.Println("成功连接到目标节点:", peerinfo.ID)
		}

		// 启动本地监听器
		go startLocalListener(host, peerinfo.ID)
	}
	/**
		// 创建服务发现器
		discovery := routing.NewRoutingDiscovery(dht)

		// 创建服务发现器
		discovery.Advertise(ctx, ServiceTag)

		// 等待其他节点发现
		time.Sleep(time.Second * 10)

		// 查找其他节点
		peers, err := discovery.FindPeers(ctx, ServiceTag)
		if err != nil {
			log.Fatal("查找节点失败:", err)
		}

		var remotePeerID peer.ID
		for p := range peers {
			if p.ID == host.ID() {
				continue
			}
			remotePeerID = p.ID
			host.Peerstore().AddAddrs(p.ID, p.Addrs, peerstore.PermanentAddrTTL)
			break
		}

		if remotePeerID == "" {
			log.Println("未找到其他节点，等待连接...")
			select {}
		}
	    **/
	// 保持主程序运行
	select {}
}
