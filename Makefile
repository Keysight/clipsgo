GO					?= go
CLIPS_VERSION		?= 6.31
CLIPS_SOURCE_URL	?= "https://downloads.sourceforge.net/project/clipsrules/CLIPS/6.31/clips_core_source_631.zip"
MAKEFILE_NAME		?= makefile
SHARED_INCLUDE_DIR	?= /usr/local/include
SHARED_LIBRARY_DIR	?= /usr/local/lib

# platform detection
PLATFORM = $(shell uname -s)

.PHONY: clips clipsgo test install-clips clean

all: clips_source clips clipsgo

clips_source:
	wget -O clips.zip $(CLIPS_SOURCE_URL)
	unzip -jo clips.zip -d clips_source

ifeq ($(PLATFORM),Darwin) # macOS
clips: clips_source
	$(MAKE) -f $(MAKEFILE_NAME) -C clips_source                            \
		CFLAGS="-std=c99 -O3 -fno-strict-aliasing -fPIC"               \
		LDLIBS="-lm"
	ld clips_source/*.o -lm -dylib -arch x86_64                            \
		-o clips_source/libclips.so
else
clips: clips_source
	$(MAKE) -f $(MAKEFILE_NAME) -C clips_source                            \
		CFLAGS="-std=c99 -O3 -fno-strict-aliasing -fPIC"               \
		LDLIBS="-lm -lrt"
	ld -G clips_source/*.o -o clips_source/libclips.so
endif

clipsgo: clips
	$(GO) build ./pkg/...

test: clipsgo
	$(GO) test ./pkg/...

install-clips: clips
	install -d $(SHARED_INCLUDE_DIR)/
	install -m 644 clips_source/clips.h $(SHARED_INCLUDE_DIR)/
	install -d $(SHARED_INCLUDE_DIR)/clips
	install -m 644 clips_source/*.h $(SHARED_INCLUDE_DIR)/clips/
	install -d $(SHARED_LIBRARY_DIR)/
	install -m 644 clips_source/libclips.so                                \
	 	$(SHARED_LIBRARY_DIR)/libclips.so.$(CLIPS_VERSION)
	ln -s $(SHARED_LIBRARY_DIR)/libclips.so.$(CLIPS_VERSION)	       \
	 	$(SHARED_LIBRARY_DIR)/libclips.so.6
	ln -s $(SHARED_LIBRARY_DIR)/libclips.so.$(CLIPS_VERSION)	       \
	 	$(SHARED_LIBRARY_DIR)/libclips.so
	-ldconfig -n -v $(SHARED_LIBRARY_DIR)

clean:
	-rm clips.zip
	-rm -fr clips_source build dist clipspy.egg-info
