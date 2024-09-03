// SPDX-License-Identifier: Apache-2.0
/**
 * Copyright (c) 2024  Panasonic Automotive Systems, Co., Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ulamulticonn

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"ula-tools/internal/ula"
	. "ula-tools/internal/ulog"
	"net"
	"reflect"
	"strconv"
	"sync"
	"time"
)

var UlaMulCon *UlaMultiConnector

var MAGIC_CODE []byte = []byte{0x55, 0x4C, 0x41, 0x30} // 'ULA0' ascii code

var Mutex struct {
	sync.Mutex
}

type UlaMultiConnector struct {
	targetAddrs []string
	sendChans   []chan string
	respChans   []chan UlaCommandResponse
}

type UlaCommandResponse struct {
	Type   string
	Result int
	Data   []ula.VirtualLayer
}

type DistribNode struct {
	Ip   string
	Port int
}

func newUlaMultiConn(targets []string) error {
	UlaMulCon = &UlaMultiConnector{
		targetAddrs: targets,
	}
	return nil
}

func retryConnectTarget(sockChan chan net.Conn, addr string) {
	for {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			sockChan <- conn
			break
		}
		DLog.Printf("Dial failed connect : %s retry\n", err)
		time.Sleep(10 * time.Millisecond)
	}
}

func connectTarget(addr string) (net.Conn, error) {
	var conn net.Conn

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	sockChan := make(chan net.Conn, 1)

	go retryConnectTarget(sockChan, addr)

	select {
	case <-ctx.Done():
		return nil, errors.New("Dial cannot connect to master")
	case conn = <-sockChan:
		ILog.Println("Dial connected to ", addr)
	}

	return conn, nil
}

func sendCommand(conn net.Conn, command string) ([]byte, uint32, error) {

	n, err := conn.Write(MAGIC_CODE)
	if err != nil || n == 0 {
		ELog.Printf("Write MAGIC_CODE error: %s \n", err)
		return nil, 0, errors.New("MAGIC_CODE Write Fail")
	}

	msgLen := uint32(len(command))
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, msgLen)

	DLog.Println("Write JSON size:", msgLen)
	n, err = conn.Write([]byte(size))
	if err != nil || n == 0 {
		ELog.Printf("Write DATA Size error: %s \n", err)
		return nil, 0, errors.New("Command Size Write Fail")
	}

	n, err = conn.Write([]byte(command))
	if err != nil || n == 0 {
		ELog.Printf("Write error: %s \n", err)
		return nil, 0, errors.New("Command Write Fail")
	}

	n, err = conn.Read(size)
	retSize := binary.BigEndian.Uint32(size)

	if err != nil || n == 0 {
		ELog.Printf("Read error: %s \n", err)
		return nil, 0, errors.New("Response Size Read Fail")
	}

	buf := make([]byte, 0)
	tmp := make([]byte, retSize)
	count := uint32(0)
	for {
		n, err = conn.Read(tmp)
		if err != nil {
			return nil, 0, errors.New("Response Read Fail")
		}
		count += uint32(n)
		buf = append(buf, tmp[:n]...)
		if count >= retSize {
			return buf, retSize, nil
		}
	}
	return buf, retSize, nil
}

func handleConnectTarget(ums *UlaMultiConnector, addr string, sendChan chan string, respChan chan UlaCommandResponse, wg *sync.WaitGroup) {

	var err error
	conn, err := connectTarget(addr)
	if err != nil {
		WLog.Println("Failed connect target: ", addr, " err: ", err)
		wg.Done()
		return
	}
	defer conn.Close()

	ums.sendChans = append(ums.sendChans, sendChan)
	ums.respChans = append(ums.respChans, respChan)
	wg.Done()
	for {
		select {
		case command := <-sendChan:

			respBuf, respSize, err := sendCommand(conn, command)
			var ucr UlaCommandResponse
			if err != nil {
				ELog.Printf("Send command Fail: %s \n", err)
			} else {
				err = json.Unmarshal([]byte(string(respBuf[:respSize])), &ucr)
				if err != nil {
					ELog.Printf("Unmarshal json command error: %s \n", err)
				}
			}

			respChan <- ucr
		}
	}
}

func (ums *UlaMultiConnector) handleConnectTargets(force bool) error {

	var wg sync.WaitGroup
	for _, targetAddr := range ums.targetAddrs {
		wg.Add(1)
		sendChan := make(chan string, 1)
		respChan := make(chan UlaCommandResponse, 1)
		go handleConnectTarget(ums, targetAddr, sendChan, respChan, &wg)
	}
	wg.Wait()

	if len(ums.sendChans) == 0 || len(ums.respChans) == 0 {
		return errors.New("All targets cannot connect master")
	}

	if !force {
		if len(ums.sendChans) < len(ums.targetAddrs) || len(ums.respChans) < len(ums.targetAddrs) {
			return errors.New(fmt.Sprintf("Some targets cannot connect master (%d < %d)", len(ums.sendChans), len(ums.targetAddrs)))
		}
	}

	return nil
}

func waitResponse(waitTime time.Duration, respChan chan UlaCommandResponse, targetAddr string, wg *sync.WaitGroup, resp *UlaCommandResponse) {

	t := time.NewTicker(waitTime * time.Second)
	defer t.Stop()
	defer wg.Done()

	select {
	case ucr := <-respChan:
		*resp = ucr
		break
	case <-t.C:
		ELog.Printf("Command response watchdog was timeout. target: %s", targetAddr)
		break
	}
}

func (ums *UlaMultiConnector) sendCommand(command string) UlaCommandResponse {

	Mutex.Lock()
	var wg sync.WaitGroup
	resps := make([]UlaCommandResponse, len(ums.sendChans))
	for chanId, sendChan := range ums.sendChans {
		wg.Add(1)
		sendChan <- command
		go waitResponse(1, ums.respChans[chanId], ums.targetAddrs[chanId], &wg, &resps[chanId])
	}
	wg.Wait()

	ret := mergeResponses(resps)
	Mutex.Unlock()

	return ret
}

func (ums *UlaMultiConnector) SendLayoutCommand(command string) error {
	if len(ums.sendChans) == 0 {
		return errors.New("Don't have connection, so cannot send layout commands")
	}
	ucr := ums.sendCommand(command)
	if ucr.Type == "result" {
		ret := ucr.Result
		if ret != 0 {
			return errors.New("SendLayoutCommand Failed")
		}
	} else {
		return errors.New("result formart type miss matched")
	}

	return nil
}

func mergeResponses(resps []UlaCommandResponse) UlaCommandResponse {
	var ret UlaCommandResponse
	for _, resp := range resps {
		ret.Type = resp.Type
		switch ret.Type {
		case "result":
			ret.Result = ret.Result | resp.Result
			break

		case "data":
			if !reflect.DeepEqual(ret.Data, resp.Data) {
				ret.Data = resp.Data
			}
			break
		}
	}
	return ret

}

func getDistribNodes(vsdPath ...string) ([]DistribNode, error) {
	var dNodes []DistribNode
	vscrnDef, err := ula.ReadVScrnDef(vsdPath...)
	if err != nil {
		return dNodes, err
	}
	dNodes = make([]DistribNode, 0)
	for _, node := range vscrnDef.Nodes {
		for _, frameworkNode := range vscrnDef.DistributedWindowSystem.FrameworkNode {
			if node.NodeId != frameworkNode.NodeId {
				continue
			}

			var tmp DistribNode
			tmp.Ip = node.Ip
			tmp.Port = frameworkNode.Ula.Port
			dNodes = append(dNodes, tmp)
		}

	}
	return dNodes, nil
}

func UlaConnectionInit(force bool, vsdPath ...string) error {
	var err error
	dNodes, err := getDistribNodes(vsdPath...)
	if err != nil {
		ELog.Println(err)
		return err
	}
	targetNum := 0
	var targetAddrs []string

	for _, d := range dNodes {
		targetAddr := d.Ip + ":" + strconv.Itoa(d.Port)
		targetAddrs = append(targetAddrs, targetAddr)
		targetNum++
	}

	err = newUlaMultiConn(targetAddrs)

	if err != nil {
		ELog.Println(err)
		return err
	}

	err = UlaMulCon.handleConnectTargets(force)
	if err != nil {
		return err
	}
	return nil
}
