all: sync run

#sourceFontList = display/ImportedFonts/*.ttf display/ImportedFonts/*.otf
#generatedFontList = $(patsubst display/ImportedFonts/%, %, sourceFontList)
#generatedGoFonts = $(patsubst %.*, %.go, generatedFontList)

sync: clean
	./sync

build: sync install

clean:
	rm -f nv

nv:
	go build

install: nv
	if [ "`hostname`" = "pi4" ]; then \
		sudo install -m 4755 -p -o root -g root  nv /usr/local/bin/ ;\
		sudo mkdir -p /var/nv ;\
		sudo cp nv.service /etc/systemd/system/ ;\
	fi

start: install
	sudo systemctl daemon-reload
	sudo systemctl enable nv
	sudo systemctl restart nv

stop:
	sudo systemctl stop nv

run: stop install
	nv -d -dd -dc -di

debug: nv
	sudo gdb --args nv -d -dd -dc -di -nl -nw

fonts: display/fontconvert/fontconvert2go
	if [ -f display/fontconvert/fontconvert2go ]; then \
  		rm display/fontconvert/fontconvert2go; \
  	fi
	make -C display/fontconvert all

