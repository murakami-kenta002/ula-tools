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

package vs2rd

import (
	"errors"
	"ula-tools/internal/ula"
	"ula-tools/internal/ula-node"
)

type WorkV2R struct {
	vdisplay ula.VirtualDisplay
	rdisplay ulanode.RealDisplay
	vlayers  []ula.VirtualLayer

	/* Geometry is Coordinated by RDisplay  */
	players []ulanode.PixelLayer
}

type Vscreen2RdisplayConverter struct {
	nodeId     int
	workV2RMap map[int]WorkV2R
}

func NewVscreen2RdisplayConverter(
	vscrn *ulanode.VirtualScreen,
	nodeId int) (*Vscreen2RdisplayConverter, error) {

	wdisplay, err := generateWorkV2R(vscrn, nodeId, &vscrn.VScrnDef)
	if err != nil {
		return nil, errors.New("NewVscreen2RdisplayConverter error")
	}

	vs2rd := &Vscreen2RdisplayConverter{
		nodeId:     nodeId,
		workV2RMap: wdisplay,
	}

	return vs2rd, nil
}

func (vs2rd *Vscreen2RdisplayConverter) DoConvert() error {

	convVDisplay2RDisplayCoordinate(vs2rd.workV2RMap)

	return nil
}

/* util function */
func convVSurface2PSurface(vsurf *ula.VirtualSurface) *ulanode.PixelSurface {

	psurf := ulanode.NewEmptyPixelSurface()

	psurf.ParentVID = vsurf.ParentVID
	psurf.VID = vsurf.VID

	psurf.PixelW = vsurf.PixelW
	psurf.PixelH = vsurf.PixelH

	psurf.PsrcX = vsurf.PsrcX
	psurf.PsrcY = vsurf.PsrcY
	psurf.PsrcW = vsurf.PsrcW
	psurf.PsrcH = vsurf.PsrcH

	psurf.PdstX = vsurf.VdstX
	psurf.PdstY = vsurf.VdstY
	psurf.PdstW = vsurf.VdstW
	psurf.PdstH = vsurf.VdstH

	psurf.Visibility = vsurf.Visibility
	return psurf
}

func convVLayer2PLayer(vlayer *ula.VirtualLayer) *ulanode.PixelLayer {

	player := ulanode.NewEmptyPixelLayer()

	player.VID = vlayer.VID

	player.PixelW = vlayer.VirtualW
	player.PixelH = vlayer.VirtualH

	player.PsrcX = vlayer.VsrcX
	player.PsrcY = vlayer.VsrcY
	player.PsrcW = vlayer.VsrcW
	player.PsrcH = vlayer.VsrcH

	player.PdstX = vlayer.VdstX
	player.PdstY = vlayer.VdstY
	player.PdstW = vlayer.VdstW
	player.PdstH = vlayer.VdstH

	player.Visibility = vlayer.Visibility

	player.Psurfaces = make([]ulanode.PixelSurface, 0)

	for _, sVsurf := range vlayer.Vsurfaces {
		copiedPSurf := convVSurface2PSurface(&sVsurf)
		player.Psurfaces = append(player.Psurfaces, *copiedPSurf)
	}

	return player
}

func convVLayers2PLayers(sVlayers []ula.VirtualLayer) (dPlayers []ulanode.PixelLayer) {
	if sVlayers == nil {
		return nil
	}

	dPlayers = make([]ulanode.PixelLayer, 0)

	for _, sVlayer := range sVlayers {
		dPlayer := convVLayer2PLayer(&sVlayer)
		dPlayers = append(dPlayers, *dPlayer)
	}

	return dPlayers
}

func isNeedForWorkV2R(vlayer *ula.VirtualLayer, arg interface{}) bool {
	vDisplayId := arg.(int)

	if vlayer.Coord == ula.COORD_GLOBAL {
		return true
	}

	if vlayer.Coord == ula.COORD_VDISPLAY &&
		vlayer.VDisplayId == vDisplayId {
		return true
	}

	return false
}

func generateWorkV2R(
	vscreen *ulanode.VirtualScreen,
	nodeId int,
	vscrnDef *ula.VScrnDef) (map[int]WorkV2R, error) {

	workV2RMap := make(map[int]WorkV2R)

	for vdspid, vdisplay := range vscreen.VirtualDisplays {

		isMe := vscrnDef.IsVDisplayInNode(nodeId, vdspid)
		if !isMe {
			continue
		}

		rdisplay := vscreen.RealDisplays[vdspid]

		wvdisp := WorkV2R{
			vdisplay: *vdisplay.Dup(),
			rdisplay: *rdisplay.Dup(),
			vlayers:  ulanode.DupVirtualLayerSliceIfNeed(vscreen.VdispVlayers[vdspid], isNeedForWorkV2R, vdspid),
		}

		workV2RMap[vdspid] = wvdisp
	}

	return workV2RMap, nil
}

func convGlobalToVDisplayCoordinateSub(
	vdisp_vx int,
	vdisp_vw int,
	vlayer_vdx int,
	vlayer_vdw int,
	vlayer_vsx int,
	vlayer_vsw int) (int, int, int, int) {

	var (
		nvlayer_vdx int
		nvlayer_vdw int
		nvlayer_vsx int
		nvlayer_vsw int
	)

	/* Assume src cutout area is the entire layer in this part */
	if vdisp_vx <= vlayer_vdx &&
		vlayer_vdx <= vdisp_vx+vdisp_vw &&
		vdisp_vx+vdisp_vw <= vlayer_vdx+vlayer_vdw {

		nvlayer_vdx = vlayer_vdx - vdisp_vx
		nvlayer_vsx = 0
		nvlayer_vdw = vdisp_vx + vdisp_vw - vlayer_vdx
		nvlayer_vsw = nvlayer_vdw

	} else if vlayer_vdx <= vdisp_vx &&
		vdisp_vx+vdisp_vw <= vlayer_vdx+vlayer_vdw {

		nvlayer_vdx = 0
		nvlayer_vsx = vdisp_vx - vlayer_vdx
		nvlayer_vdw = vdisp_vw
		nvlayer_vsw = nvlayer_vdw

	} else if vlayer_vdx <= vdisp_vx &&
		vdisp_vx <= vlayer_vdx+vlayer_vdw &&
		vlayer_vdx+vlayer_vdw <= vdisp_vx+vdisp_vw {

		nvlayer_vdx = 0
		nvlayer_vsx = vdisp_vx - vlayer_vdx
		nvlayer_vdw = vlayer_vdx + vlayer_vdw - vdisp_vx
		nvlayer_vsw = nvlayer_vdw

	} else if vdisp_vx <= vlayer_vdx &&
		vlayer_vdx+vlayer_vdw <= vdisp_vx+vdisp_vw {

		nvlayer_vdx = vlayer_vdx - vdisp_vx
		nvlayer_vsx = 0
		nvlayer_vdw = vlayer_vdw
		nvlayer_vsw = nvlayer_vdw

	} else {
		nvlayer_vdx = 0
		nvlayer_vsx = 0
		nvlayer_vdw = 0
		nvlayer_vsw = 0
	}

	/* Corresponds to src cutout position */
	nvlayer_vsx = nvlayer_vsx*vlayer_vsw/vlayer_vdw + vlayer_vsx
	nvlayer_vsw = nvlayer_vsw * vlayer_vsw / vlayer_vdw

	return nvlayer_vdx, nvlayer_vdw, nvlayer_vsx, nvlayer_vsw

}

func convGlobalToVDisplayCoordinate(sVlayer *ula.VirtualLayer,
	vdisp *ula.VirtualDisplay) *ula.VirtualLayer {

	vdisp_vx := vdisp.VirtualX
	vdisp_vy := vdisp.VirtualY
	vdisp_vw := vdisp.VirtualW
	vdisp_vh := vdisp.VirtualH

	newVlayer := sVlayer.Dup()

	vlayer_vdx := sVlayer.VdstX
	vlayer_vdy := sVlayer.VdstY
	vlayer_vdw := sVlayer.VdstW
	vlayer_vdh := sVlayer.VdstH

	vlayer_vsx := sVlayer.VsrcX
	vlayer_vsy := sVlayer.VsrcY
	vlayer_vsw := sVlayer.VsrcW
	vlayer_vsh := sVlayer.VsrcH

	nvlayer_vdx, nvlayer_vdw, nvlayer_vsx, nvlayer_vsw :=
		convGlobalToVDisplayCoordinateSub(vdisp_vx, vdisp_vw, vlayer_vdx, vlayer_vdw, vlayer_vsx, vlayer_vsw)
	nvlayer_vdy, nvlayer_vdh, nvlayer_vsy, nvlayer_vsh :=
		convGlobalToVDisplayCoordinateSub(vdisp_vy, vdisp_vh, vlayer_vdy, vlayer_vdh, vlayer_vsy, vlayer_vsh)

	newVlayer.VdstX = nvlayer_vdx
	newVlayer.VdstW = nvlayer_vdw
	newVlayer.VsrcX = nvlayer_vsx
	newVlayer.VsrcW = nvlayer_vsw

	newVlayer.VdstY = nvlayer_vdy
	newVlayer.VsrcY = nvlayer_vsy
	newVlayer.VdstH = nvlayer_vdh
	newVlayer.VsrcH = nvlayer_vsh

	return newVlayer
}

func convToVDisplayCoordinate(sVlayers []ula.VirtualLayer,
	vdisp *ula.VirtualDisplay) []ula.VirtualLayer {

	dVlayers := make([]ula.VirtualLayer, 0)

	for _, vlayer := range sVlayers {

		var newVlayer ula.VirtualLayer
		if vlayer.Coord == ula.COORD_VDISPLAY {
			newVlayer = *vlayer.Dup()
		} else {
			newVlayer = *convGlobalToVDisplayCoordinate(&vlayer, vdisp)
		}

		dVlayers = append(dVlayers, newVlayer)
	}

	return dVlayers
}

func convToRDisplayCoordinate(sVlayers []ula.VirtualLayer,
	vdisp *ula.VirtualDisplay, rdisp *ulanode.RealDisplay) []ula.VirtualLayer {

	dVlayers := make([]ula.VirtualLayer, 0)

	vdisp_vw := vdisp.VirtualW
	vdisp_vh := vdisp.VirtualH

	rdisp_pixw := rdisp.PixelW
	rdisp_pixh := rdisp.PixelH

	for _, vlayer := range sVlayers {

		newVlayer := vlayer.Dup()

		vlayer_vdx := vlayer.VdstX
		vlayer_vdy := vlayer.VdstY
		vlayer_vdw := vlayer.VdstW
		vlayer_vdh := vlayer.VdstH

		/* Convert RealDisplay Geometory */
		newVlayer.VdstX = vlayer_vdx * rdisp_pixw / vdisp_vw
		newVlayer.VdstW = vlayer_vdw * rdisp_pixw / vdisp_vw
		newVlayer.VdstY = vlayer_vdy * rdisp_pixh / vdisp_vh
		newVlayer.VdstH = vlayer_vdh * rdisp_pixh / vdisp_vh

		dVlayers = append(dVlayers, *newVlayer)
	}

	return dVlayers
}

func convVDisplay2RDisplayCoordinate(workV2RMap map[int]WorkV2R) {

	for key, workV2R := range workV2RMap {
		tmpVlayers := convToVDisplayCoordinate(workV2R.vlayers, &workV2R.vdisplay)
		tmpVlayers2 := convToRDisplayCoordinate(tmpVlayers, &workV2R.vdisplay, &workV2R.rdisplay)
		tmpPlayers := convVLayers2PLayers(tmpVlayers2)
		workV2R.players = tmpPlayers
		workV2RMap[key] = workV2R
	}
}

func (vs2rd *Vscreen2RdisplayConverter) GetNodePixelScreens() (*ulanode.NodePixelScreens, error) {

	pscrns := make([]ulanode.PixelScreen, 0)

	for _, workV2R := range vs2rd.workV2RMap {

		pscrn, err := ulanode.NewPixelScreen(&workV2R.rdisplay, workV2R.players)
		if err != nil {
			return nil, err
		}
		pscrns = append(pscrns, *pscrn)
	}

	spscrns, err := ulanode.NewNodePixelScreens(vs2rd.nodeId, pscrns)
	if err != nil {
		return nil, err
	}

	return spscrns, nil
}
