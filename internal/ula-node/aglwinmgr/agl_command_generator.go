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
	"errors"
	"ula-tools/internal/ula-node"
)

type workAgl struct {
	rdisplay ulanode.RealDisplay

	/* Geometry is Coordinated by RDisplay  */
	players []ulanode.PixelLayer
}

type AglCommandGenerator struct {
	workAglMap map[int]workAgl
}

func NewAglCommandGenerator(
	sps *ulanode.NodePixelScreens) (*AglCommandGenerator, error) {

	splitSps, err := splitLayer(sps.Dup())
	if err != nil {
		return nil, errors.New("splitLayer error")
	}
	wAglMap, err := generateWorkAgl(splitSps)
	if err != nil {
		return nil, errors.New("generateWorkAgl error")
	}

	aglltg := &AglCommandGenerator{
		workAglMap: wAglMap,
	}

	return aglltg, nil
}

func (aglltg *AglCommandGenerator) GenerateLocalCommandReq(acdata *ulanode.ApplyCommandData) ([]*ulanode.LocalCommandReq, error) {
	ltqs := []*ulanode.LocalCommandReq{}

	wAglMap := aglltg.workAglMap

	ret, err := pickupInitialVScreen(wAglMap)
	if err == nil && ret != nil {
		ltqs = append(ltqs, ret)
	}

	return ltqs, nil
}

func generateWorkAgl(
	spscrns *ulanode.NodePixelScreens) (map[int]workAgl, error) {

	workAglMap := make(map[int]workAgl)

	for _, pscrn := range spscrns.Pscreens {
		wvdisp := workAgl{
			rdisplay: *pscrn.Rdisplay.Dup(),
			players:  ulanode.DupPixelLayerSlice(pscrn.Players),
		}
		workAglMap[wvdisp.rdisplay.RDisplayId] = wvdisp
	}

	err := isValidWorkAglMap(workAglMap)

	return workAglMap, err
}

func isValidWorkAglMap(wAglMap map[int]workAgl) error {

	rdisplayMap := make(map[int]bool)
	for _, workAgl := range wAglMap {
		if rdisplayMap[workAgl.rdisplay.RDisplayId] {
			return errors.New("PixelScreens has duplicate RDisplayId")
		}
		rdisplayMap[workAgl.rdisplay.RDisplayId] = true

		playerMap := make(map[int]bool)
		for _, player := range workAgl.players {
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
	workAglMap map[int]workAgl,
	pickupPlayersMap map[int][]ulanode.PixelLayer) (*ulanode.LocalCommandReq, error) {

	dcomms := make([]ulanode.RdisplayCommandData, 0)

	for key, workAgl := range workAglMap {
		pickupPlayers := pickupPlayersMap[key]
		if len(pickupPlayers) == 0 {
			continue
		}
		dcomm, err := ulanode.NewRdisplayCommandData(&workAgl.rdisplay, pickupPlayers)
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
	wAglMap map[int]workAgl) (*ulanode.LocalCommandReq, error) {

	lcomm := ""
	pickupPlayersMap := make(map[int][]ulanode.PixelLayer, len(wAglMap))

	for key, wAgl := range wAglMap {
		pickupPlayers := pickupPlayersMap[key]

		if len(wAgl.players) != 0 {
			lcomm = "initial_vscreen"
			pickupPlayers = ulanode.DupPixelLayerSlice(wAgl.players)

			pickupPlayersMap[key] = pickupPlayers
		}
	}

	if lcomm == "" {
		return nil, nil
	}
	ltq, err := generateLocalCommand(lcomm, wAglMap, pickupPlayersMap)
	if err != nil {
		return nil, err
	}

	return ltq, nil
}
