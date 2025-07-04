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
    - [How to control layouts on Weston ivi-shell](#how-to-control-layouts-on-weston-ivi-shell)
      - [How to install uhmi-ivi-wm](#how-to-install-uhmi-ivi-wm)
      - [Run Weston](#run-weston)
      - [Run Wayland application on Weston](#run-wayland-application-on-weston)
      - [Run uhmi-ivi-wm](#run-uhmi-ivi-wm)
      - [Control layouts on Weston ivi-shell by ULA](#control-layouts-on-weston-ivi-shell-by-ula)
    - [How to control layouts on RVGPU compositor](#how-to-control-layouts-on-rvgpu-compositor)
      - [How to install RVGPU compositor](#how-to-install-rvgpu-compositor)
      - [Run RVGPU compositor](#run-rvgpu-compositor)
      - [Run Wayland application on RVGPU compositor](#run-wayland-application-on-rvgpu-compositor)
      - [Control layouts on RVGPU compositor by ULA](#control-layouts-on-rvgpu-compositor-by-ula)

## Repository structure

```
.
├── cmd
│   ├── Makefile
│   ├── ula-distrib-com
│   │   ├── Makefile
│   │   └── ula_distrib_com.go
│   └── ula-node
│       ├── main.go
│       └── Makefile
├── CONTRIBUTING.md
├── example
│   ├── dwm
│   │   └── EGLWLInputEventExample
│   │       └── dwm_initial_layout.json
│   ├── initial-vscreen
│   │   ├── global
│   │   │   └── initial-vscreen.json
│   │   └── vdisplay
│   │       └── initial-vscreen.json
│   └── vsd
│       ├── iviwinmgr
│       │   └── virtual-screen-def.json
│       └── rvgpuwinmgr
│           └── virtual-screen-def.json
├── internal
│   ├── Makefile
│   ├── ula
│   │   ├── common_type.go
│   │   ├── env.go
│   │   └── virtual_screen_def.go
│   ├── ula-client
│   │   ├── core
│   │   │   └── core_type.go
│   │   ├── dwmapi
│   │   │   ├── dwmapi.go
│   │   │   └── Makefile
│   │   ├── Makefile
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
│   │   ├── common_type.go
│   │   ├── iviwinmgr
│   │   │   ├── ivi_command_generator.go
│   │   │   ├── ivi_layer_split.go
│   │   │   ├── iviwinmgr.go
│   │   │   ├── iviwinmgr_protocol.go
│   │   │   └── Makefile
│   │   ├── Makefile
│   │   ├── rvgpuwinmgr
│   │   │   ├── Makefile
│   │   │   ├── rvgpu_command_generator.go
│   │   │   ├── rvgpuwinmgr.go
│   │   │   └── rvgpuwinmgr_protocol.go
│   │   ├── ula_parser.go
│   │   ├── virtual_screen.go
│   │   └── vs2rd
│   │       ├── Makefile
│   │       └── vscreen_to_rdisplay_converter.go
│   └── ulog
│       ├── Makefile
│       └── ulog.go
├── LICENSE.md
├── Makefile
├── make_ula.sh
├── pkg
│   └── ula-client-lib
│       ├── Makefile
│       └── ula_client.go
└── README.md
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
You can easily build ULA framework using make.
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
To run ULA, some Json files are required.
* virtual-screen-def.json: execution environment such as display and node (SoCs/VMs/PCs) informations.
* initial-vscreen.json: application layout information such as position and size.
* dwm_initial_layout.json: application layout information used by dwm API.

Json files need to be created correctly for your execution environment.
Sample Json files are located in the "$GOPATH/src/ula-tools/example" directory.

To control layouts, initial-vscreen.json and dwm_initial_layout.json have following parameters:

- vlayer: define a virtual layer that represents a group of surfaces within the virtual screen. Each layer has a unique Virtual ID (VID) and can contain multiple surfaces. virtual_w and virtual_h define vlayer's size. The layer's source (vsrc_x, vsrc_y, vsrc_w, vsrc_h) and destination (vdst_x, vdst_y, vdst_w, vdst_h) coordinates determine where and how large the layer appears on the virtual screen.
- vsurface: define individual surfaces within the virtual layer. Each surface also has a VID, and its pixel dimensions (pixel_w, pixel_h) represent the actual size of the content. The source (psrc_x, psrc_y, psrc_w, psrc_h) and destination (vdst_x, vdst_y, vdst_w, vdst_h) coordinates determine the portion of the content to display and its location within the layer.
- coord: vlayer is possible to set the position in two coordinate systems. In the global coordinate system, it defines where it is in relation to the origin of the virtual screen. In the vdisplay coordinate system, it defines where it is in relation to the origin of the display with the specified ID (vdisplay_id).
- visibility: define visibility of vlayer/vsurface. Display when it is 1, and hide when it is 0.


**Note:** [Here](https://docs.automotivelinux.org/en/master/#06_Component_Documentation/11_Unified_HMI/) is the documentation for verifying the operation of the Unified HMI framework on AGL and the detailed explanation about Json files.

## How to control layouts on Weston ivi-shell
ULA has a plugin for supporting Weston ivi-shell.
When you want to use ULA to control layouts on Weston, you should prepare `uhmi-ivi-wm`, which is one the Unified HMI frameworks.
The ULA plugin can change layouts through `uhmi-ivi-wm` on Weston.

### How to install uhmi-ivi-wm
For instructions on how to install `uhmi-ivi-wm`, please refer to the [README](https://github.com/unified-hmi/uhmi-ivi-wm).

### Run Weston
```
weston --width=1920 --height=1080 --outout-count=2 &
```
**Note:** The option "--output-count" indicates the number of Westons to be launched. Please specify the number according to the configuration file.

### Run Wayland application on Weston
You can run the wayland-ivi-extension sample app EGLWLInputEventExample to see how the layout works. Of course it will work with other wayland apps that support ivi_application.
```
EGLWLInputEventExample &
```
**Note:** The launched application's surface ID is used in the json file.

### Control layouts on Weston ivi-shell by ULA
After launching wayland applications on Weston, you can control their layouts on Weston by using ULA tools and `uhmi-ivi-wm`.

#### Run uhmi-ivi-wm
To send the application layout information generated by ULA, `uhmi-ivi-wm` needs to be launched.

```
uhmi-ivi-wm &
```

#### Run ula-node
ula-node has a porting layer to determine which plugin to use, `iviwinmgr` or `rvgpuwinmgr`.
If you define the "compositor" section in virtual-screen-def.json, ula-node branches to use `rvgpuwinmgr` plugin, and if you don't define it, `iviwinmgr` plugin is used.

ula-node receives initial display layout commands from ula-distrib-com and generates local commands from virtual-screen-def.json to send the json file to the `uhmi-ivi-wm` or `rvgpu-renderer`.
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

#### Run ula-distrib-com
ula-distrib-com sends the applications layout file to ula-node.
This ula-distrib-com needs virtual-screen-def.json and initial-vscreen.json, so please set virtual-screen-def.json with argument and set initial-vscreen.json with iostream.

- Options of ula-distrib-com
  - -d: verbose debug log
  - -f: force the execution of the application even if some nodes are not alive.
  - -v: verbose info log (default: true)

```
cat <path to initial-vscreen.json> | ula-distrib-com <path to virtual-screen-def.json>
```

#### Run dwm API
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
export DWMPATH="<path to the directory that contains the dwm_initial_layout.json>"
export LD_LIBRARY_PATH="<path to libulaclient>"
./sample.out
```
**Note:** The default path is as follws:  
  - VSDPATH: "/etc/uhmi-framework/virtual-screen-def.json"  
  - DWMPATH: "/var/local/uhmi-app/dwm"

## How to control layouts on RVGPU compositor
Remote VIRTIO GPU (RVGPU) is one of the main components of Unified HMI frameworks.
RVGPU is a client-server based rendering engine, which allows to render 3D on one device (client) and display it via network on another device (server).
RVGPU has compositor functions, and it can composite multiple applications on it.
By combining it with ULA, it is possible to control the layouts of applications on the RVGPU compositor.

### How to install RVGPU compositor
For instructions on how to install RVGPU, please refer to the [README](https://github.com/unified-hmi/remote-virtio-gpu).

### Run RVGPU compositor
For instructions on how to run RVGPU compositor, please refer to the [README](https://github.com/unified-hmi/remote-virtio-gpu).

#### Run rvgpu-renderer
To combine ULA with RVGPU compositor, you have to run `rvgpu-renderer` with "-l" option.
Additionaly, you need to add "-d" option to use the same socket domain to connect ula-node with `rvgpu-renderer`.

e.g. launch `rvgpu-renderer` with the socket domain "rvgpu-compositor-0"
```
rvgpu-renderer -b 1920x1080@0,0 -p 36000 -a -d rvgpu-compositor-0 -l
```

#### Run rvgpu-proxy
After launching `rvgpu-renderer`, you can run multiple `rvgpu-proxy` and applications on one `rvgpu-renderer`.
Please add the application name, which is the same as defined in initial-vscreen.json or dwm_initial_layout.json, with "-i" option.

e.g. launch `rvgpu-proxy` with the application name "wayland-app"
```
rvgpu-proxy -s 1920x1080@0,0 -n 127.0.0.1:36000 -i wayland-app
```

### Run Wayland application on RVGPU compositor
If you wish to display Wayland applications remotely, you can optionally install `rvgpu-wlproxy`, which can be used as a lightweight Wayland server instead of Weston for similar functionality. For detailed installation, configuration, and usage instructions, please refer to the [README](https://github.com/unified-hmi/rvgpu-wlproxy).

### Control layouts on RVGPU compositor by ULA
After launching Wayland applications with these steps, you can control layouts on RVGPU compositor by using ULA tools.
For instructions on how to run ULA tools in detail, please refer to the [Control layouts on Weston ivi-shell by ULA](#control-layouts-on-weston-ivi-shell-by-ula)

#### Run ula-node
To connect with `rvgpu-renderer`, please define the "compositor" section in virtual-screen-def.json and add socket domain name in that section, using the "sock_domain_name" key, with the value added on `rvgpu-renderer` as the "-d" option.
The sample virtual-screen-def.json is located in the "$GOPATH/src/ula-tools/example/vsd/rvgpuwinmgr/virtual-screen-def.json".

After creating virtual-screen-def.json, please run the command below.
```
ula-node -f <path to virtual-screen-def.json>
```

#### Run ula-distrib-com
To control layouts of the specific application running on RVGPU compositor, please add the "appli_name" key, with the value added on `rvgpu-proxy` as the "-i" option, in both the "vlayer" and "vsurface" sections.
The sample initial-vscreen.json is located in the "$GOPATH/src/ula-tools/example/global/initial-vscreen.json".

After creating initial-vscreen.json, please run the command below.
```
cat <path to initial-vscreen.json> | ula-distrib-com <path to virtual-screen-def.json>
```

#### Run dwm API
To control layouts of the specific application running on RVGPU compositor, please add the "application_name" key, with the value added on `rvgpu-proxy` as "-i" option.

After creating dwm_initial_layout.json, please run the commands below.
```
export VSDPATH="<path to virtual-screen-def.json>"
export DWMPATH="<path to the directory that contains the dwm_initial_layout.json>"
export LD_LIBRARY_PATH="<path to libulaclient>"
./sample.out
```
