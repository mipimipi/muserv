# muserv

[![Go Report Card](https://goreportcard.com/badge/gitlab.com/mipimipi/muserv)](https://goreportcard.com/report/gitlab.com/mipimipi/muserv)
[![REUSE status](https://api.reuse.software/badge/gitlab.com/mipimipi/muserv)](https://api.reuse.software/info/gitlab.com/mipimipi/muserv)
[![pipeline status](https://gitlab.com/mipimipi/muserv/badges/master/pipeline.svg)](https://gitlab.com/mipimipi/muserv/-/commits/master)

Simple command line [UPnP](https://en.wikipedia.org/wiki/Universal_Plug_and_Play) music server for Linux that allows a flexible structuring of your music in content hierarchies. Supported music file types include [MP3](https://en.wikipedia.org/wiki/MP3), [FLAC](https://en.wikipedia.org/wiki/FLAC), [Ogg Vorbis](https://en.wikipedia.org/wiki/Vorbis), [Opus](https://en.wikipedia.org/wiki/Opus_(audio_format)), [AAC](https://en.wikipedia.org/wiki/Advanced_Audio_Coding), [Alac](https://en.wikipedia.org/wiki/Apple_Lossless) and [MP4/M4a](https://en.wikipedia.org/wiki/MPEG-4_Part_14).  In addition muserve can read [M3U playlists](https://en.wikipedia.org/wiki/M3U) in simple and extended format.

## Installation

### Installation with Package Managers

For [Arch Linux](https://archlinux.org/) (and other Linux distros that can install packages from the Arch User Repository) there's a [muserv package in AUR](https://aur.archlinux.org/packages/muserv-git/).

### Manual Installation

#### Build muserv
muserv is written in [Golang](https://golang.org/) and thus requires the installation of [Go](https://golang.org/project/). Make sure that you've set the environment variable `GOPATH` accordingly, and make also sure that [git](https://git-scm.com/) is installed.

To download muserv and all dependencies, open a terminal and enter

    $ go get gitlab.com/mipimipi/muserv

After that, build muserv by executing

    $ cd $GOPATH/src/gitlab.com/mipimipi/muserv
    $ make

Finally, execute

    $ make install

as `root` to copy the muserv binary to `/usr/bin`.

#### Create the muserv system user

muserv requires the system user `muserv`. To create it with systemd, just execute

    $ cp $GOPATH/src/gitlab.com/mipimipi/muserv/cfg/sysusers.conf /usr/lib/sysusers.d/muserv.conf
    $ systemd-sysusers

as `root`. Without systemd, create the user with the corresponding command of your distribution. The user does neither require a home directory nor a shell.

Make sure that the muserv system user has read access to your music directory.

#### Create directories

Create a cache and a log directory for muserv (per default, that's `/var/cache/muserv` and `/var/log/muserv`):

    $ mkdir /var/cache/muserv
    $ mkdir /var/log/muserv

Make sure that the muserv system user has write access to both directories.

#### Copy configuration files

To copy the muserv configuration files, execute

    $ mkdir /etc/muserv
    $ cd $GOPATH/src/gitlab.com/mipimipi/muserv/cfg
    $ cp config-default.json /etc/muserv
    $ cp ConnectionManager.xml /etc/muserv
    $ cp ContentDirectory.xml /etc/muserv

as `root`.

## Configuration

Copy the default configuration by executing 

    $ cp /etc/muserv/config-default.json /etc/muserv/config.json

as `root`.

The only required change is to set the music directory. Therefore, edit `/etc/muserv/config.json` as `root` and set `music_dir` to the absolute path of your music directory. [Here](doc/configuration.md) you find a more detailed description of the configuration options.

## Run muserv

Execute

    $ muserv test
    
to check if the muserv configuration is complete and consistent.

muserv comes with a systemd service. Start it by executing 

    $ systemctl start muserv.service
    
as `root`. If you did not install muserv via package manager, you have to copy the service file prior to that by executing

    $ cp $GOPATH/src/gitlab.com/mipimipi/muserv/systemd/muserv.service /etc/systemd/system/muserv.service

as `root`.

Now, you can see the muserv status by entering the URL <IP-address-of-your-server:8008> in a browser.

Have fun with muserv :)

## Limitations

So far, muserv does not support the following features:

* searching for music
* creation of playlists
* manipulation of tag values (for display or sorting, for example)
* UPnP "tracking changes option" - However, muserv keeps track of changes in your music library.