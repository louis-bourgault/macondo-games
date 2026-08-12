# Regenerate the embed package's genhdr (qstrs, root pointers, module defs).
#
# The bare embed generator scans only py/ sources, so MP_QSTR_* / root pointers
# / module registrations from vfs.c, modos.c, pyexec.c and readline.c would be
# missing (link errors: undefined MP_QSTR_mount, missing vfs_mount_table root
# pointer, unknown `os` module). Running this make inside the generated
# micropython_embed directory re-scans ALL sources in the package before the
# firmware archive is compiled.
#
# Invoked by the justfile `embed` recipe, after the extra sources are copied in.
MICROPYTHON_TOP ?= ../..

BUILD := build-local

include $(MICROPYTHON_TOP)/py/mkenv.mk

CFLAGS += -I. -Iport -I$(TOP) -I$(BUILD)

include $(MICROPYTHON_TOP)/py/py.mk

# Extra sources the bare generator omitted but our firmware compiles.
# NOTE: must come BEFORE mkrules.mk — GNU make expands rule prerequisites
# immediately, so the qstr.i.last rule must see SRC_QSTR including these files.
SRC_QSTR += $(wildcard extmod/*.c) $(wildcard shared/runtime/*.c) shared/readline/readline.c

include $(MICROPYTHON_TOP)/py/mkrules.mk

# mpconfigport.h lives in port/ (QSTR_GLOBAL_DEPENDENCIES references it as a
# bare filename); mkrules' $(HEADER_BUILD) rule creates build-local/genhdr.
vpath mpconfigport.h port

.PHONY: genhdr
genhdr: $(addprefix $(BUILD)/genhdr/, mpversion.h qstrdefs.generated.h moduledefs.h root_pointers.h)