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

const UHMI_AGL_WM_SOCK string = "/tmp/uhmi-agl-wm_sock"

var MAGIC_CODE []byte = []byte{0x55, 0x4C, 0x41, 0x30} // 'ULA0' ascii code

const VERSION string = "1.0.0"
const OPACITY float64 = 1.0

type AglSurfaceJson struct {
	Id         int     `json:"id"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	SrcX       int     `json:"src_x"`
	SrcY       int     `json:"src_y"`
	SrcW       int     `json:"src_w"`
	SrcH       int     `json:"src_h"`
	DstX       int     `json:"dst_x"`
	DstY       int     `json:"dst_y"`
	DstW       int     `json:"dst_w"`
	DstH       int     `json:"dst_h"`
	Opacity    float64 `json:"opacity"`
	Visibility int     `json:"visibility"`
}

type AglLayerJson struct {
	Id         int              `json:"id"`
	Width      int              `json:"width"`
	Height     int              `json:"height"`
	SrcX       int              `json:"src_x"`
	SrcY       int              `json:"src_y"`
	SrcW       int              `json:"src_w"`
	SrcH       int              `json:"src_h"`
	DstX       int              `json:"dst_x"`
	DstY       int              `json:"dst_y"`
	DstW       int              `json:"dst_w"`
	DstH       int              `json:"dst_h"`
	Opacity    float64          `json:"opacity"`
	Visibility int              `json:"visibility"`
	Surface    []AglSurfaceJson `json:"surfaces"`
}

type AglRDisplay struct {
	RDisplayId  int            `json:"id"`
	InsertOrder string         `json:"insert_order"`
	ReferenceId int            `json:"referenceID"`
	Layers      []AglLayerJson `json:"layers"`
}

type InitialScreenProtocol struct {
	Version  string        `json:"version"`
	Command  string        `json:"command"`
	RDisplay []AglRDisplay `json:"screens"`
}
