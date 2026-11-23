#ifndef FS
#define FS
#include "lfs.h"

extern lfs_t lfs;
int fs_init(void);
#define LFS_CACHE_SIZE 256

#endif