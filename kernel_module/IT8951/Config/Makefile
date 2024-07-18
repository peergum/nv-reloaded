obj-m := DEV_Config.o
#ifdef (BCM)
#	obj-m += dev_hardware_SPI.o RPI_gpiod.o
#endif
subdir-m := IT8951/e-Paper IT8951/Config
ccflags-m := -DBCM
ldflags-m := -lbcm

all:
	make -C /lib/modules/$(shell uname -r)/build M=$(PWD) modules

clean:
	make -C /lib/modules/$(shell uname -r)/build M=$(PWD) clean
