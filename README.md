# gomuq

gomuq provides 2 CLI tools used by Home Assistant host system.

## Tools

### nt

`nt` is a CLI tool to wrap st-link functionalities and provide helper methods to search for USB devices.

#### nt Usage

```bash
NAME:
   nt - A new cli application

USAGE:
   nt [global options] command [command options] [arguments...]

VERSION:
   0.0.1

DESCRIPTION:
   nt is a simple CLI tool that wraps st-link functionalities and provides helper methods to search for USB devices

COMMANDS:
   reset, r     resets the STM32 board that is currently connected
   flash, f     flashes given binary to the STM32 board
     OPTIONS:
       --binary value, -b value  binary to flash
   list, l, ls  list all connected USB devices
   search, s    search connected USB devices
     OPTIONS:
       --term value, -t value  term to search for
   help, h      Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

### uamq

`uamq` is a sort of relay program. It is used to decode and encode UART messages to MQTT messages and vice versa.

#### Message Definitions

Each UART message field gets its own MQTT topic (separate for sub/pub) and Redis key

|  Go Struct     |         MQTT Pub       |       MQTT Sub      | Redis key |
| -------------- | ---------------------- |-------------------- | ---------- |
|  m.data["b0"]  |  hass/listen_hass/b0   | hass/listen_go/b0   |    b0      |
|  m.data["b1"]  |  hass/listen_hass/b1   | hass/listen_go/b1   |    b1      |
|  m.data["b2"]  |  hass/listen_hass/b2   | hass/listen_go/b2   |    b2      |
|  m.data["b3"]  |  hass/listen_hass/b3   | hass/listen_go/b3   |    b3      |
|  m.data["b4"]  |  hass/listen_hass/b4   | hass/listen_go/b4   |    b4      |
|  m.data["b5"]  |  hass/listen_hass/b5   | hass/listen_go/b5   |    b5      |
|  m.data["b6"]  |  hass/listen_hass/b6   | hass/listen_go/b6   |    b6      |
|  m.data["b7"]  |  hass/listen_hass/b7   | hass/listen_go/b7   |    b7      |
|  m.data["b8"]  |  hass/listen_hass/b8   | hass/listen_go/b8   |    b8      |
|  m.data["b9"]  |  hass/listen_hass/b9   | hass/listen_go/b9   |    b9      |
| m.data["b10"]  |  hass/listen_hass/b10  | hass/listen_go/b10  |    b10     |
| m.data["b11"]  |  hass/listen_hass/b11  | hass/listen_go/b11  |    b11     |
| m.data["b12"]  |  hass/listen_hass/b12  | hass/listen_go/b12  |    b12     |
| m.data["b13"]  |  hass/listen_hass/b13  | hass/listen_go/b13  |    b13     |
| m.data["b14"]  |  hass/listen_hass/b14  | hass/listen_go/b14  |    b14     |
| m.data["b15"]  |  hass/listen_hass/b15  | hass/listen_go/b15  |    b15     |
| m.data["b16"]  |  hass/listen_hass/b16  | hass/listen_go/b16  |    b16     |
| m.data["b17"]  |  hass/listen_hass/b17  | hass/listen_go/b17  |    b17     |
| m.data["b18"]  |  hass/listen_hass/b18  | hass/listen_go/b18  |    b18     |
| m.data["b19"]  |  hass/listen_hass/b19  | hass/listen_go/b19  |    b19     |
| m.data["b20"]  |  hass/listen_hass/b20  | hass/listen_go/b20  |    b20     |
| m.data["b21"]  |  hass/listen_hass/b21  | hass/listen_go/b21  |    b21     |
| m.data["b22"]  |  hass/listen_hass/b22  | hass/listen_go/b22  |    b22     |
| m.data["b23"]  |  hass/listen_hass/b23  | hass/listen_go/b23  |    b23     |
|  m.data["f0"]  |  hass/listen_hass/f0   | hass/listen_go/f0   |    f0      |
|  m.data["f1"]  |  hass/listen_hass/f1   | hass/listen_go/f1   |    f1      |
|  m.data["f2"]  |  hass/listen_hass/f2   | hass/listen_go/f2   |    f2      |
|  m.data["i0"]  |  hass/listen_hass/i0   | hass/listen_go/i0   |    i0      |
|  m.data["i1"]  |  hass/listen_hass/i1   | hass/listen_go/i1   |    i1      |
|  m.data["i2"]  |  hass/listen_hass/i2   | hass/listen_go/i2   |    i2      |

#### Additional components

`uamq` relies on a Redis server and an MQTT broker to operate.

##### Redis

Currently a Redis instance running at localhost:6379 is required.
To achieve this, consider running Redis inside a Docker container:

```bash
docker pull redis:7.0-alpine
docker run -d --restart unless-stopped -p 6379:6379 redis:7.0-alpine
```

##### MQTT

We need an MQTT broker running at localhost:1883/

To achieve this, check [MQTT setup](mqtt.md).

#### Config file

`uamq` config must be stored in a `.yaml` file. You need to pass the path to that file as a parameter. A template file is created when building `uamq` under the `dist` directory.

The file should have the following fields:

```yaml
mqtt:
  host: localhost
  port: 1883

redis:
  host: localhost
  port: 6379

uart:
# identifier of the usb device (you can use nt to find the name and/or id of your device)
  name: arduino
  baudrate: 115200

log:
# log file path
  target: /home/hass/log/uamq/uamq.log
```

#### uamq Usage

```bash
uamq is a utility to decode and encode UART messages to MQTT messages and vice versa

Usage:
        uamq -h           Displays this help message
        uamq --help       Displays this help message
        uamq config.yaml  Starts the uamq utility
```

## Build from source

### Requirements

#### go

You need to have `go1.18` or higher installed (for instructions, check [here](https://go.dev/doc/install)).

#### libusb-1.0-0-dev

You need to have `libusb-1.0-0-dev` installed. For Debian-based distros:

```bash
sudo apt-get install -y libusb-1.0-0-dev
```

### Build and install

```bash
make
sudo make install
```

### Uninstall

```bash
sudo make uninstall
```
