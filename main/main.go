package main

import (
	"fmt"
	"log"
	"net"
	rpc_go "rpc-go"
	"rpc-go/client"
	"sync"
	"time"
)

func startServer(addr chan string) {
	// pick a free port
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	rpc_go.Accept(l)
}

//func main() {
//	addr := make(chan string)
//	go startServer(addr)
//
//	// in fact, following code is like a simple rpc_go client
//	conn, _ := net.Dial("tcp", <-addr)
//	defer func() {
//		_ = conn.Close()
//	}()
//	time.Sleep(time.Second)
//
//	// send options
//
//	_ = json.NewEncoder(conn).Encode(rpc_go.DefaultOption)
//	cc := codec.NewGobCodec(conn)
//
//	// send request & receive response
//	for i := 0; i < 5; i++ {
//		h := &codec.Header{
//			ServiceMethod: "Foo.Sum",
//			Seq:           uint64(i),
//		}
//
//		// 向server写
//		_ = cc.Write(h, fmt.Sprintf("rpc-go req %d", h.Seq))
//
//		// 从server返回中读取header
//		_ = cc.ReadHeader(h)
//		// 从server返回中读取body
//		var reply string
//		_ = cc.ReadBody(&reply)
//		log.Println("reply:", reply)
//	}
//
//}

func main() {
	log.SetFlags(0)

	addr := make(chan string)
	go startServer(addr)

	cli, _ := client.Dial("tcp", <-addr)
	defer func() { _ = cli.Close() }()

	time.Sleep(time.Second)
	// send request & receive response

	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := fmt.Sprintf("rpc-go req %d", i)
			var reply string
			if err := cli.Call("Foo.Sum", args, &reply); err != nil {
				log.Fatal("call foo.Sum error:", err)
			}
			log.Println("reply:", reply)
		}(i)
	}
	wg.Wait()
}