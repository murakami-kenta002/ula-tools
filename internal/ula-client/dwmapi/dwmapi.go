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

package dwmapi

import (
	"ula-tools/internal/ula"
	"ula-tools/internal/ula-client/readclusterapp"
	"ula-tools/internal/ula-client/ulacommgen"
	"ula-tools/internal/ula-client/ulamulticonn"
	. "ula-tools/internal/ulog"
)

func DwmSetSystemLayout() {

	calayoutTree, err := readclusterapp.ReadCALayoutTreeFromCfg()
	if err != nil {
		ELog.Println(err)
		return
	}

	var layoutComm string
	layoutComm, err = ulacommgen.GenerateUlaCommInitialVscreen(calayoutTree)
	if err != nil {
		ELog.Println(err)
		return
	}

	err = ulamulticonn.UlaMulCon.SendLayoutCommand(layoutComm)
	if err != nil {
		ELog.Println(err)
		return
	}
}

func DwmInit() {
	force := ula.GetEnvBool("ULA_FORCE", false)
	err := ulamulticonn.UlaConnectionInit(force)
	if err != nil {
		ELog.Println(err)
		return
	}
}
