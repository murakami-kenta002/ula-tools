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

import "C"
import (
	"ula-tools/internal/ula-client/dwmapi"
)

//export dwm_set_system_layout
func dwm_set_system_layout() C.int {
	val := dwmapi.DwmSetSystemLayout()
	return C.int(val)
}

//export dwm_init
func dwm_init() C.int {
	val := dwmapi.DwmInit()
	return C.int(val)
}

func main() {}
