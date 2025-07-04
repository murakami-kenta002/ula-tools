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

package ula

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

type VScrnDef struct {
	Def2D struct {
		Size struct {
			VirtualW int `json:"virtual_w"`
			VirtualH int `json:"virtual_h"`
		} `json:"size"`

		VirtualDisplays []struct {
			DispName   string `json:"disp_name"`
			VDisplayId int    `json:"vdisplay_id"`
			VirtualX   int    `json:"virtual_x"`
			VirtualY   int    `json:"virtual_y"`
			VirtualW   int    `json:"virtual_w"`
			VirtualH   int    `json:"virtual_h"`
		} `json:"virtual_displays"`
	} `json:"virtual_screen_2d"`

	RealDisplays []struct {
		NodeId     int `json:"node_id"`
		VDisplayId int `json:"vdisplay_id"`
		PixelW     int `json:"pixel_w"`
		PixelH     int `json:"pixel_h"`
		RDisplayId int `json:"rdisplay_id"`
	} `json:"real_displays"`

	Nodes []struct {
		NodeId   int    `json:"node_id"`
		HostName string `json:"hostname"`
		Ip       string `json:"ip"`
	} `json:"node"`

	DistributedWindowSystem struct {
		FrameworkNode []struct {
			NodeId int `json:"node_id"`
			Ula    struct {
				Debug     bool `json:"debug"`
				DebugPort int  `json:"debug_port"`
				Port      int  `json:"port"`
			} `json:"ula"`
			Compositor []struct {
				VDisplayIds    []int  `json:"vdisplay_ids"`
				SockDomainName string `json:"sock_domain_name"`
			} `json:"compositor"`
		} `json:"framework_node"`
	} `json:"distributed_window_system"`

	VirtualSafetyArea []struct {
		VirtualX int `json:"virtual_x"`
		VirtualY int `json:"virtual_y"`
		VirtualW int `json:"virtual_w"`
		VirtualH int `json:"virtual_h"`
	} `json:"virtual_safety_area"`
}

func ReadVScrnDef(vsdPath ...string) (*VScrnDef, error) {
	var fname string
	if len(vsdPath) > 0 && vsdPath[0] != "" {
		fname = vsdPath[0]
	} else {
		fname = GetEnvString("VSDPATH", VSCRNDEF_FILE)
	}
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}

	jsonBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var vscrnDef VScrnDef
	err = json.Unmarshal(jsonBytes, &vscrnDef)
	if err != nil {
		return nil, errors.New("json decode error")
	}

	return &vscrnDef, nil
}

func (vdef *VScrnDef) IsVDisplayInNode(nodeId int, vDisplayId int) bool {

	for _, rdisplay := range vdef.RealDisplays {

		if rdisplay.VDisplayId == vDisplayId && rdisplay.NodeId == nodeId {
			return true
		}

	}

	return false
}

func (vdef *VScrnDef) GetVDisplays() []VirtualDisplay {
	vDisplays := make([]VirtualDisplay, 0)
	for _, vdisplay := range vdef.Def2D.VirtualDisplays {
		var tmp VirtualDisplay
		tmp.DispName = vdisplay.DispName
		tmp.VDisplayId = vdisplay.VDisplayId
		tmp.VirtualX = vdisplay.VirtualX
		tmp.VirtualY = vdisplay.VirtualY
		tmp.VirtualW = vdisplay.VirtualW
		tmp.VirtualH = vdisplay.VirtualH
		vDisplays = append(vDisplays, tmp)
	}
	return vDisplays
}
