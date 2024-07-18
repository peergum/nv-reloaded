obj-m := EPD_IT8951.o

#ccflags-m := -I$(PWD)/../Config

all:
	make -C /lib/modules/$(shell uname -r)/build M=$(PWD) modules

clean:
	make -C /lib/modules/$(shell uname -r)/build M=$(PWD) clean
