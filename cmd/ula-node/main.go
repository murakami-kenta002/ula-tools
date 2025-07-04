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

package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"strconv"
	_ "strings"
	"sync"
	_ "time"
	"ula-tools/internal/ula"
	"ula-tools/internal/ula-node"
	"ula-tools/internal/ula-node/iviwinmgr"
	"ula-tools/internal/ula-node/rvgpuwinmgr"
	"ula-tools/internal/ula-node/vs2rd"
	. "ula-tools/internal/ulog"
)

var MAGIC_CODE []byte = []byte{0x55, 0x4C, 0x41, 0x30} // 'ULA0' ascii code

var Mutex struct {
	sync.Mutex
}

func isIpv4(ip string) bool {
	if net.ParseIP(ip) != nil {
		for i := 0; i < len(ip); i++ {
			switch ip[i] {
			case '.':
				return true
			case ':':
				return false
			}
		}
	}

	return false
}

func getIpv4AddrsOfAllInterfaces() ([]string, error) {

	var ipaddrs []string

	ifaces, err := net.Interfaces()
	if err != nil {
		return ipaddrs, err
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if isIpv4(ip.String()) {
				ipaddrs = append(ipaddrs, ip.String())
			}
		}
	}

	if len(ipaddrs) == 0 {
		return ipaddrs, errors.New("Cannot Find Ipv4Addr from Interfaces")
	} else {
		return ipaddrs, nil
	}

}

func getIpAddrByNodeIdAndIpCandidateList(ipAddrs []string, nodeId int, vscrnDef *ula.VScrnDef) (string, error) {

	for _, r := range vscrnDef.Nodes {
		if nodeId == r.NodeId {
			for _, ipaddr := range ipAddrs {
				if ipaddr == r.Ip {
					return ipaddr, nil
				}
			}
		}
	}

	return "0.0.0.0", errors.New("The acquired IP address does not exist in VScrnDef json")
}

func getNodeIdByHostName(hostname string, vscrnDef *ula.VScrnDef) (int, error) {

	for _, r := range vscrnDef.Nodes {
		if hostname == r.HostName {
			return r.NodeId, nil
		}
	}

	return -1, errors.New("Cannot Find My NodeId from VScrnDef json")
}

func getPort(nodeId int, vscrnDef *ula.VScrnDef) (int, error) {

	for _, r := range vscrnDef.DistributedWindowSystem.FrameworkNode {
		if nodeId == r.NodeId {
			return r.Ula.Port, nil
		}
	}

	return -1, errors.New("Cannot Find My Port from VScrnDef json")
}

func readConnection(conn net.Conn) ([]byte, uint32, error) {

	cbio := bufio.NewReader(conn)

	magicBuf := make([]byte, 4)

	_, err := io.ReadFull(cbio, magicBuf[:4])
	if err != nil {
		if err == io.EOF {
			return nil, 0, err
		} else {
			DLog.Printf("Magic Size Read Failed: %s\n", err)
			return nil, 0, errors.New("zero byte read(maybe connection check)")
		}
	}
	if reflect.DeepEqual(magicBuf, MAGIC_CODE) == false {
		return nil, 0, errors.New("Magic Code Read Fail")
	}

	szBuf := make([]byte, 4)
	_, err = io.ReadFull(cbio, szBuf[:4])
	if err != nil {
		return nil, 0, errors.New("Command Size Read Fail")
	}

	recvSize := binary.BigEndian.Uint32(szBuf[:4])
	if recvSize == 0 {
		return nil, recvSize, errors.New("zero byte read(meaningless)")
	}

	recvBuf := make([]byte, recvSize)
	_, err = io.ReadFull(cbio, recvBuf[:recvSize])
	if err != nil {
		return recvBuf, recvSize, errors.New(fmt.Sprintf("Command Read Fail: %s \n", err))
	}
	return recvBuf, recvSize, nil
}

func readCommandLoop(conn net.Conn, jsonChan chan map[string]interface{}, listenerId int, retChansMap map[int]interface{}) {

	defer conn.Close()

	for {
		recvBuf, recvSize, err := readConnection(conn)
		if err != nil {
			if err == io.EOF {
				DLog.Printf("Ula-node zero byte read(maybe Client closed the connection)\n")
			} else {
				ELog.Printf("Ula-node command Read Fail: %s \n", err)
			}
			break
		}
		mJson := make(map[string]interface{})
		err = json.Unmarshal([]byte(string(recvBuf[:recvSize])), &mJson)
		if err != nil {
			ELog.Printf("Unmarshal json command error: %s \n", err)
			break
		}

		mJson["listener_id"] = listenerId

		jsonChan <- mJson

		Mutex.Lock()
		retChan := retChansMap[listenerId].(chan map[string]interface{})
		Mutex.Unlock()
		select {
		case retJson := <-retChan:
			retBuf, _ := json.Marshal(retJson)
			msgLen := uint32(len(retBuf))
			size := make([]byte, 4)
			binary.BigEndian.PutUint32(size, msgLen)
			n, err := conn.Write([]byte(size))
			if err != nil || n == 0 {
				ELog.Printf("Write DATA Size error: %s \n", err)
			}

			n, err = conn.Write(retBuf)
			if err != nil || n == 0 {
				ELog.Printf("Write error: %s \n", err)
			}
			break
		}

	}
	Mutex.Lock()
	delete(retChansMap, listenerId)
	Mutex.Unlock()
}

func processCommandLoop(
	vscrn *ulanode.VirtualScreen,
	nodeId int,
	reqChan chan ulanode.LocalCommandReq,
	respChan chan ulanode.LocalCommandReq,
	jsonChan chan map[string]interface{},
	retChansMap map[int]interface{},
	plugin ulanode.LocalCommandGenerator,
) {

	vs2rdConv, err := vs2rd.NewVscreen2RdisplayConverter(vscrn, nodeId)
	if err != nil {
		return
	}

	var vsconv ulanode.GeometoryConverter = vs2rdConv
	err = vsconv.DoConvert()
	if err != nil {
		return
	}

	spscrns, err := vsconv.GetNodePixelScreens()
	if err != nil {
		return
	}

	for {
		var mJson map[string]interface{}
		select {
		case mJson = <-jsonChan:
			break
		}

		ret := 0
		listenerId := mJson["listener_id"].(int)

		vscrnNew := vscrn.Dup()

		acdata, err := vscrnNew.ApplyCommand(mJson)
		if err != nil {
			ret = -1
			commResponseResult(ret, listenerId, retChansMap)
			continue
		}

		if acdata.Command == "get_vlayer" {
			commResponseData(acdata.Vlayers, listenerId, retChansMap)
			continue
		}

		vs2rdConv, err := vs2rd.NewVscreen2RdisplayConverter(vscrnNew, nodeId)
		if err != nil {
			ret = -1
			commResponseResult(ret, listenerId, retChansMap)
			continue
		}

		var vsconv ulanode.GeometoryConverter = vs2rdConv
		err = vsconv.DoConvert()
		if err != nil {
			ret = -1
			commResponseResult(ret, listenerId, retChansMap)
			continue
		}

		spscrnsNew, err := vsconv.GetNodePixelScreens()
		if err != nil {
			ret = -1
			commResponseResult(ret, listenerId, retChansMap)
			continue
		}

		reqs, err := plugin.GenerateLocalCommandReq(acdata, spscrnsNew, spscrns)
		if err != nil {
			ret = -1
			commResponseResult(ret, listenerId, retChansMap)
			continue
		}

		ret = submitCommand(reqs, reqChan, respChan)

		vscrn = vscrnNew
		spscrns = spscrnsNew

		commResponseResult(ret, listenerId, retChansMap)
	}
}

func submitCommand(
	reqs []*ulanode.LocalCommandReq,
	reqChan chan ulanode.LocalCommandReq,
	respChan chan ulanode.LocalCommandReq,
) int {

	ret := 0
	for _, req := range reqs {
		reqChan <- *req
		select {
		case lcr := <-respChan:
			ret = lcr.Ret
			break
		}
	}
	return ret
}

func commResponseResult(
	result int,
	listenerId int,
	retChansMap map[int]interface{},
) {
	Mutex.Lock()
	respChan := retChansMap[listenerId].(chan map[string]interface{})
	Mutex.Unlock()
	retJson := map[string]interface{}{
		"type":   "result",
		"result": result,
	}
	respChan <- retJson
}

func commResponseData(
	vlayers []ula.VirtualLayer,
	listenerId int,
	retChansMap map[int]interface{},
) {
	Mutex.Lock()
	respChan := retChansMap[listenerId].(chan map[string]interface{})
	Mutex.Unlock()
	retJson := map[string]interface{}{
		"type": "data",
		"data": vlayers,
	}
	respChan <- retJson
}

func mainLoop(
	listener net.Listener,
	vscrn *ulanode.VirtualScreen,
	nodeId int,
	reqChan chan ulanode.LocalCommandReq,
	respChan chan ulanode.LocalCommandReq,
	plugin ulanode.LocalCommandGenerator) {

	jsonChan := make(chan map[string]interface{}, 1)
	retChansMap := make(map[int]interface{})
	go processCommandLoop(vscrn, nodeId, reqChan, respChan, jsonChan, retChansMap, plugin)

	listenerId := 0
	for {
		conn, err := listener.Accept()
		if err != nil {
			ELog.Printf("Accept error: %s", err)
			continue
		}
		retChan := make(chan map[string]interface{}, 1)
		Mutex.Lock()
		retChansMap[listenerId] = retChan
		Mutex.Unlock()
		go readCommandLoop(conn, jsonChan, listenerId, retChansMap)
		listenerId += 1
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "%s [option] | listenIp listenPort nodeId\n", os.Args[0])

	fmt.Fprintf(os.Stderr, "[option]\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = printUsage

	var (
		verbose      bool
		debug        bool
		vScrnDefFile string
		keyNodeId    int
		keyHostName  string
	)

	flag.BoolVar(&verbose, "v", true, "verbose info log")
	flag.BoolVar(&debug, "d", true, "verbose debug log")
	flag.StringVar(&vScrnDefFile, "f", "", "virtual-screen-def.json file Path")
	flag.IntVar(&keyNodeId, "N", -1, "search ula-node param by node_id from VScrnDef file")
	flag.StringVar(&keyHostName, "H", "", "search ula-node param by hostname from VScrnDef file")

	flag.Parse()

	if verbose == true {
		ILog.SetOutput(os.Stderr)
	}

	if debug == true {
		DLog.SetOutput(os.Stderr)
	}

	DLog.Printf("ARG0:%s, ARG1:%s, ARG2:%s", flag.Arg(0), flag.Arg(1), flag.Arg(2))

	vscrnDef, err := ula.ReadVScrnDef(vScrnDefFile)
	if err != nil {
		ELog.Println("ReadVScrnDef error : ", err)
		return
	}

	var (
		listenIp   string
		listenPort int
		nodeId     int
	)

	if len(flag.Args()) == 0 {

		if keyNodeId != -1 && keyHostName != "" {
			ELog.Println("error multiple search keys")
			return
		}

		if keyHostName == "" {
			keyHostName, err = os.Hostname()
			if err != nil {
				ELog.Println("os.Hostname error : ", err)
				return
			}
		}

		if keyNodeId != -1 {
			nodeId = keyNodeId
		} else {
			nodeId, err = getNodeIdByHostName(keyHostName, vscrnDef)
			if err != nil {
				ELog.Println("getNodeIdByHostName error : ", err)
				return
			}
		}

		ipAddrs, err := getIpv4AddrsOfAllInterfaces()
		if err != nil {
			ELog.Println("getIpv4AddrsOfAllInterfaces error : ", err)
			return
		}
		listenIp, err = getIpAddrByNodeIdAndIpCandidateList(ipAddrs, nodeId, vscrnDef)
		if err != nil {
			ELog.Println("getIpAddrByNodeIdAndIpCandidateList error : ", err)
			return
		}

		listenPort, err = getPort(nodeId, vscrnDef)
		if err != nil {
			ELog.Println("getPort error : ", err)
			return
		}
	} else if len(flag.Args()) == 3 {
		listenIp = flag.Arg(0)
		listenPort, _ = strconv.Atoi(flag.Arg(1))
		nodeId, _ = strconv.Atoi(flag.Arg(2))
	} else {
		printUsage()
		return
	}

	prefix := "ula-node-" + strconv.Itoa(nodeId)
	SetLogPrefix(prefix)
	DLog.Println(listenIp, ":", listenPort)

	listenAddr := listenIp + ":" + strconv.Itoa(listenPort)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		ELog.Printf("Listen error: %s", err)
		return
	}

	defer listener.Close()

	vscrn, err := ulanode.NewVirtualScreen(vscrnDef)
	if err != nil {
		ELog.Println("Could not generate initial screen.", err)
		return
	}

	reqChan := make(chan ulanode.LocalCommandReq, 5)
	respChan := make(chan ulanode.LocalCommandReq, 5)

	var plugin ulanode.LocalCommandGenerator
	if rvgpuwinmgr.IsRvgpuCompositor(vscrnDef, nodeId) {
		plugin = rvgpuwinmgr.RvgpuPlugin{}
	} else {
		plugin = iviwinmgr.IviPlugin{}
	}
	go plugin.Start(reqChan, respChan)

	mainLoop(listener, vscrn, nodeId, reqChan, respChan, plugin)
}
