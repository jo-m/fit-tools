# FIT tools

Some simple tools to handle Garmin `.fit` and `.gpx` files.

- [fitsort](fitsort/main.go) — Sort `.fit` files by timestamp. `go install jo-m.ch/go/fit-tools/fitsort@main` (based on https://github.com/muktihari/fit)
- [gpxrename](gpxrename.py) — Rename `.gpx` files based on the track name in their metadata. `./gpxrename.py *.gpx`
