
# aliases
all: tinygo
tinygo: build/tinygo.exe

# Default build and source directories, as created by `make llvm-build`.
LLVM_BUILDDIR ?= llvm-build
CLANG_SRC ?= llvm-project/clang
LLD_SRC ?= llvm-project/lld

.PHONY: all tinygo build/tinygo.exe test $(LLVM_BUILDDIR) llvm-source clean fmt gen-device gen-device-nrf gen-device-avr

LLVM_COMPONENTS = all-targets analysis asmparser asmprinter bitreader bitwriter codegen core coroutines debuginfodwarf executionengine instrumentation interpreter ipo irreader linker lto mc mcjit objcarcopts option profiledata scalaropts support target

LLVM_CMAKE_FLAGS += -DLLVM_TARGETS_TO_BUILD="X86;ARM;AArch64;WebAssembly" -DLLVM_EXPERIMENTAL_TARGETS_TO_BUILD="AVR;RISCV" -DCMAKE_BUILD_TYPE=Release -DLLVM_ENABLE_ASSERTIONS=OFF -DLIBCLANG_BUILD_STATIC=ON -DLLVM_ENABLE_TERMINFO=OFF -DLLVM_ENABLE_ZLIB=OFF -DLLVM_ENABLE_PROJECTS="clang;lld" -DLLVM_TOOL_CLANG_TOOLS_EXTRA_BUILD=OFF

ifeq ($(OS),Windows_NT)
    # LLVM compiled using MinGW on Windows appears to have problems with threads.
    # Without this flag, linking results in errors like these:
    #     libLLVMSupport.a(Threading.cpp.obj):Threading.cpp:(.text+0x55): undefined reference to `std::thread::hardware_concurrency()'
    LLVM_CMAKE_FLAGS += -DLLVM_ENABLE_THREADS=OFF
    # note: the following might work?
    LIBCLANG_PATH = $(abspath $(LLVM_BUILDDIR))/lib/liblibclang.dll.a
    #LIBCLANG_PATH = \
    #    $(abspath $(LLVM_BUILDDIR))/tools/clang/tools/libclang/CMakeFiles/libclang.dir/ARCMigrate.cpp.obj \
    #    $(abspath $(LLVM_BUILDDIR))/tools/clang/tools/libclang/CMakeFiles/libclang.dir/BuildSystem.cpp.obj \
    #    $(abspath $(LLVM_BUILDDIR))/tools/clang/tools/libclang/CMakeFiles/libclang.dir/CIndex.cpp.obj \
    #    $(abspath $(LLVM_BUILDDIR))/tools/clang/tools/libclang/CMakeFiles/libclang.dir/CIndexCodeCompletion.cpp.obj \
    #    $(abspath $(LLVM_BUILDDIR))/tools/clang/tools/libclang/CMakeFiles/libclang.dir/CIndexCXX.cpp.obj \
    #    $(abspath $(LLVM_BUILDDIR))/tools/clang/tools/libclang/CMakeFiles/libclang.dir/CIndexDiagnostic.cpp.obj \
    #    $(abspath $(LLVM_BUILDDIR))/tools/clang/tools/libclang/CMakeFiles/libclang.dir/CIndexer.cpp.obj \
    #    $(abspath $(LLVM_BUILDDIR))/tools/clang/tools/libclang/CMakeFiles/libclang.dir/CIndexHigh.cpp.obj \
    #    $(abspath $(LLVM_BUILDDIR))/tools/clang/tools/libclang/CMakeFiles/libclang.dir/CIndexInclusionStack.cpp.obj \
    #    $(abspath $(LLVM_BUILDDIR))/tools/clang/tools/libclang/CMakeFiles/libclang.dir/CIndexUSRs.cpp.obj \
    #    $(abspath $(LLVM_BUILDDIR))/tools/clang/tools/libclang/CMakeFiles/libclang.dir/CXCompilationDatabase.cpp.obj \
    #    $(abspath $(LLVM_BUILDDIR))/tools/clang/tools/libclang/CMakeFiles/libclang.dir/CXCursor.cpp.obj \
    #    $(abspath $(LLVM_BUILDDIR))/tools/clang/tools/libclang/CMakeFiles/libclang.dir/CXIndexDataConsumer.cpp.obj \
    #    $(abspath $(LLVM_BUILDDIR))/tools/clang/tools/libclang/CMakeFiles/libclang.dir/CXLoadedDiagnostic.cpp.obj \
    #    $(abspath $(LLVM_BUILDDIR))/tools/clang/tools/libclang/CMakeFiles/libclang.dir/CXSourceLocation.cpp.obj \
    #    $(abspath $(LLVM_BUILDDIR))/tools/clang/tools/libclang/CMakeFiles/libclang.dir/CXStoredDiagnostic.cpp.obj \
    #    $(abspath $(LLVM_BUILDDIR))/tools/clang/tools/libclang/CMakeFiles/libclang.dir/CXString.cpp.obj \
    #    $(abspath $(LLVM_BUILDDIR))/tools/clang/tools/libclang/CMakeFiles/libclang.dir/CXType.cpp.obj \
    #    $(abspath $(LLVM_BUILDDIR))/tools/clang/tools/libclang/CMakeFiles/libclang.dir/Indexing.cpp.obj
    START_GROUP = -Wl,--start-group
    END_GROUP = -Wl,--end-group
else
    LIBCLANG_PATH = $(abspath $(LLVM_BUILDDIR))/lib/libclang.a
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Linux)
        START_GROUP = -Wl,--start-group
        END_GROUP = -Wl,--end-group
    endif
endif

CLANG_LIBS = $(LIBCLANG_PATH) $(START_GROUP) -lclangAnalysis -lclangARCMigrate -lclangAST -lclangASTMatchers -lclangBasic -lclangCodeGen -lclangCrossTU -lclangDriver -lclangDynamicASTMatchers -lclangEdit -lclangFormat -lclangFrontend -lclangFrontendTool -lclangHandleCXX -lclangHandleLLVM -lclangIndex -lclangLex -lclangParse -lclangRewrite -lclangRewriteFrontend -lclangSema -lclangSerialization -lclangStaticAnalyzerCheckers -lclangStaticAnalyzerCore -lclangStaticAnalyzerFrontend -lclangTooling -lclangToolingASTDiff -lclangToolingCore -lclangToolingInclusions $(END_GROUP) -lstdc++

LLD_LIBS = $(START_GROUP) -llldCOFF -llldCommon -llldCore -llldDriver -llldELF -llldMachO -llldMinGW -llldReaderWriter -llldWasm -llldYAML $(END_GROUP)


# For static linking.
ifneq ("$(wildcard $(LLVM_BUILDDIR)/bin/llvm-config*)","")
    CGO_CPPFLAGS=$(shell $(LLVM_BUILDDIR)/bin/llvm-config --cppflags) -I$(abspath $(CLANG_SRC))/include -I$(abspath $(LLD_SRC))/include
    CGO_CXXFLAGS=-std=c++11
    CGO_LDFLAGS=-L$(abspath $(LLVM_BUILDDIR)/lib) $(CLANG_LIBS) $(LLD_LIBS) $(shell $(LLVM_BUILDDIR)/bin/llvm-config --ldflags --libs --system-libs $(LLVM_COMPONENTS)) -static-libgcc -static-libstdc++ -static -lversion #-Wl,--whole-archive -lpthread -Wl,--no-whole-archive -std=c++11
endif


clean:
	@rm -rf build

FMT_PATHS = ./*.go cgo compiler interp ir loader src/device/arm src/examples src/machine src/os src/reflect src/runtime src/sync src/syscall
fmt:
	@gofmt -l -w $(FMT_PATHS)
fmt-check:
	@unformatted=$$(gofmt -l $(FMT_PATHS)); [ -z "$$unformatted" ] && exit 0; echo "Unformatted:"; for fn in $$unformatted; do echo "  $$fn"; done; exit 1


gen-device: gen-device-avr gen-device-nrf gen-device-sam gen-device-sifive gen-device-stm32

gen-device-avr:
	./tools/gen-device-avr.py lib/avr/packs/atmega src/device/avr/
	./tools/gen-device-avr.py lib/avr/packs/tiny src/device/avr/
	go fmt ./src/device/avr

gen-device-nrf:
	./tools/gen-device-svd.py lib/nrfx/mdk/ src/device/nrf/ --source=https://github.com/NordicSemiconductor/nrfx/tree/master/mdk
	go fmt ./src/device/nrf

gen-device-sam:
	./tools/gen-device-svd.py lib/cmsis-svd/data/Atmel/ src/device/sam/ --source=https://github.com/posborne/cmsis-svd/tree/master/data/Atmel
	go fmt ./src/device/sam

gen-device-sifive:
	./tools/gen-device-svd.py lib/cmsis-svd/data/SiFive-Community/ src/device/sifive/ --source=https://github.com/AdaCore/svd2ada/tree/master/CMSIS-SVD/SiFive-Community
	go fmt ./src/device/sifive

gen-device-stm32:
	./tools/gen-device-svd.py lib/cmsis-svd/data/STMicro/ src/device/stm32/ --source=https://github.com/posborne/cmsis-svd/tree/master/data/STMicro
	go fmt ./src/device/stm32


# Get LLVM sources.
llvm-project/README.md:
	git clone -b release/8.x https://github.com/llvm/llvm-project
llvm-source: llvm-project/README.md

# Configure LLVM.
TINYGO_SOURCE_DIR=$(shell pwd)
$(LLVM_BUILDDIR)/build.ninja: llvm-source
	mkdir -p $(LLVM_BUILDDIR); cd $(LLVM_BUILDDIR); cmake -G Ninja $(TINYGO_SOURCE_DIR)/llvm-project/llvm $(LLVM_CMAKE_FLAGS)

# Build LLVM.
$(LLVM_BUILDDIR): $(LLVM_BUILDDIR)/build.ninja
	cd $(LLVM_BUILDDIR); ninja


# Build the Go compiler.
build/tinygo.exe:
	@if [ ! -f "$(LLVM_BUILDDIR)/bin/llvm-config" ]; then echo "Fetch and build LLVM first by running:"; echo "  make llvm-source"; echo "  make $(LLVM_BUILDDIR)"; exit 1; fi
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" go build -o build/tinygo.exe -tags byollvm .

test:
	CGO_CPPFLAGS="$(CGO_CPPFLAGS)" CGO_CXXFLAGS="$(CGO_CXXFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" go test -v -tags byollvm .

tinygo-test:
	cd tests/tinygotest && tinygo test

.PHONY: smoketest
smoketest:
	# test all examples
	tinygo build -size short -o test.elf -target=pca10040            examples/blinky1
	tinygo build -size short -o test.elf -target=pca10040            examples/adc
	tinygo build -size short -o test.elf -target=pca10040            examples/blinkm
	tinygo build -size short -o test.elf -target=pca10040            examples/blinky2
	tinygo build -size short -o test.elf -target=pca10040            examples/button
	tinygo build -size short -o test.elf -target=pca10040            examples/button2
	tinygo build -size short -o test.elf -target=pca10040            examples/echo
	tinygo build -size short -o test.elf -target=circuitplay-express examples/i2s
	tinygo build -size short -o test.elf -target=pca10040            examples/mcp3008
	tinygo build -size short -o test.elf -target=microbit            examples/microbit-blink
	tinygo build -size short -o test.elf -target=pca10040            examples/pwm
	tinygo build -size short -o test.elf -target=pca10040            examples/serial
	tinygo build -size short -o test.elf -target=pca10040            examples/test
	# test all targets/boards
	tinygo build             -o test.wasm -tags=pca10040             examples/blinky2
	tinygo build -size short -o test.elf -target=microbit            examples/echo
	tinygo build -size short -o test.elf -target=nrf52840-mdk        examples/blinky1
	tinygo build -size short -o test.elf -target=pca10031            examples/blinky1
	tinygo build -size short -o test.elf -target=bluepill            examples/blinky1
	tinygo build -size short -o test.elf -target=reelboard           examples/blinky1
	tinygo build -size short -o test.elf -target=reelboard           examples/blinky2
	tinygo build -size short -o test.elf -target=pca10056            examples/blinky1
	tinygo build -size short -o test.elf -target=pca10056            examples/blinky2
	tinygo build -size short -o test.elf -target=itsybitsy-m0        examples/blinky1
	tinygo build -size short -o test.elf -target=feather-m0          examples/blinky1
	tinygo build -size short -o test.elf -target=trinket-m0          examples/blinky1
	tinygo build -size short -o test.elf -target=circuitplay-express examples/blinky1
	tinygo build -size short -o test.elf -target=stm32f4disco        examples/blinky1
	tinygo build -size short -o test.elf -target=stm32f4disco        examples/blinky2
	tinygo build -size short -o test.elf -target=circuitplay-express examples/i2s
	tinygo build -size short -o test.elf -target=gameboy-advance     examples/gba-display
ifneq ($(AVR), 0)
	tinygo build -size short -o test.elf -target=arduino             examples/blinky1
	tinygo build -size short -o test.elf -target=digispark           examples/blinky1
endif
ifneq ($(RISCV), 0)
	tinygo build -size short -o test.elf -target=hifive1b            examples/blinky1
endif
	tinygo build             -o wasm.wasm -target=wasm               examples/wasm/export
	tinygo build             -o wasm.wasm -target=wasm               examples/wasm/main

release: build/tinygo.exe gen-device
	@mkdir -p build/release/tinygo/bin
	@mkdir -p build/release/tinygo/lib/clang/include
	@mkdir -p build/release/tinygo/lib/CMSIS/CMSIS
	@mkdir -p build/release/tinygo/lib/compiler-rt/lib
	@mkdir -p build/release/tinygo/lib/nrfx
	@mkdir -p build/release/tinygo/pkg/armv6m-none-eabi
	@mkdir -p build/release/tinygo/pkg/armv7m-none-eabi
	@mkdir -p build/release/tinygo/pkg/armv7em-none-eabi
	@echo copying source files
	@cp -p  build/tinygo.exe             build/release/tinygo/bin
	@cp -p $(abspath $(CLANG_SRC))/lib/Headers/*.h build/release/tinygo/lib/clang/include
	@cp -rp lib/CMSIS/CMSIS/Include      build/release/tinygo/lib/CMSIS/CMSIS
	@cp -rp lib/CMSIS/README.md          build/release/tinygo/lib/CMSIS
	@cp -rp lib/compiler-rt/lib/builtins build/release/tinygo/lib/compiler-rt/lib
	@cp -rp lib/compiler-rt/LICENSE.TXT  build/release/tinygo/lib/compiler-rt
	@cp -rp lib/compiler-rt/README.txt   build/release/tinygo/lib/compiler-rt
	@cp -rp lib/nrfx/*                   build/release/tinygo/lib/nrfx
	@cp -rp src                          build/release/tinygo/src
	@cp -rp targets                      build/release/tinygo/targets
	ls -l ./build
	file ./build/tinygo.exe
	./build/tinygo.exe version
	./build/tinygo.exe build-builtins -target=armv6m-none-eabi  -o build/release/tinygo/pkg/armv6m-none-eabi/compiler-rt.a
	./build/tinygo.exe build-builtins -target=armv7m-none-eabi  -o build/release/tinygo/pkg/armv7m-none-eabi/compiler-rt.a
	./build/tinygo.exe build-builtins -target=armv7em-none-eabi -o build/release/tinygo/pkg/armv7em-none-eabi/compiler-rt.a
	tar -czf build/release.tar.gz -C build/release tinygo
