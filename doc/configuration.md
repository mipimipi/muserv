# Configuration settings

muserv provides default configuration settings. The only additional configuration that must be made for the music directory. 

Below we describe all configuration settings.

* `cnt`

  Content-related parameters

  - `music_dir`

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

    Here, the content hierarchies that are shown in the UPnP clients are configured. Possible hierarchies are:
    
    - Genre -> AlbumArtist -> Album -> Track
    - Genre -> Artist -> Track
    - Genre -> Album -> Track
    - Genre -> Track
    - AlbumArtist -> Album -> Track
    - Artist -> Track
    - Track

    UPnP clients show the hierarchies in the same sequence as they are configured here. Some hierarachies are preconfigured. Just adjust or remove them or add additional hierarchies. Each hierarchy needs a name. That's the name that is also displayed by the clients. The name of the preconfigured hierarchies can be adjusted. Hierarchies are configured as list of levels (in the first hierarchy "Genre" represents one level, for example). For each level two configurations must be made:

    1. `type` represents the object or tag (`genre` or `track`, for example). The type of the last level of each hierarchy must by `track`.
    1. `sort` are the sorting criteria. They define how the data is sorted inside that level. It consists of a list of attributes preceded by the character `+` or `-` which defines if the sort order is ascending or descending for that attribute. Albums can be sorted by the attributes `title`, `year` and `lastChange`, tracks by the attributes `title`, `year`, `trackNo`, `discNo` and `lastChange`. For all other types (`genre`, `albumartist`, `artist`) no attributes are supported. These are just sorted by the content of the coresponding tag.

    Example (latest albums by genre):

            {
                "name": "Latest Albums by Genre",
                "levels": [
                    {
                        "type": "genre",
                        "sort": ["+"]
                    },
                    {
                        "type": "album",
                        "sort": ["-lastChange"]
                    },
                    {
                        "type": "track",
                        "sort": ["+discNo","+trackNo"]
                    }
                ]
            },

    Note that within an album, tracks are sorted first by disc number and then by track number. With this configuration, albums with multiple discs can be handled.          

  - `show_folders`

    Whether the folder hierarchy shall be shown or not. If it shall be shown, it's the last hierarchy sequence of hierarchies that UPnP clients display.

  - `folder_hierarchy_name`

    The name of the folder hierarchy that is shown by UPnP clients.

  - `log_dir`
  
    Log directory. Default is `/var/log/muserv`. The muserv system user must have write access to it.

  - `log_level`  

    Here, the verbosity of the muserv log can be configured. Possible values are (ordered by increasing verbosity): `panic`, `fatal`, `error`, `warn`, `info`, `debug`, `trace`. Default is `fatal`.
