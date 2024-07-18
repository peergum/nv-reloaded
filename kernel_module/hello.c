#include <linux/module.h>
#include <linux/kernel.h>
#include <linux/types.h>

#include "EPD_IT8951.h"

MODULE_LICENSE("GPLv3");

int hello_init(void) {
    pr_info("Hello World :)\n");
    IT8951_Dev_Info info = EPD_IT8951_Init(-2150);
    pr_info("WxH=%u,$u", info.Panel_W, info.Panel_H);
    return 0;
}

void hello_exit(void) {
    pr_info("Goodbye World!\n");
}

module_init(hello_init);
module_exit(hello_exit);