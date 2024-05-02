all: sync run

#sourceFontList = display/ImportedFonts/*.ttf display/ImportedFonts/*.otf
#generatedFontList = $(patsubst display/ImportedFonts/%, %, sourceFontList)
#generatedGoFonts = $(patsubst %.*, %.go, generatedFontList)

sync: clean
	./sync

clean:
	rm -f nv

nv:
	go build

install: nv
	if [ "`hostname`" = "pi4" ]; then \
		sudo chown root nv ;\
		sudo chmod u+s nv ;\
		sudo cp nv.service /etc/systemd/system/ ;\
	fi

start: install
	sudo systemctl daemon-reload
	sudo systemctl restart nv

run: install
	sudo systemctl stop nv
	./nv -debug -epd -doc -display -input

debug: nv
	sudo gdb --args ./nv -debug -epd -doc -display

fonts: display/fontconvert/fontconvert2go
	if [ -f display/fontconvert/fontconvert2go ]; then rm display/fontconvert/fontconvert2go; fi
	make -C display/fontconvert all

