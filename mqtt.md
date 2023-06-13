# MQTT setup

We need to use a custom MQTT image based on Ubuntu 22.04.

## Dockerfile

The contents of the Dockerfile are the following:

```Dockerfile
FROM ubuntu:22.04

RUN DEBIAN_FRONTEND=noninteractive apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get upgrade -y && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y software-properties-common && \
    DEBIAN_FRONTEND=noninteractive apt-get update && \
    DEBIAN_FRONTEND=noninteractive add-apt-repository ppa:mosquitto-dev/mosquitto-ppa -y && \
    DEBIAN_FRONTEND=noninteractive apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y mosquitto

ARG USER_ID
ARG GROUP_ID

RUN addgroup --gid $GROUP_ID hass
RUN adduser --disabled-password --gecos '' --uid $USER_ID --gid $GROUP_ID hass
USER hass

CMD ["mosquitto", "-c", "/mosquitto/config/mosquitto.conf"]
```

## Build image

Run the following command to build the Docker image:

```bash
docker build --build-arg USER_ID=$(id -u) --build-arg GROUP_ID=$(id -g) -t gntouts/mqtt .
```

> Note: You need to build the image on the Home Assistan host machine

## Create dirs and files for the container

```bash
mkdir -p ~/mqtt/mosquitto/config/
mkdir -p ~/mqtt/mosquitto/log/
mkdir -p ~/mqtt/mosquitto/data/

# Create log file
touch ~/mqtt/mosquitto/log/mosquitto.log

# Fill in the required fields in config file
cat << EOF > ~/mqtt/mosquitto/config/mosquitto.conf
allow_anonymous true
listener 1883
persistence true
persistence_location /mosquitto/data/
log_dest file /mosquitto/log/mosquitto.log
include_dir /etc/mosquitto/conf.d
EOF
```

## Run the container

```bash
docker run -d --restart unless-stopped -p 1883:1883 --user "$(id -u):$(id -g)" -v ~/mqtt/mosquitto:/mosquitto/ gntouts/mqtt
```
