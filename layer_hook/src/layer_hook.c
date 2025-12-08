#define _GNU_SOURCE
#include <dlfcn.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/mman.h>
#include <sys/stat.h>
#include <wayland-client-core.h>
#include "wayland-client-protocol.h"
#include "wlr-layer-shell-unstable-v1-client-protocol.h"

#include "wayland-client-protocol-code.h"

// Stub for xdg_popup_interface (referenced by layer-shell protocol but unused)
const struct wl_interface xdg_popup_interface = { "xdg_popup", 1, 0, NULL, 0, NULL };

#define MMAP_PATH "/tmp/kittybg.mmap"
#define MAX_PANELS 64
#define DEBUG 0

#if DEBUG
#define DEBUG_PRINT(fmt, ...) fprintf(stderr, "[layer_hook] " fmt "\n", ##__VA_ARGS__)
#else
#define DEBUG_PRINT(fmt, ...) do {} while(0)
#endif

// Shared memory structure for keyboard mode control
struct panel_entry {
    int32_t pid;      // 0 = empty slot
    uint8_t mode;     // 0=NONE, 1=EXCLUSIVE, 2=ON_DEMAND
    uint8_t _pad[3];  // Alignment padding
};

struct keyboard_state {
    uint64_t version;                     // Bumped on any write
    struct panel_entry panels[MAX_PANELS];
};

// Function pointer types for wayland-client functions
typedef int (* PFN_wl_display_flush)(struct wl_display *display);
typedef int (* PFN_wl_display_dispatch_pending)(struct wl_display *display);
typedef struct wl_proxy* (* PFN_wl_proxy_marshal_array_flags)(struct wl_proxy*, uint32_t, const struct wl_interface*, uint32_t, uint32_t, union wl_argument*);
typedef uint32_t (* PFN_wl_proxy_get_version)(struct wl_proxy*);

static struct wl_display *g_display = NULL;
static struct zwlr_layer_surface_v1 *g_layer_surface = NULL;
static struct wl_surface *g_wl_surface = NULL;

// mmap state
static struct keyboard_state *g_mmap_ptr = NULL;
static int g_mmap_fd = -1;
static uint64_t g_last_version = 0;
static int32_t g_current_mode = -1;  // Track current mode to avoid redundant applies
static pid_t g_my_pid = 0;

// Global flag: 0 = hooks disabled, 1 = hooks enabled
static int g_hooks_enabled = 0;

// Recursion guard for mode changes
static int g_in_mode_change = 0;

// Dynamically loaded wayland-client functions
static struct {
    void* handle;
    PFN_wl_display_flush display_flush;
    PFN_wl_proxy_marshal_array_flags proxy_marshal_array_flags;
    PFN_wl_proxy_get_version proxy_get_version;
} wl_client;

static int load_wayland_client(void) {
    if (wl_client.handle)
        return 0;

    wl_client.handle = dlopen("libwayland-client.so.0", RTLD_LAZY | RTLD_LOCAL);
    if (!wl_client.handle) {
        DEBUG_PRINT("Failed to dlopen libwayland-client.so.0: %s", dlerror());
        return -1;
    }

    wl_client.display_flush = dlsym(wl_client.handle, "wl_display_flush");
    wl_client.proxy_marshal_array_flags = dlsym(wl_client.handle, "wl_proxy_marshal_array_flags");
    wl_client.proxy_get_version = dlsym(wl_client.handle, "wl_proxy_get_version");

    if (!wl_client.display_flush || !wl_client.proxy_marshal_array_flags || !wl_client.proxy_get_version) {
        DEBUG_PRINT("Failed to load wayland-client entry points");
        dlclose(wl_client.handle);
        wl_client.handle = NULL;
        return -1;
    }

    DEBUG_PRINT("Loaded wayland-client functions");
    return 0;
}

static int open_mmap(void) {
    if (g_mmap_ptr)
        return 0;

    g_mmap_fd = open(MMAP_PATH, O_RDONLY);
    if (g_mmap_fd < 0) {
        // File doesn't exist yet - that's OK, will try again later
        return -1;
    }

    g_mmap_ptr = mmap(NULL, sizeof(struct keyboard_state), PROT_READ, MAP_SHARED, g_mmap_fd, 0);
    if (g_mmap_ptr == MAP_FAILED) {
        DEBUG_PRINT("Failed to mmap %s", MMAP_PATH);
        close(g_mmap_fd);
        g_mmap_fd = -1;
        g_mmap_ptr = NULL;
        return -1;
    }

    DEBUG_PRINT("Opened mmap file %s (pid=%d)", MMAP_PATH, g_my_pid);
    return 0;
}

// Check mmap for mode change, returns new mode or -1 if no change
static int check_mmap_mode(void) {
    // Try to open mmap if not already open
    if (!g_mmap_ptr) {
        if (open_mmap() != 0)
            return -1;
    }

    // Check if version changed
    uint64_t current_version = g_mmap_ptr->version;
    if (current_version == g_last_version)
        return -1;

    g_last_version = current_version;

    // Scan for our PID
    for (int i = 0; i < MAX_PANELS; i++) {
        if (g_mmap_ptr->panels[i].pid == g_my_pid) {
            uint8_t mode = g_mmap_ptr->panels[i].mode;
            if (mode <= 2 && mode != g_current_mode) {
                DEBUG_PRINT("mmap: found mode %u for pid %d (was %d)", mode, g_my_pid, g_current_mode);
                g_current_mode = mode;
                return mode;
            }
            return -1;  // Mode unchanged
        }
    }

    return -1;  // PID not found
}

__attribute__((constructor))
static void init(void) {
    g_my_pid = getpid();

    // Check if we're running in kitty/kitten process
    char exe[256] = {0};
    ssize_t len = readlink("/proc/self/exe", exe, sizeof(exe) - 1);
    if (len > 0) {
        exe[len] = '\0';
        DEBUG_PRINT("Library loaded into process: %s (pid=%d)", exe, g_my_pid);

        if (strstr(exe, "kitty") != NULL || strstr(exe, "kitten") != NULL) {
            DEBUG_PRINT("Kitty process detected, enabling hooks");
            g_hooks_enabled = 1;
        } else {
            DEBUG_PRINT("Not a kitty process, disabling hooks and clearing LD_PRELOAD");
            g_hooks_enabled = 0;
            unsetenv("LD_PRELOAD");
            return;
        }
    }
}

static PFN_wl_proxy_marshal_array_flags original_wl_proxy_marshal_array_flags = NULL;

// Provide wl_proxy_get_version for inline protocol functions
uint32_t wl_proxy_get_version(struct wl_proxy *proxy) {
    static PFN_wl_proxy_get_version original = NULL;
    if (!original) {
        original = dlsym(RTLD_NEXT, "wl_proxy_get_version");
        if (!original) {
            void *wl_handle = dlopen("libwayland-client.so.0", RTLD_LAZY | RTLD_NOLOAD);
            if (wl_handle) {
                original = dlsym(wl_handle, "wl_proxy_get_version");
            }
        }
    }
    return original ? original(proxy) : 0;
}

// Direct wayland protocol calls using wl_argument arrays
#define LAYER_SURFACE_SET_KEYBOARD_INTERACTIVITY 4
#define SURFACE_COMMIT 6

static void apply_keyboard_mode(uint32_t mode) {
    if (!g_layer_surface || !g_wl_surface || !wl_client.proxy_marshal_array_flags) {
        DEBUG_PRINT("Cannot apply mode: missing layer_surface=%p, wl_surface=%p, marshal_fn=%p",
            (void*)g_layer_surface, (void*)g_wl_surface, (void*)wl_client.proxy_marshal_array_flags);
        return;
    }

    // Set keyboard interactivity: opcode 4, 1 arg (uint32_t mode)
    union wl_argument args_mode[1];
    args_mode[0].u = mode;
    wl_client.proxy_marshal_array_flags(
        (struct wl_proxy *)g_layer_surface,
        LAYER_SURFACE_SET_KEYBOARD_INTERACTIVITY,
        NULL,
        wl_client.proxy_get_version((struct wl_proxy *)g_layer_surface),
        0,
        args_mode
    );

    // Commit surface: opcode 6, no args
    wl_client.proxy_marshal_array_flags(
        (struct wl_proxy *)g_wl_surface,
        SURFACE_COMMIT,
        NULL,
        wl_client.proxy_get_version((struct wl_proxy *)g_wl_surface),
        0,
        NULL
    );

    DEBUG_PRINT("Applied keyboard mode %u", mode);
}

struct wl_proxy *wl_proxy_marshal_array_flags(
    struct wl_proxy *proxy, uint32_t opcode,
    const struct wl_interface *interface, uint32_t version,
    uint32_t flags, union wl_argument *args) {

    if (!original_wl_proxy_marshal_array_flags) {
        original_wl_proxy_marshal_array_flags = dlsym(RTLD_NEXT, "wl_proxy_marshal_array_flags");
        if (!original_wl_proxy_marshal_array_flags) {
            void *wl_handle = dlopen("libwayland-client.so.0", RTLD_LAZY | RTLD_NOLOAD);
            if (wl_handle) {
                original_wl_proxy_marshal_array_flags = dlsym(wl_handle, "wl_proxy_marshal_array_flags");
            }
            if (!original_wl_proxy_marshal_array_flags) {
                DEBUG_PRINT("FATAL: Failed to find original wl_proxy_marshal_array_flags");
                return NULL;
            }
        }
        DEBUG_PRINT("Found original wl_proxy_marshal_array_flags at %p", (void *)original_wl_proxy_marshal_array_flags);
    }

    // If hooks disabled or in mode change, pass through immediately
    if (!g_hooks_enabled || g_in_mode_change) {
        return original_wl_proxy_marshal_array_flags(proxy, opcode, interface, version, flags, args);
    }

    // Check mmap for mode change
    if (g_layer_surface && g_wl_surface) {
        int new_mode = check_mmap_mode();
        if (new_mode >= 0) {
            g_in_mode_change = 1;
            apply_keyboard_mode((uint32_t)new_mode);
            g_in_mode_change = 0;
        }
    }

    // Call original function
    struct wl_proxy *result = original_wl_proxy_marshal_array_flags(
        proxy, opcode, interface, version, flags, args);

    // Intercept layer surface creation
    if (result && interface && interface->name) {
        if (strcmp(interface->name, "zwlr_layer_surface_v1") == 0) {
            DEBUG_PRINT("Intercepted layer surface creation");
            g_layer_surface = (struct zwlr_layer_surface_v1 *)result;
            if (args && args[1].o) {
                g_wl_surface = (struct wl_surface *)args[1].o;
                DEBUG_PRINT("Captured wl_surface: %p", (void *)g_wl_surface);
            }
            DEBUG_PRINT("Stored layer surface: %p", (void *)g_layer_surface);
        }
    }

    return result;
}

struct wl_display *wl_display_connect(const char *name) {
    static struct wl_display *(*original_wl_display_connect)(const char *) = NULL;

    if (!original_wl_display_connect) {
        original_wl_display_connect = dlsym(RTLD_NEXT, "wl_display_connect");
        if (!original_wl_display_connect) {
            void *wl_handle = dlopen("libwayland-client.so.0", RTLD_LAZY | RTLD_NOLOAD);
            if (wl_handle) {
                original_wl_display_connect = dlsym(wl_handle, "wl_display_connect");
            }
            if (!original_wl_display_connect) {
                DEBUG_PRINT("FATAL: Failed to find original wl_display_connect");
                return NULL;
            }
        }
        DEBUG_PRINT("Found original wl_display_connect at %p", (void *)original_wl_display_connect);
    }

    if (!g_hooks_enabled) {
        return original_wl_display_connect(name);
    }

    struct wl_display *display = original_wl_display_connect(name);

    if (display) {
        if (load_wayland_client() != 0) {
            DEBUG_PRINT("Failed to load wayland-client, functionality will be limited");
            return display;
        }

        g_display = display;
        DEBUG_PRINT("Captured display connection: %p", (void *)display);

        // Try to open mmap (may not exist yet)
        open_mmap();

        // Note: LD_PRELOAD is NOT cleared here - we need it to propagate
        // to kitty server if kitten spawns it. Non-kitty processes clear
        // it in the constructor based on exe name check.
    }

    return display;
}

// Hook wl_display_flush to check mmap
static PFN_wl_display_flush original_wl_display_flush = NULL;

int wl_display_flush(struct wl_display *display) {
    if (!original_wl_display_flush) {
        original_wl_display_flush = dlsym(RTLD_NEXT, "wl_display_flush");
        if (!original_wl_display_flush) {
            void *wl_handle = dlopen("libwayland-client.so.0", RTLD_LAZY | RTLD_NOLOAD);
            if (wl_handle) {
                original_wl_display_flush = dlsym(wl_handle, "wl_display_flush");
            }
            if (!original_wl_display_flush) {
                DEBUG_PRINT("FATAL: Failed to find original wl_display_flush");
                return -1;
            }
        }
    }

    if (!g_hooks_enabled) {
        return original_wl_display_flush(display);
    }

    // Check mmap for mode change
    if (g_layer_surface && g_wl_surface && !g_in_mode_change) {
        int new_mode = check_mmap_mode();
        if (new_mode >= 0) {
            g_in_mode_change = 1;
            apply_keyboard_mode((uint32_t)new_mode);
            g_in_mode_change = 0;
        }
    }

    return original_wl_display_flush(display);
}

// Hook wl_display_dispatch_pending to check mmap during event loop
static PFN_wl_display_dispatch_pending original_wl_display_dispatch_pending = NULL;

int wl_display_dispatch_pending(struct wl_display *display) {
    if (!original_wl_display_dispatch_pending) {
        original_wl_display_dispatch_pending = dlsym(RTLD_NEXT, "wl_display_dispatch_pending");
        if (!original_wl_display_dispatch_pending) {
            void *wl_handle = dlopen("libwayland-client.so.0", RTLD_LAZY | RTLD_NOLOAD);
            if (wl_handle) {
                original_wl_display_dispatch_pending = dlsym(wl_handle, "wl_display_dispatch_pending");
            }
            if (!original_wl_display_dispatch_pending) {
                DEBUG_PRINT("FATAL: Failed to find original wl_display_dispatch_pending");
                return -1;
            }
        }
    }

    if (!g_hooks_enabled) {
        return original_wl_display_dispatch_pending(display);
    }

    // Check mmap for mode change
    if (g_layer_surface && g_wl_surface && !g_in_mode_change) {
        int new_mode = check_mmap_mode();
        if (new_mode >= 0) {
            g_in_mode_change = 1;
            apply_keyboard_mode((uint32_t)new_mode);
            g_in_mode_change = 0;

            // Flush immediately
            if (wl_client.display_flush) {
                wl_client.display_flush(display);
            }
        }
    }

    return original_wl_display_dispatch_pending(display);
}

__attribute__((destructor))
static void cleanup(void) {
    DEBUG_PRINT("Cleaning up layer hook");

    if (g_mmap_ptr && g_mmap_ptr != MAP_FAILED) {
        munmap(g_mmap_ptr, sizeof(struct keyboard_state));
        g_mmap_ptr = NULL;
    }

    if (g_mmap_fd >= 0) {
        close(g_mmap_fd);
        g_mmap_fd = -1;
    }

    if (wl_client.handle) {
        dlclose(wl_client.handle);
        wl_client.handle = NULL;
    }
}
