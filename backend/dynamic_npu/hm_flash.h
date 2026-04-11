#ifndef HM_FLASH_H
#define HM_FLASH_H

#include <stdint.h>

int hm_flash_read(int dev_index, uint32_t offset, void *buf, uint32_t len);
int hm_flash_program(int dev_index, uint32_t offset, const void *buf, uint32_t len);

#endif // HM_FLASH_H
