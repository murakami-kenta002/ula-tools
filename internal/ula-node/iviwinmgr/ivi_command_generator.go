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

package iviwinmgr

import (
	"errors"
	"ula-tools/internal/ula-node"
)

type workIvi struct {
	rdisplay ulanode.RealDisplay
	players  []ulanode.PixelLayer
}

type IviPlugin struct{}

func (plugin IviPlugin) GenerateLocalCommandReq(acdata *ulanode.ApplyCommandData, sps *ulanode.NodePixelScreens, oldSps *ulanode.NodePixelScreens) ([]*ulanode.LocalCommandReq, error) {
	ltqs := []*ulanode.LocalCommandReq{}

	splitOldSps, err := splitLayer(oldSps.Dup())
	if err != nil {
		return nil, errors.New("splitLayer error")
	}
	oldwIviMap, err := generateWorkIvi(splitOldSps)
	if err != nil {
		return nil, errors.New("generateWorkIvi error")
	}

	splitSps, err := splitLayer(sps.Dup())
	if err != nil {
		return nil, errors.New("splitLayer error")
	}
	wIviMap, err := generateWorkIvi(splitSps)
	if err != nil {
		return nil, errors.New("generateWorkIvi error")
	}

	ret, err := pickupInitialVScreen(wIviMap, oldwIviMap)
	if err == nil && ret != nil {
		ltqs = append(ltqs, ret)
	}

	return ltqs, nil
}

func generateWorkIvi(
	spscrns *ulanode.NodePixelScreens) (map[int]workIvi, error) {

	workIviMap := make(map[int]workIvi)

	for _, pscrn := range spscrns.Pscreens {
		wvdisp := workIvi{
			rdisplay: *pscrn.Rdisplay.Dup(),
			players:  ulanode.DupPixelLayerSlice(pscrn.Players),
		}
		workIviMap[wvdisp.rdisplay.RDisplayId] = wvdisp
	}

	err := isValidWorkIviMap(workIviMap)

	return workIviMap, err
}

func isValidWorkIviMap(wIviMap map[int]workIvi) error {

	rdisplayMap := make(map[int]bool)
	for _, workIvi := range wIviMap {
		if rdisplayMap[workIvi.rdisplay.RDisplayId] {
			return errors.New("PixelScreens has duplicate RDisplayId")
		}
		rdisplayMap[workIvi.rdisplay.RDisplayId] = true

		playerMap := make(map[int]bool)
		for _, player := range workIvi.players {
			if playerMap[player.VID] {
				return errors.New("RealDisplay has duplicate PixelLayer VID")
			}
			playerMap[player.VID] = true

			psurfaceMap := make(map[int]bool)
			for _, psurface := range player.Psurfaces {
				if psurfaceMap[psurface.VID] {
					return errors.New("PixelLayer has duplicate PixelSurface VID")
				}
				psurfaceMap[psurface.VID] = true
			}
		}
	}

	return nil
}

func generateLocalCommand(lcomm string,
	workIviMap map[int]workIvi,
	pickupPlayersMap map[int][]ulanode.PixelLayer) (*ulanode.LocalCommandReq, error) {

	dcomms := make([]ulanode.RdisplayCommandData, 0)

	for key, workIvi := range workIviMap {
		pickupPlayers := pickupPlayersMap[key]
		if len(pickupPlayers) == 0 {
			continue
		}
		dcomm, err := ulanode.NewRdisplayCommandData(&workIvi.rdisplay, pickupPlayers)
		if err != nil {
			return nil, err
		}

		dcomms = append(dcomms, *dcomm)
	}

	ltq, err := ulanode.NewEmptyLocalCommandReq()
	if err != nil {
		return nil, err
	}
	ltq.Command = lcomm
	ltq.RDComms = dcomms

	return ltq, nil
}

func pickupInitialVScreen(
	wIviMap map[int]workIvi,
	oldwIviMap map[int]workIvi) (*ulanode.LocalCommandReq, error) {

	lcomm := ""
	pickupPlayersMap := make(map[int][]ulanode.PixelLayer, len(wIviMap))

	for key, wIvi := range wIviMap {
		oldwIvi := oldwIviMap[key]
		pickupPlayers := pickupPlayersMap[key]

		if len(oldwIvi.players) == 0 && len(wIvi.players) != 0 {
			lcomm = "initial_vscreen"
			pickupPlayers = ulanode.DupPixelLayerSlice(wIvi.players)

			oldwIviMap[key] = wIvi
			pickupPlayersMap[key] = pickupPlayers
		}
	}

	if lcomm == "" {
		return nil, nil
	}
	ltq, err := generateLocalCommand(lcomm, wIviMap, pickupPlayersMap)
	if err != nil {
		return nil, err
	}

	return ltq, nil
}
