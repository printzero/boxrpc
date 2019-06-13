/*
 * Copyright (c) 2019 Ashish Shekar
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 *
 */

package boxrpc

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	GobEnde = iota + 1
	JSONEnde
)

var WaitInterval = time.Second * 2
var counter = 0

type boxConn struct {
	sync.Mutex
	authDone bool
	conn     net.Conn
	reader   *bufio.Reader
	//ConnectionHandler
}

//type ConnectionHandler struct {
//	BeforeExec       func()
//	AfterExec        func()
//	OnAuthentication func() bool
//	OnConnect        func()
//	OnDisconnect     func()
//}

//type ActionCall struct {
//	sync.Mutex
//	Context   context.Context
//	Authority string      `json:"authority"`
//	Ende      string      `json:"ende"`
//	Payload   interface{} `json:"payload"`
//}
//
//type ActionError struct {
//	Message string
//	Status  string
//	Ok      bool
//}

//type Ende interface {
//	Encode()
//	Decode()
//}

//type actions map[string]ActionCall

//type endeSlice map[string]Ende

func Register(action string, actionFunc func() error) {
}

func UnRegister(action string) {
}

//
//func (bc BoxConn) SetEnde(ende Ende) {}
//
//func defConnectionHandler() ConnectionHandler {
//	ch := ConnectionHandler{}
//	ch.OnConnect = defOnConnect
//	ch.OnDisconnect = defOnDisconnect
//	ch.OnAuthentication = defOnAuth
//	return ch
//}
//
//func defOnConnect() {
//}
//
//func defOnDisconnect() {
//
//}
//
//func defOnAuth() bool {
//	return false
//}

// ---------------------------

// Listen hooks to a specific socket address in your network and awaits for client connection
func Listen(ctx context.Context, network, addr string) error {
	ctx, _ = context.WithCancel(ctx)

	listener, err := net.Listen(network, addr)

	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return errors.New("boxrpc: exited due to an error in function that caused context cancellation")
	default:
		break
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		bc := boxConn{
			authDone: false,
			conn:     conn,
		}

		go handleConn(ctx, bc)
	}

}

func handleConn(ctx context.Context, bConn boxConn) {
	counter++
	log.Println("Accepting new conn:", counter)
	bConn.reader = bufio.NewReader(bConn.conn)
	exitChan := make(chan int, 1)
	defer bConn.conn.Close()

	// check for _auth action
	authTimer := time.AfterFunc(WaitInterval, func() {
		if !bConn.authDone {
			_, _ = bConn.conn.Write([]byte("_ERRNOAUTH"))
			exitChan <- 1
		}
	})
	// reader blocking code
	go readAction(authTimer, &exitChan, bConn)

	<-exitChan
}

func readAction(t *time.Timer, exitChan *chan int, bConn boxConn) {
	for {
		action, err := bConn.reader.ReadString('\n')
		action = strings.TrimRight(action, "\n")
		fmt.Println("Action:", action)

		if err != nil {
			*exitChan <- 1
		}

		if action == "A" {
			bConn.Lock()
			t.Stop()
			bConn.authDone = true
			bConn.Unlock()
			_, _ = bConn.conn.Write([]byte("COOL"))
		}

		if err == io.EOF {
			fmt.Println("1 connection down")
		}
	}
}

//
//func ListenTLS(ctx context.Context, network, addr string, config tls.Config) (BoxConn, error) {
//	return BoxConn{}, nil
//}
