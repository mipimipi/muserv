# Configuration settings

muserv provides default configuration settings. The only additional configuration that must be made for the music directory. 

Below we describe all configuration settings.

* `cnt`

  Content-related parameters

  -  `music_dir`

    Absolute path of music directory. The muserv system user must have read access. There is no default.

  - `separator`

    Some tags can have multiple value. `separator` contains the string that is used as separator for these values. Often `\\` or `;` is used. Default is `;`.

  - `cache_dir`

    Cache directory for muserv. Default is `/var/cache/muserv`. The muserv system user must have write access to this directory.

  - `update_mode` 

    To keep the muserv content up to date if anything in the music directory is changed, muserv provides an update mechanism that runs regularly. `update_mode` specifies which mode is used for that. Two different modes are possible:
    
      - `notify` uses the [inotify system](https://en.wikipedia.org/wiki/Inotify) of the Linux kernel
      - `scan` scans the music directory for changes

    Default is `notify`.    

  - `update_interval` 

     Time in seconds after which muserv checks for changes in the music direcrtory. Default is `60` (one minute).

* `upnp`

  UPnP-related parameters

  - `interfaces` 

    Network interfaces to serve. Per default (i.e. if no interface is configured) all interfaces are served.

  - `port`

    Port for HTTP requests to the server (device and service descriptions, SOAP, media transfer etc.). Default is `8008`.

  - `server_name`
  
    Server name that is shown by UPnP clients. Default is `Music Server`.

  - `uuid`

    Unique identifier for the server. It must be of the form `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`. If `uuid` is not set (which is the default), muserv creates a new identifier. 

  - `max_age`

    Time interval in seconds for the renewal of the alive notification. Default is `86400` (i.e. once per day).
  
  - `status_dir` 
  
    File to persist status information (absolute path). Default is `<CACHE-DIR>/status.json`.

  - `device`
  
    This clubs together parameters that are required for the XML description of the root device. There's a default value for each of these parameters. 

  - `hierarchies`

    Here, the content hierarchies that are shown in the UPnP clients are configured. Per default all available hierarchies are configured. Just remove those that you do not need. You can also change the sequence of the hierarchies. UPnP clients show the hierarchies in the same sequence as they are configured here.
    Each hierarchy needs a name. That's the name that is also displayed by the clients. The name can be adjusted. Some hierarchies have an id. These hierarchies have a special implementation. The id must not be changed, and these hierarchies must not have levels. The other hierarchies don't have an id, but are configured by their levels. `levels: ["genre","albumartist","album"]` for instance, means that Genre is the highest level. For each Genre the tracks are then organized by AlbumArtist followed by Album. The lowest level is the track itself, which is not configured explicitly.

  - `log_dir`
  
    Log directory. Default is `/var/log/muserv`. The muserv system user must have write access to it.

  - `log_level`  

    Here, the verbosity of the muserv log can be configured. Possible values are (ordered by increasing verbosity): `panic`, `fatal`, `error`, `warn`, `info`, `debug`, `trace`. Default is `fatal`.