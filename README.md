# Unified Layout Tools (ULA)

> Unified Layout Tools (ULA) is a framework for achieving flexible layout of applications.
> This framework enables mapping of physical displays onto a virtual screen.
> It provides the ability to apply layout settings such as display position and size on the virtual screen to launched applications.

## Contents

- [Unified Layout Tools (ULA)](#architecture)
  - [Contents](#contents)
  - [Repository structure](#repository-structure)
  - [How to install](#how-to-install)
    - [Golang setup](#golang-setup)
    - [Build ULA framework](#build-ula-framework)
  - [How to Use](#how-to-use)
    - [Json settings](#json-settings)
    - [How to install uhmi-ivi-wm](#how-to-install-uhmi-ivi-wm)
    - [Run Weston](#Run-Weston)
    - [Run Wayland Application](#Run-Wayland-Application)
    - [Run uhmi-ivi-wm](#Run-uhmi-ivi-wm)
    - [Run ula-node](#Run-ula-node)
    - [Run ula-distrib-com](#Run-ula-distrib-com)
    - [Run dwm API](#Run-dwm-API)

## Repository structure

```
.
├── CONTRIBUTING.md
├── LICENSE.md
├── Makefile
├── README.md
├── cmd
│   ├── Makefile
│   ├── ula-distrib-com
│   │   ├── Makefile
│   │   └── ula_distrib_com.go
│   └── ula-node
│       ├── Makefile
│       └── main.go
├── example
│   ├── dwm
│   │   └── EGLWLInputEventExample
│   │       └── dwm_initial_layout.json
│   ├── global
│   │   ├── initial-vscreen.json
│   │   └── virtual-screen-def.json
│   └── vdisplay
│       ├── initial-vscreen.json
│       └── virtual-screen-def.json
├── internal
│   ├── Makefile
│   ├── ula
│   │   ├── common_type.go
│   │   ├── env.go
│   │   └── virtual_screen_def.go
│   ├── ula-client
│   │   ├── Makefile
│   │   ├── core
│   │   │   └── core_type.go
│   │   ├── dwmapi
│   │   │   ├── Makefile
│   │   │   └── dwmapi.go
│   │   ├── readclusterapp
│   │   │   ├── Makefile
│   │   │   └── read_cluster_app.go
│   │   ├── ulacommgen
│   │   │   ├── Makefile
│   │   │   ├── ula_comm_generator.go
│   │   │   └── ula_comm_generator_protocol.go
│   │   └── ulamulticonn
│   │       ├── Makefile
│   │       └── ula_multi_conn.go
│   ├── ula-node
│   │   ├── Makefile
│   │   ├── common_type.go
│   │   ├── iviwinmgr
│   │   │   ├── Makefile
│   │   │   ├── ivi_command_generator.go
│   │   │   ├── ivi_layer_split.go
│   │   │   ├── iviwinmgr.go
│   │   │   └── iviwinmgr_protocol.go
│   │   ├── ula_parser.go
│   │   ├── virtual_screen.go
│   │   └── vs2rd
│   │       ├── Makefile
│   │       └── vscreen_to_rdisplay_converter.go
│   └── ulog
│       ├── Makefile
│       └── ulog.go
└── pkg
    └── ula-client-lib
        ├── Makefile
        └── ula_client.go
```

# How to install
## Golang setup
Before building ULA API, you need to install and configure Golang.
```
sudo apt-get install golang
export GOPATH=<your go work directory>
export GOBIN=$GOPATH/bin
export PATH=$GOBIN:$PATH
```

## Build ULA framework
You can easily build ULA framweork using make.
```
mkdir -p $GOPATH/src
cd $GOPATH/src
git clone https://github.com/unified-hmi/ula-tools.git
cd ula-tools
make
```

# How to use
ULA is possible to flexibly configure and apply layouts for applications.

## Json settings
To run ULA, three Json files are required.
* virtual-screen-def.json: execution environment such as display and node (SoCs/VMs/PCs) informations.
* initial-vscreen.json: application layout information such as position and size.
* dwm_initial_layout.json: application layout information used by dwm API.


Json files need to be created correctly for your execution environment.
Sample Json files are locates in the "$GOPATH/ula-tools/example" directory.

- Parameters of initial-vscreen.json
  - vlayer: define a virtual layer that represents a group of surfaces within the virtual screen. Each layer has a unique Virtual ID (VID) and can contain multiple surfaces. virtual_w and virtual_h define vlayer's size. The layer's source (vsrc_x, vsrc_y, vsrc_w, vsrc_h) and destination (vdst_x, vdst_y, vdst_w, vdst_h) coordinates determine where and how large the layer appears on the virtual screen.
  - vsurface: define individual surfaces within the virtual layer. Each surface also has a VID, and its pixel dimensions (pixel_w, pixel_h) represent the actual size of the content. The source (psrc_x, psrc_y, psrc_w, psrc_h) and destination (vdst_x, vdst_y, vdst_w, vdst_h) coordinates determine the portion of the content to display and its location within the layer.
  - coord: vlayer is possible to set the position in two coordinate systems. In the global coordinate system, it defines where it is in relation to the origin of the virtual screen. In the vdisplay coordinate system, it defines where it is in relation to the origin of the display with the specified ID (vdisplay_id).
  - visibility: define visibility of vlayer/vsurface. Display when it is 1, and hide when it is 0.


## How to install uhmi-ivi-wm
When using ULA, this module enables application display on Weston.
For instructions on how to install uhmi-ivi-wm, please refer to the [README](https://github.com/unified-hmi/uhmi-ivi-wm/blob/main/README.md).

## Run Weston
```
weston --width=1920 --height=1080 --outout-count=2 &
```
**Note:** The option "--output-count" indicates the number of Westons to be launched. Please specify the number according to the configuration file.

## Run Wayland Application
You can run the wayland-ivi-extension sample app EGLWLInputEventExample to see how the layout works. Of course it will work with other wayland apps that support ivi_application.
```
EGLWLInputEventExample &
```
**Note:** The launched application's surface ID is used in the json file.

## Run uhmi-ivi-wm
To send the application layout information generated by ULA, uhmi-ivi-wm needs to be launched.

```
uhmi-ivi-wm &
```

## Run ula-node
ula-node receives initial display layout commands from ula-distrib-com and generates local commands from virtual-screen-def.json to send the json file to uhmi-ivi-wm.
This ula-node needs virtual-screen-def.json, so please set it with "-f" option. (default path is /etc/uhmi-framework/virtual-screen-def.json)

- Options of ula-node
  - -H: search ula-node param by hostname from VScrnDef file
  - -N: search ula-node param by node_id from VScrnDef file (default: -1)
  - -d: verbose debug log
  - -f: virtual-screen-def.json file Path (default: "/etc/uhmi-framework/virtual-screen-def.json")
  - -v: verbose info log (default true)

```
ula-node -f <path to virtual-screen-def.json> &
```

## Run ula-distrib-com
ula-distrib-com sends the applications layout file to ula-node.
This ula-distrib-com needs virtual-screen-def.json and initial-vscreen.json, so please set virtual-screen-def.json with argument and set initial-vscreen.json with iostream.

- Options of ula-distrib-com
  - -d: verbose debug log
  - -f: force the execution of the application even if some nodes are not alive.
  - -v: verbose info log (default: true)

```
cat <path to initial-vscreen.json> | ula-distrib-com <path to virtual-screen-def.json>
```

## Run dwm API
ULA also provides a C language shared library (default: generated in $GOPATH/pkg/libulaclient).
By using the library's API, it's easy to display the application's layout.

- Description of dwm API
  - dwm_init: initialize the parameters
  - dwm_set_system_layout: display the layout according to dwm_initial_layout.json

Please create a C language source code with content similar to the following as an example:
```c
#include <stdio.h>
#include "libulaclient.h"

int main(void)  
{
	dwm_init();
	dwm_set_system_layout();

	return 0;
}
```

Compile this source code with gcc as follows:
```
gcc -I./ -L./ sample.c -lulaclient -o sample.out
```

Execute the dwm API with a command as follows:
```
export VSDPATH="<path to virtual-screen-def.json>"
export DWMPATH="<path to dwm_initial_layout.json>"
export LD_LIBRARY_PATH="<path to libulaclient>"
./sample.out
```
**Note:** The default path is as follws:  
  - VSDPATH: "/etc/uhmi-framework/virtual-screen-def.json"  
  - DWMPATH: "/var/local/uhmi-app/dwm"