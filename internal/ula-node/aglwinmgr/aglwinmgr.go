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

package aglwinmgr

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"reflect"
	"time"
	"ula-tools/internal/ula-node"
	. "ula-tools/internal/ulog"
)

func retryConnectTarget(sockChan chan net.Conn, stopChan chan struct{}) {
	for {
		conn, err := net.Dial("unix", UHMI_AGL_WM_SOCK)
		if err == nil {
			sockChan <- conn
			break
		}
		time.Sleep(10 * time.Millisecond)

		select {
		case <-stopChan:
			return
		default:
		}
	}
}

func connectTarget() net.Conn {
	var conn net.Conn

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	sockChan := make(chan net.Conn, 1)
	stopChan := make(chan struct{})

	go retryConnectTarget(sockChan, stopChan)

	select {
	case <-ctx.Done():
		ELog.Println("Dial cannot connect to uhmi-agl-wm")
		close(stopChan)
		return nil
	case conn = <-sockChan:
		ILog.Println("Dial connected to uhmi-agl-wm")
	}

	DLog.Printf("connect OK\n")
	return conn
}

func Start(reqChan chan ulanode.LocalCommandReq, respChan chan ulanode.LocalCommandReq) {

	conn := connectTarget()
	if conn != nil {
		defer conn.Close()
	} else {
		os.Exit(1)
	}
	for {
		select {
		case wVDsp := <-reqChan:
			//fmt.Println ("[aglwinmgr reqChan]", wVDsp)
			ret := sendUhmiAglWmJson(conn, wVDsp)
			lcr := ulanode.LocalCommandReq{}
			lcr.Ret = ret
			respChan <- lcr
			break
		}
	}

}

func genInitialScreenProtocolJson(req ulanode.LocalCommandReq) (string, error) {
	var aglRDisp []AglRDisplay

	for _, rdcomm := range req.RDComms {

		var agllayers []AglLayerJson

		for _, player := range rdcomm.Players {

			var aglsurfs []AglSurfaceJson

			for _, psurf := range player.Psurfaces {

				aglSurf := AglSurfaceJson{
					Id:         psurf.VID,
					AppId:      psurf.AppID,
					Width:      psurf.PixelW,
					Height:     psurf.PixelH,
					SrcX:       psurf.PsrcX,
					SrcY:       psurf.PsrcY,
					SrcW:       psurf.PsrcW,
					SrcH:       psurf.PsrcH,
					DstX:       psurf.PdstX,
					DstY:       psurf.PdstY,
					DstW:       psurf.PdstW,
					DstH:       psurf.PdstH,
					Opacity:    OPACITY,
					Visibility: psurf.Visibility,
				}

				aglsurfs = append(aglsurfs, aglSurf)
			}

			aglLayer := AglLayerJson{
				Id:         player.VID,
				Width:      player.PixelW,
				Height:     player.PixelH,
				SrcX:       player.PsrcX,
				SrcY:       player.PsrcY,
				SrcW:       player.PsrcW,
				SrcH:       player.PsrcH,
				DstX:       player.PdstX,
				DstY:       player.PdstY,
				DstW:       player.PdstW,
				DstH:       player.PdstH,
				Opacity:    OPACITY,
				Visibility: player.Visibility,
				Surface:    aglsurfs,
			}

			agllayers = append(agllayers, aglLayer)
		}

		rdisp := AglRDisplay{
			RDisplayId: rdcomm.Rdisplay.RDisplayId,
			Layers:     agllayers,
		}

		aglRDisp = append(aglRDisp, rdisp)
	}

	aglProto := InitialScreenProtocol{
		Version:  VERSION,
		Command:  "initial_screen",
		RDisplay: aglRDisp,
	}

	jsonBytes, err := json.Marshal(aglProto)
	if err != nil {
		ELog.Println("JSON Marshal error:", err)
	}

	msg := string(jsonBytes)

	return msg, nil
}

func sendMagicCode(conn net.Conn) error {

	n, err := conn.Write(MAGIC_CODE)
	if err != nil || n == 0 {
		return errors.New(fmt.Sprintf("Write error: %s \n", err))
	}

	buf := make([]byte, 4)
	n, err = conn.Read(buf)
	if err != nil || n == 0 {
		return errors.New(fmt.Sprintf("Read error: %s \n", err))
	}
	if reflect.DeepEqual(buf, MAGIC_CODE) == false {
		return errors.New(fmt.Sprintf("Read magic code false: %s\n", buf))
	}
	DLog.Printf("Read magic code: %s", buf)

	return nil
}

func sendJsonMsg(conn net.Conn, msg string) error {

	msgLen := uint32(len(msg))
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, msgLen)

	DLog.Println("Write JSON size:", msgLen)
	n, err := conn.Write([]byte(size))
	if err != nil || n == 0 {
		return errors.New(fmt.Sprintf("Write error: %s \n", err))
	}

	DLog.Println("Write JSON msg:", msg)
	n, err = conn.Write([]byte(msg))
	if err != nil || n == 0 {
		return errors.New(fmt.Sprintf("Write error: %s \n", err))
	}

	buf := make([]byte, 4)
	n, err = conn.Read(buf)
	if err != nil || n == 0 {
		return errors.New(fmt.Sprintf("Read error: %s \n", err))
	}
	ret := binary.BigEndian.Uint32(buf)
	DLog.Printf("Read ret: %x", ret)
	return nil
}

func sendUhmiAglWmJson(conn net.Conn, req ulanode.LocalCommandReq) int {

	DLog.Println("sendUhmiAglWmJson's Command:", req.Command)

	msg := ""
	var err error

	switch req.Command {
	case "initial_vscreen":
		msg, err = genInitialScreenProtocolJson(req)
	case "local_comm":
		return 0
	default:
		fmt.Println("Error req.Command")
		return -1
	}
	if err != nil {
		fmt.Println("Error ProtocolJson")
		return -1
	}

	fmt.Println("[aglwinmgr reqChan]", req)
	if conn == nil {
		return -1
	}

	err = sendMagicCode(conn)
	if err != nil {
		fmt.Println("Error SendRecv MagicCode")
		return -1
	}

	err = sendJsonMsg(conn, msg)
	if err != nil {
		fmt.Println("Error SendRecv JsonMsg")
		return -1
	}

	return 0
}
