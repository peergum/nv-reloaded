all: sync run

NV_ARGS = -d -di -dc -dd -nl -dio -epd
TEST_ARGS = -d -dd -dc -di -dsugar -dio
PI4 = -m pi4
PI0_2W = -m pizero2w

#sourceFontList = display/ImportedFonts/*.ttf display/ImportedFonts/*.otf
#generatedFontList = $(patsubst display/ImportedFonts/%, %, sourceFontList)
#generatedGoFonts = $(patsubst %.*, %.go, generatedFontList)

go:
	sh ./install_go

sync: clean
	./sync

build: install_nv

clean:
	rm -f nv

nv:
	go build -mod=mod

install_nv: nv
	if [ "`hostname`" != "Sparta.local" ]; then \
		sudo install -m a=rxs -p -o root -g i2c  nv /usr/local/bin/ ;\
		sudo mkdir -p /var/nv ;\
		sudo cp nv.service /etc/systemd/system/ ;\
	fi

start: install_nv
	sudo systemctl daemon-reload
	sudo systemctl enable nv
	sudo systemctl restart nv

stop:
	sudo systemctl stop nv

run: stop install_nv
	nv $(NV_ARGS) ${`hostname`=="pi4"?$PI4:$PI0_2W}

test: stop install_nv
	nv $(TEST_ARGS) ${`hostname`=="pi4"?$PI4:$PI0_2W}

debug: sync nv
	sudo gdb --args nv $(TEST_ARGS) ${`hostname`=="pi4"?$PI4:$PI0_2W}

fonts: display/fontconvert
	if [ -f display/fontconvert/fontconvert2go ]; then \
  		rm display/fontconvert/fontconvert2go; \
  	fi
	make -C display/fontconvert all

