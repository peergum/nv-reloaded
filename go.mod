module nv

go 1.22.6

toolchain go1.22.7

require (
	github.com/fsnotify/fsnotify v1.7.0
	github.com/peergum/IT8951-go v1.0.0
	github.com/peergum/pi-sugar v0.0.6
	periph.io/x/conn/v3 v3.7.1
	periph.io/x/host/v3 v3.8.2
)

require (
	github.com/peergum/go-rpio/v5 v5.0.3 // indirect
	golang.org/x/sys v0.4.0 // indirect
)

replace github.com/peergum/pi-sugar => ./power/pi-sugar

replace github.com/peergum/go-rpio/v5 => ./go-rpio

replace github.com/peergum/IT8951-go => ./it8951

replace periph.io/x/host/v3 => ./periph.io/host
