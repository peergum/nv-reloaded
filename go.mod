module nv

go 1.22.2

require (
	github.com/fsnotify/fsnotify v1.7.0
	github.com/peergum/IT8951-go v0.0.7
	github.com/peergum/pi-sugar v0.0.6
)

require (
	github.com/peergum/go-rpio/v5 v5.0.3 // indirect
	golang.org/x/sys v0.4.0 // indirect
)

//replace github.com/peergum/pi-sugar => ./power/pisugar
//replace github.com/peergum/go-rpio/v5 => ./go-rpio
