# Code and main assets for our face detection demo snap

The face detection demo snap enables people to show multiple faces of snaps. It comprehends:
* a face-detection-service, which can:
  * toggle on/off face detection (a webcam is required)
  * collect stats over time and store it in a sqlite database
  * serve via a webserver (on http://IP:8080) those results in a single page app, with graph history, last webcam screenshot, last image with detected faces circled (note that the html/css/javascript code is in another repo)
  * use websocket to connect multiple clients, and refresh data to each web page without needing to reload it
  * enable "fun" mode where detected faces circle are replaced by distribution logo attributed randomly
* a face-detection-cli tool, which can:
  * enable/disable face detection webcam (the webserver will still be served though). No new data is collected when face detection is disabled
  * toggle between normal/fun rendering mode
  * quit the service

## Update and revert

The application is buggy on purpose with version **2.0**. The data from previous run will be destroyed (and the web page data refresh to reflects this) and it instructs the web page to turn RED.
This enables to illustrate the `snap revert` functionality where the previous version will be restored (service restarted) as well as previous data which will be repopulated on the web page.

## Technical details

### snapcraft.yaml

A snapcraft.yaml is provided which demonstrates multiple features!
 * building a golang app
 * shipping a service and a cli tool
 * copying local assets
 * referencing other repository (the web code)
 * give some network-related security permission
 * have different security configuration per application

### generated files

This service generates some files available in `$SNAP_DATA` (root project directory if ran from master without this variable set):
 * configuration (saved by the service for persistency over restart) in `settings`
 * sqlite database contentstorage main data in `storage.db`
 * `screencapture.png` and `screendetected.png` for latest captured images.
