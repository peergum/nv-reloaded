all: clean fontconvert fontconvert2go importfonts

CC     = gcc
CFLAGS = -Wall \
	-I/usr/include/freetype2 -L/usr/lib/aarch64-linux-gnu -L/usr/local/lib -lfreetype \
    -I/opt/homebrew/include/freetype2 -L/opt/homebrew/lib
LIBS   = -lfreetype

fontconvert2go: fontconvert2go.c
	$(CC) $(CFLAGS) $< $(LIBS) -o $@
	strip $@

fontconvert: fontconvert.c
	$(CC) $(CFLAGS) $< $(LIBS) -o $@
	strip $@

importfonts:
	rm -fr ../Fonts ../fonts-go
	./convert.sh

clean:
	rm -f fontconvert fontconvert2go
