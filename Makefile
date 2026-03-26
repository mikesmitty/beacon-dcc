TINYGO ?= tinygo
GOTOOLCHAIN ?= go1.25.8
TARGET ?= metro-rp2350
SHIELD ?= EX-MotorShield8874
SERIAL ?= usb

GIT := $(shell git rev-parse --short HEAD)
VER := $(shell git describe --tags 2>/dev/null || echo v0.0.0)

# git describe --tags
# <tag>-<commit count>-g<commit hash>
# e.g. v0.35.0-235-gdeb48dd5

LDFLAGS := -X main.board='$(TARGET)' -X main.gitSHA='$(GIT)' -X main.shieldName='$(SHIELD)' -X main.version='$(VER)'

.PHONY: flash gdb build uf2

flash:
	GOTOOLCHAIN=$(GOTOOLCHAIN) $(TINYGO) flash \
		-target=$(TARGET) \
		-serial=$(SERIAL) \
		-ldflags="$(LDFLAGS)"

gdb:
	GOTOOLCHAIN=$(GOTOOLCHAIN) $(TINYGO) gdb \
		-target=$(TARGET) \
		-serial=$(SERIAL) \
		-ldflags="$(LDFLAGS)"

build:
	GOTOOLCHAIN=$(GOTOOLCHAIN) $(TINYGO) build \
		-target=$(TARGET) \
		-serial=$(SERIAL) \
		-ldflags="$(LDFLAGS)"

uf2:
	GOTOOLCHAIN=$(GOTOOLCHAIN) $(TINYGO) build \
		-target=$(TARGET) \
		-serial=$(SERIAL) \
		-ldflags="$(LDFLAGS)" \
		-o beacon-dcc.uf2