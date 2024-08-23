all: sync run

NV_ARGS = -d -di -dc -dd -nl
TEST_ARGS = -d -dd -dc -di -dsugar -dio
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
	go build

install_nv: nv
	if [ "`hostname`" = "pi4" ]; then \
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
	nv $(NV_ARGS)

test: stop install_nv
	nv $(TEST_ARGS)

debug: sync nv
	sudo gdb --args nv $(TEST_ARGS)

fonts: display/fontconvert
	if [ -f display/fontconvert/fontconvert2go ]; then \
  		rm display/fontconvert/fontconvert2go; \
  	fi
	make -C display/fontconvert all

