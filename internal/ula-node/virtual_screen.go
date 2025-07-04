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

package ulanode

import (
	"ula-tools/internal/ula"
	. "ula-tools/internal/ulog"
)

type VirtualScreen struct {
	VScrnDef ula.VScrnDef

	VirtualWidth  int
	VirtualHeight int

	VirtualDisplays   map[int]ula.VirtualDisplay
	RealDisplays      map[int]RealDisplay
	VdispVlayers      map[int][]ula.VirtualLayer
	VdispVsafetyAreas map[int][]ula.VirtualSafetyArea
}

func NewVirtualScreen(vscrnDef *ula.VScrnDef) (*VirtualScreen, error) {

	vscreen := VirtualScreen{
		VScrnDef:          *vscrnDef,
		VirtualWidth:      vscrnDef.Def2D.Size.VirtualW,
		VirtualHeight:     vscrnDef.Def2D.Size.VirtualH,
		VirtualDisplays:   make(map[int]ula.VirtualDisplay),
		RealDisplays:      make(map[int]RealDisplay),
		VdispVlayers:      make(map[int][]ula.VirtualLayer),
		VdispVsafetyAreas: make(map[int][]ula.VirtualSafetyArea),
	}

	for _, r := range vscrnDef.Def2D.VirtualDisplays {
		vscreen.VirtualDisplays[r.VDisplayId] = ula.VirtualDisplay{
			DispName:   r.DispName,
			VDisplayId: r.VDisplayId,
			VirtualX:   r.VirtualX,
			VirtualY:   r.VirtualY,
			VirtualW:   r.VirtualW,
			VirtualH:   r.VirtualH,
		}
		vscreen.VdispVlayers[r.VDisplayId] = make([]ula.VirtualLayer, 0)
	}

	for _, r := range vscrnDef.RealDisplays {
		vscreen.RealDisplays[r.VDisplayId] = RealDisplay{
			NodeId:     r.NodeId,
			VDisplayId: r.VDisplayId,
			PixelW:     r.PixelW,
			PixelH:     r.PixelH,
			RDisplayId: r.RDisplayId,
		}
	}

	vVdispVsafetyAreas := make([]ula.VirtualSafetyArea, 0)
	for _, r := range vscrnDef.VirtualSafetyArea {
		vVdispVsafetyArea := ula.VirtualSafetyArea{
			VirtualX: r.VirtualX,
			VirtualY: r.VirtualY,
			VirtualW: r.VirtualW,
			VirtualH: r.VirtualH,
		}
		vVdispVsafetyAreas = append(vVdispVsafetyAreas, vVdispVsafetyArea)
	}
	for _, r := range vscrnDef.Def2D.VirtualDisplays {
		vscreen.VdispVsafetyAreas[r.VDisplayId] = vVdispVsafetyAreas
	}

	return &vscreen, nil
}

func (vscreen *VirtualScreen) Dup() *VirtualScreen {

	copiedVDsps := make(map[int]ula.VirtualDisplay)
	copiedRDsps := make(map[int]RealDisplay)
	copiedVDispVLayers := make(map[int][]ula.VirtualLayer)
	copiedVdispVsafetyAreas := make(map[int][]ula.VirtualSafetyArea)

	for vdspid, vdisplay := range vscreen.VirtualDisplays {
		copiedVDsp := vdisplay
		copiedRDsp := vscreen.RealDisplays[vdspid]
		copiedVLayers := vscreen.VdispVlayers[vdspid]
		copiedVSafetyAreas := vscreen.VdispVsafetyAreas[vdspid]

		copiedVDsps[vdspid] = copiedVDsp
		copiedRDsps[vdspid] = copiedRDsp
		copiedVDispVLayers[vdspid] = DupVirtualLayerSlice(copiedVLayers)
		copiedVdispVsafetyAreas[vdspid] = DupVirtualSafetyAreaSlice(copiedVSafetyAreas)
	}

	copiedVscreen := *vscreen
	copiedVscreen.VirtualDisplays = copiedVDsps
	copiedVscreen.RealDisplays = copiedRDsps
	copiedVscreen.VdispVlayers = copiedVDispVLayers
	copiedVscreen.VdispVsafetyAreas = copiedVdispVsafetyAreas
	return &copiedVscreen
}

func (vscrn *VirtualScreen) ApplyCommand(mJson map[string]interface{}) (*ApplyCommandData, error) {

	command := mJson["command"].(string)
	DLog.Println("command=", command)

	var chgIds []IdPair
	var err error
	var vlayers []ula.VirtualLayer

	switch command {
	case "initial_vscreen":
		DLog.Println("@@INITIAL_VSCREEN@@")
		chgIds, err = initVirtualScreen(vscrn, mJson)
	case "get_vlayer":
		DLog.Println("@@GET_LAYER@@")
		vlayers, err = getVlayerParams(vscrn, mJson)
	default:
		chgIds = make([]IdPair, 0)
	}

	if err != nil {
		return nil, err
	}

	var vids []int
	for _, vdisp := range vscrn.VirtualDisplays {
		vids = make([]int, 0)
		for _, vlayer := range vscrn.VdispVlayers[vdisp.VDisplayId] {
			vids = append(vids, vlayer.VID)
		}
		DLog.Println("ApplyCommand command: ", command, " vdispId: ", vdisp.VDisplayId, " vids: ", vids)
	}

	acdata := &ApplyCommandData{Command: command, ChgIds: chgIds, Vlayers: vlayers}

	return acdata, nil

}

func getVlayerParams(vscreen *VirtualScreen, mJson map[string]interface{}) ([]ula.VirtualLayer, error) {
	retVlayers := make([]ula.VirtualLayer, 0)
	layers := mJson["vlayer"].([]interface{})

	for _, vdisp := range vscreen.VirtualDisplays {
		for _, vlayer := range vscreen.VdispVlayers[vdisp.VDisplayId] {
			for _, mLayer := range layers {
				layerId, err := getIntFromJson(mLayer.(map[string]interface{}), "VID")
				if err != nil {
					return retVlayers, err
				}
				if vlayer.VID == layerId {
					retVlayers = append(retVlayers, vlayer)
				}
			}
		}
	}

	return retVlayers, nil
}

func initVirtualScreen(vscreen *VirtualScreen, mJson map[string]interface{}) ([]IdPair, error) {
	err := fillVscreenFromParam(vscreen, mJson)
	if err != nil {
		return make([]IdPair, 0), err
	}

	return make([]IdPair, 0), nil
}
