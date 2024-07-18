obj-m := hello.o
subdir-m := IT8951/e-Paper IT8951/Config

#CFLAGS = -I $(PWD)/IT8951/Config -I $(PWD)/IT8951/e-Paper

all:
	make -C /lib/modules/$(shell uname -r)/build M=$(PWD) modules

clean:
	make -C /lib/modules/$(shell uname -r)/build M=$(PWD) clean
