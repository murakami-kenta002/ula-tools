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
	"errors"
	"ula-tools/internal/ula"
	. "ula-tools/internal/ulog"
)

func getStringFromJson(mJson map[string]interface{}, key string) (string, error) {

	tval := mJson[key]
	if tval == nil {
		return "", errors.New("Error in getStringFromJson")
	}
	val := tval.(string)

	return val, nil
}

func getIntFromJson(mJson map[string]interface{}, key string) (int, error) {

	tval := mJson[key]
	if tval == nil {
		return 0, errors.New("Error in getIntFromJson")
	}
	val := int(tval.(float64))

	return val, nil
}

func getIntFromJsonDef(mJson map[string]interface{}, key string, defval int) (int, error) {

	tval := mJson[key]
	if tval == nil {
		return defval, nil
	}
	val := int(tval.(float64))

	return val, nil
}

func getSliceFromJson(mJson map[string]interface{}, key string) ([]interface{}, error) {

	tval := mJson[key]
	if tval == nil {
		return make([]interface{}, 0), errors.New("Error in getSliceFromJson")
	}
	val := tval.([]interface{})

	return val, nil
}

func getCoordFromJson(mJson map[string]interface{}, key string) (ula.Coord, error) {

	tval := mJson[key]
	if tval == nil {
		return 0, errors.New("getCoordFromJson")
	}

	defval := ula.COORD_GLOBAL
	if tval == "global" {
		defval = ula.COORD_GLOBAL
	} else if tval == "vdisplay" {
		defval = ula.COORD_VDISPLAY
	} else {
		return 0, errors.New("getCoordFromJson")
	}

	return defval, nil
}

func getCoordFromJsonDef(mJson map[string]interface{}, key string, defval ula.Coord) (ula.Coord, error) {

	tval := mJson[key]
	if tval == nil {
		return defval, nil
	}

	if tval == "global" {
		defval = ula.COORD_GLOBAL
	} else if tval == "vdisplay" {
		defval = ula.COORD_VDISPLAY
	} else {
		return 0, errors.New("getCoordFromJsonDef")
	}

	return defval, nil
}

func generateSurfaceFromParam(layerId int, mSurface map[string]interface{}) (*ula.VirtualSurface, error) {

	surfaceId, err := getIntFromJson(mSurface, "VID")
	if err != nil {
		return nil, err
	}

	appId, err := getStringFromJson(mSurface, "APPID")
	if err != nil {
		return nil, err
	}

	pixelW, err := getIntFromJson(mSurface, "pixel_w")
	if err != nil {
		return nil, err
	}

	pixelH, err := getIntFromJson(mSurface, "pixel_h")
	if err != nil {
		return nil, err
	}

	psrcX, err := getIntFromJson(mSurface, "psrc_x")
	if err != nil {
		return nil, err
	}

	psrcY, err := getIntFromJson(mSurface, "psrc_y")
	if err != nil {
		return nil, err
	}

	psrcW, err := getIntFromJson(mSurface, "psrc_w")
	if err != nil {
		return nil, err
	}

	psrcH, err := getIntFromJson(mSurface, "psrc_h")
	if err != nil {
		return nil, err
	}

	vdstX, err := getIntFromJson(mSurface, "vdst_x")
	if err != nil {
		return nil, err
	}

	vdstY, err := getIntFromJson(mSurface, "vdst_y")
	if err != nil {
		return nil, err
	}

	vdstW, err := getIntFromJson(mSurface, "vdst_w")
	if err != nil {
		return nil, err
	}

	vdstH, err := getIntFromJson(mSurface, "vdst_h")
	if err != nil {
		return nil, err
	}

	if pixelW < 0 || pixelH < 0 {
		return nil, errors.New("pixel_w and pixel_h should not have negative values")
	}

	if psrcX < 0 || psrcY < 0 || psrcW < 0 || psrcH < 0 {
		return nil, errors.New("psrc regions should not have negative values")
	}

	if vdstX < 0 || vdstY < 0 || vdstW < 0 || vdstH < 0 {
		return nil, errors.New("vdst regions should not have negative values")
	}

	visibility, err := getIntFromJsonDef(mSurface, "visibility", 1)
	if err != nil {
		return nil, err
	}

	vsurface := ula.VirtualSurface{
		ParentVID:  layerId,
		VID:        surfaceId,
		AppID:      appId,
		PixelW:     pixelW,
		PixelH:     pixelH,
		PsrcX:      psrcX,
		PsrcY:      psrcY,
		PsrcW:      psrcW,
		PsrcH:      psrcH,
		VdstX:      vdstX,
		VdstY:      vdstY,
		VdstW:      vdstW,
		VdstH:      vdstH,
		Visibility: visibility,
	}

	return &vsurface, nil
}

func generateLayerFromParam(mLayer map[string]interface{}, genSurfaces bool, existingVlayer *ula.VirtualLayer) (*ula.VirtualLayer, error) {

	layerId, err := getIntFromJson(mLayer, "VID")
	if err != nil {
		if existingVlayer == nil {
			return nil, err
		} else {
			layerId = existingVlayer.VID
		}
	}

	var defCoord ula.Coord
	if existingVlayer != nil {
		defCoord = existingVlayer.Coord
	} else {
		defCoord = ula.COORD_GLOBAL
	}
	coord, err := getCoordFromJsonDef(mLayer, "coord", defCoord)
	if err != nil {
		ELog.Println("error in generateLayerFromParam")
		return nil, err
	}

	vdisplayId := -1
	if coord == ula.COORD_VDISPLAY {
		defVdisplayId := -1
		if existingVlayer != nil {
			defVdisplayId = existingVlayer.VDisplayId
		}
		vdisplayId, err = getIntFromJsonDef(mLayer, "vdisplay_id", defVdisplayId)
		if err != nil {
			ELog.Println("error in generateLayerFromParam")
			return nil, err
		}
	}

	virtualW, err := getIntFromJson(mLayer, "virtual_w")
	if err != nil {
		if existingVlayer == nil {
			ELog.Println("error in generateLayerFromParam")
			return nil, err
		} else {
			virtualW = existingVlayer.VirtualW
		}
	}

	virtualH, err := getIntFromJson(mLayer, "virtual_h")
	if err != nil {
		if existingVlayer == nil {
			ELog.Println("error in generateLayerFromParam")
			return nil, err
		} else {
			virtualH = existingVlayer.VirtualH
		}
	}

	vsrcX, err := getIntFromJson(mLayer, "vsrc_x")
	if err != nil {
		if existingVlayer == nil {
			ELog.Println("error in generateLayerFromParam")
			return nil, err
		} else {
			vsrcX = existingVlayer.VsrcX
		}
	}

	vsrcY, err := getIntFromJson(mLayer, "vsrc_y")
	if err != nil {
		if existingVlayer == nil {
			ELog.Println("error in generateLayerFromParam")
			return nil, err
		} else {
			vsrcY = existingVlayer.VsrcY
		}
	}

	vsrcW, err := getIntFromJson(mLayer, "vsrc_w")
	if err != nil {
		if existingVlayer == nil {
			ELog.Println("error in generateLayerFromParam")
			return nil, err
		} else {
			vsrcW = existingVlayer.VsrcW
		}
	}

	vsrcH, err := getIntFromJson(mLayer, "vsrc_h")
	if err != nil {
		if existingVlayer == nil {
			ELog.Println("error in generateLayerFromParam")
			return nil, err
		} else {
			vsrcH = existingVlayer.VsrcH
		}
	}

	vdstX, err := getIntFromJson(mLayer, "vdst_x")
	if err != nil {
		if existingVlayer == nil {
			ELog.Println("error in generateLayerFromParam")
			return nil, err
		} else {
			vdstX = existingVlayer.VdstX
		}
	}

	vdstY, err := getIntFromJson(mLayer, "vdst_y")
	if err != nil {
		if existingVlayer == nil {
			ELog.Println("error in generateLayerFromParam")
			return nil, err
		} else {
			vdstY = existingVlayer.VdstY
		}
	}

	vdstW, err := getIntFromJson(mLayer, "vdst_w")
	if err != nil {
		if existingVlayer == nil {
			ELog.Println("error in generateLayerFromParam")
			return nil, err
		} else {
			vdstW = existingVlayer.VdstW
		}
	}

	vdstH, err := getIntFromJson(mLayer, "vdst_h")
	if err != nil {
		if existingVlayer == nil {
			ELog.Println("error in generateLayerFromParam")
			return nil, err
		} else {
			vdstH = existingVlayer.VdstH
		}
	}

	if virtualW < 0 || virtualH < 0 {
		return nil, errors.New("virtual_w and virtual_h should not have negative values")
	}

	if vsrcX < 0 || vsrcY < 0 || vsrcW < 0 || vsrcH < 0 {
		return nil, errors.New("vsrc regions should not have negative values")
	}

	if vdstX < 0 || vdstY < 0 || vdstW < 0 || vdstH < 0 {
		return nil, errors.New("vdst regions should not have negative values")
	}

	if vdstW == 0 || vdstH == 0 {
		return nil, errors.New("vdst_w and vdst_h need to have values other than 0")
	}

	defVisibility := 1
	if existingVlayer != nil {
		defVisibility = existingVlayer.Visibility
	}
	visibility, err := getIntFromJsonDef(mLayer, "visibility", defVisibility)
	if err != nil {
		ELog.Println("error in generateLayerFromParam")
		return nil, err
	}

	vsurfaces := make([]ula.VirtualSurface, 0)
	if genSurfaces {
		surfaces, err := getSliceFromJson(mLayer, "vsurface")
		if err != nil {
			ELog.Println("error in generateLayerFromParam")
			return nil, err
		}

		for _, mSurface := range surfaces {
			newVsurface, err := generateSurfaceFromParam(layerId, mSurface.(map[string]interface{}))
			if err != nil {
				ELog.Println("error in generateLayerFromParam")
				return nil, err
			}
			vsurfaces = append(vsurfaces, *newVsurface)
		}
	}

	vlayer := ula.VirtualLayer{
		VID:        layerId,
		Coord:      coord,
		VDisplayId: vdisplayId,
		VirtualW:   virtualW,
		VirtualH:   virtualH,
		VsrcX:      vsrcX,
		VsrcY:      vsrcY,
		VsrcW:      vsrcW,
		VsrcH:      vsrcH,
		VdstX:      vdstX,
		VdstY:      vdstY,
		VdstW:      vdstW,
		VdstH:      vdstH,
		Visibility: visibility,
		Vsurfaces:  vsurfaces,
	}

	return &vlayer, nil
}

/* fill initial virtual layer, virtual window param*/
func fillVscreenFromParam(vscreen *VirtualScreen, mJson map[string]interface{}) error {

	vlayers := make([]ula.VirtualLayer, 0)

	layers, err := getSliceFromJson(mJson, "vlayer")
	if err != nil {
		ELog.Println("err in fillVscreenFromParam")
		return err
	}

	for _, mLayer := range layers {
		var existingVlayer *ula.VirtualLayer
		newVlayer, err := generateLayerFromParam(mLayer.(map[string]interface{}), true, existingVlayer)
		if err != nil {
			ELog.Println("err in fillVscreenFromParam")
			return err
		}
		vlayers = append(vlayers, *newVlayer)
	}

	for _, vdisp := range vscreen.VirtualDisplays {
		vscreen.VdispVlayers[vdisp.VDisplayId] = vlayers
	}

	return nil
}
