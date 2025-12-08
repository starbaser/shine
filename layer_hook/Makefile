CC = gcc
CFLAGS = -Wall -Wextra -fPIC -O2 -I/usr/include
LDFLAGS = -shared -ldl -L/usr/lib -lwayland-client

BUILD_DIR = build
SRC_DIR = src
PROTO_DIR = $(SRC_DIR)/wayland-protocols

WAYLAND_XML = $(PROTO_DIR)/wayland.xml
WAYLAND_HEADER = $(PROTO_DIR)/wayland-client-protocol.h
WAYLAND_CODE = $(PROTO_DIR)/wayland-client-protocol-code.h

LAYER_XML = $(PROTO_DIR)/wlr-layer-shell-unstable-v1.xml
LAYER_HEADER = $(PROTO_DIR)/wlr-layer-shell-unstable-v1-client-protocol.h
LAYER_CODE = $(PROTO_DIR)/wlr-layer-shell-unstable-v1-protocol.c

PROTOCOL_HEADERS = $(WAYLAND_HEADER) $(LAYER_HEADER)
PROTOCOL_CODE = $(WAYLAND_CODE) $(LAYER_CODE)

OUTPUT_LIB = $(BUILD_DIR)/libshine-layer.so

.PHONY: all clean distclean protocols

all: $(OUTPUT_LIB)

protocols: $(PROTOCOL_HEADERS) $(PROTOCOL_CODE)

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

$(WAYLAND_HEADER): $(WAYLAND_XML)
	wayland-scanner client-header $< $@

$(WAYLAND_CODE): $(WAYLAND_XML)
	wayland-scanner private-code $< $@

$(LAYER_HEADER): $(LAYER_XML)
	wayland-scanner client-header $< $@

$(LAYER_CODE): $(LAYER_XML)
	wayland-scanner private-code $< $@

$(OUTPUT_LIB): $(SRC_DIR)/layer_hook.c $(PROTOCOL_HEADERS) $(PROTOCOL_CODE) | $(BUILD_DIR)
	$(CC) $(CFLAGS) -I$(PROTO_DIR) -o $@ $< $(LAYER_CODE) $(LDFLAGS)

clean:
	rm -rf $(BUILD_DIR)

distclean: clean
	rm -f $(PROTOCOL_HEADERS) $(PROTOCOL_CODE)
