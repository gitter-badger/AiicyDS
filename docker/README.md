# Docker for Peach

## Usage

To keep your data out of Docker container, we do a volume (`/var/aiicyds` -> `/data/aiicyds`) here, and you can change it based on your situation.

```
# Pull image from Docker Hub.
$ docker pull Aiicy/AiicyDS

# Create local directory for volume.
$ mkdir -p /var/aiicyds

# Use `docker run` for the first time. 
# Peach will complain about missing custom app.ini, leave it there and see Settings section below.
$ docker run --name=aiicyds -p 5555:5555 -v /var/aiicyds:/data/aiicyds Aiicy/AiicyDS

# Use `docker start` if you have stopped it.
$ docker start aiicyds
```

Files will be store in local path `/var/aiicyds` in my case.

Directory `/var/aiicyds` keeps Git repositories and Gogs data:

    /var/aiicyds
    |-- custom
    |-- data
    |-- log

### Volume with data container

If you're more comfortable with mounting data to a data container, the commands you execute at the first time will look like as follows:

```
# Create data container
docker run --name=aiicyds-data --entrypoint /bin/true Aiicy/AiicyDS

# Use `docker run` for the first time.
docker run --name=aiicyds --volumes-from aiicyds-data -p 5555:5555 Aiicy/AiicyDS
```

#### Using Docker 1.9 Volume command

```
# Create docker volume.
$ docker volume create --name aiicyds-data

# Use `docker run` for the first time.
$ docker run --name=aiicyds -p 5555:5555 -v aiicyds-data:/data/aiicyds Aiicy/AiicyDS
```

## Settings



## Upgrade

:exclamation::exclamation::exclamation:<span style="color: red">**Make sure you have volumed data to somewhere outside Docker container**</span>:exclamation::exclamation::exclamation:

Steps to upgrade Peach with Docker:

- `docker pull Aiicy/AiicyDS`
- `docker stop aiicyds`
- `docker rm aiicyds`
- Finally, create container as the first time and don't forget to do same volume and port mapping.

## Known Issues

- The docker container can not currently be build on Raspberry 1 (armv6l) as our base image `alpine` does not have a `go` package available for this platform.
