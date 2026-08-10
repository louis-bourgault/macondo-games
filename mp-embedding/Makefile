SHELL := /bin/sh

ROOT := $(CURDIR)
MICROPYTHON_ROOT ?= $(HOME)/src/micropython
MP_TOP := $(MICROPYTHON_ROOT)
EMBED_SRC_DIR := $(MICROPYTHON_ROOT)/examples/embedding
EMBED_DIR := $(EMBED_SRC_DIR)/micropython_embed
BUILD_DIR := $(ROOT)/build
GO_DIR := $(ROOT)/bridge
HOST_DIR := $(ROOT)/host

CC ?= cc
GO ?= go
AR ?= ar

PROG := $(ROOT)/spike
GO_ARCHIVE := $(GO_DIR)/libbridge.a
HOST_OBJ := $(BUILD_DIR)/host/main.o
MP_BUILD_STAMP := $(BUILD_DIR)/micropython.stamp

CFLAGS += -I$(ROOT)
CFLAGS += -I$(EMBED_DIR)
CFLAGS += -I$(EMBED_DIR)/port
CFLAGS += -I$(EMBED_SRC_DIR)
CFLAGS += -I$(MP_TOP)
CFLAGS += -Wall -Wextra -Og -fno-common

.PHONY: all clean embed go-bridge host micropython

all: $(PROG)

embed:
	$(MAKE) -C $(EMBED_SRC_DIR) -f micropython_embed.mk MICROPYTHON_TOP=../..

go-bridge: $(GO_ARCHIVE)

$(GO_ARCHIVE): $(GO_DIR)/main.go
	cd $(GO_DIR) && $(GO) build -buildmode=c-archive -o libbridge.a main.go

$(GO_DIR)/libbridge.h: $(GO_ARCHIVE)

micropython: $(MP_BUILD_STAMP)

$(BUILD_DIR)/host/main.o: $(HOST_DIR)/main.c $(GO_ARCHIVE) | embed
	@mkdir -p $(dir $@)
	$(CC) $(CFLAGS) -c $< -o $@

$(MP_BUILD_STAMP): $(EMBED_SRC_DIR)/mpconfigport.h
	@mkdir -p $(BUILD_DIR)/micropython
	@set -e; \
	$(MAKE) embed; \
	find $(EMBED_DIR) -name '*.c' | sort | while read src; do \
		obj=$(BUILD_DIR)/$${src#$(EMBED_DIR)/}; \
		obj=$${obj%.c}.o; \
		mkdir -p $$(dirname "$$obj"); \
		$(CC) $(CFLAGS) -c "$$src" -o "$$obj"; \
	done; \
	touch $@

$(PROG): $(HOST_OBJ) $(GO_ARCHIVE) $(MP_BUILD_STAMP)
	@objs="$$(find $(BUILD_DIR) -name '*.o' ! -path '$(BUILD_DIR)/host/*' | sort)"; \
	$(CC) -o $@ $(HOST_OBJ) $$objs -Wl,--whole-archive $(GO_ARCHIVE) -Wl,--no-whole-archive -lm -lpthread

clean:
	rm -rf $(BUILD_DIR) $(PROG)
	@if [ -d "$(EMBED_SRC_DIR)" ]; then $(MAKE) -C $(EMBED_SRC_DIR) -f micropython_embed.mk clean MICROPYTHON_TOP=../..; fi
	cd $(GO_DIR) && rm -f libbridge.a libbridge.h