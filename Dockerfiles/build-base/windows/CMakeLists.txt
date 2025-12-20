cmake_minimum_required(VERSION 3.16)
project(libmodbus_win C)

option(MODBUS_ENABLE_RTU "Build RTU backend (modbus-rtu.c)" ON)
option(MODBUS_ENABLE_TCP "Build TCP backend (modbus-tcp.c)" ON)
option(MODBUS_FIX_MSVC_SOCKOPT_WARNINGS "Fix MSVC winsock C4133 warnings" ON)

set(SRC_DIR "${CMAKE_CURRENT_SOURCE_DIR}/src")

# --- 1) Prepare shim include dir (build-only includes live here) ---
set(SHIM_INC_DIR "${CMAKE_CURRENT_BINARY_DIR}/include")
file(MAKE_DIRECTORY "${SHIM_INC_DIR}")

# We expect config.h to exist in ${SRC_DIR}/config.h (you generate it via src/win32/configure.js)
if(NOT EXISTS "${SRC_DIR}/win32/config.h")
  message(FATAL_ERROR
    "config.h not found at ${SRC_DIR}/win32/config.h. "
    "Run src/win32/configure.js and copy config.h into src/ before configuring with CMake.")
endif()

# --- 2) Generate shim modbus.h (static-friendly MODBUS_API) ---
file(READ "${SRC_DIR}/modbus.h" MODBUS_H_TXT)

set(MODBUS_API_BLOCK
"#if defined(_MSC_VER)
# if defined(DLLBUILD)
/* define DLLBUILD when building the DLL */
#  define MODBUS_API __declspec(dllexport)
# else
#  define MODBUS_API __declspec(dllimport)
# endif
#else
# define MODBUS_API
#endif")

set(MODBUS_API_BLOCK_REPL
"#if defined(_MSC_VER)
  /* Static library: do NOT use dllimport/dllexport in public headers */
# if defined(DLLBUILD)
#  define MODBUS_API __declspec(dllexport)
# else
#  define MODBUS_API
# endif
#else
# define MODBUS_API
#endif")

string(REPLACE "${MODBUS_API_BLOCK}" "${MODBUS_API_BLOCK_REPL}" MODBUS_H_TXT "${MODBUS_H_TXT}")
file(WRITE "${SHIM_INC_DIR}/modbus.h" "${MODBUS_H_TXT}")

# Copy remaining headers needed to compile (build-only set includes private/config)
file(COPY "${SRC_DIR}/modbus-tcp.h"         DESTINATION "${SHIM_INC_DIR}")
file(COPY "${SRC_DIR}/modbus-rtu.h"         DESTINATION "${SHIM_INC_DIR}")
file(COPY "${SRC_DIR}/modbus-version.h"     DESTINATION "${SHIM_INC_DIR}")

# Build-only headers required by .c files
file(COPY "${SRC_DIR}/win32/config.h"             DESTINATION "${SHIM_INC_DIR}")
file(COPY "${SRC_DIR}/modbus-private.h"     DESTINATION "${SHIM_INC_DIR}")
file(COPY "${SRC_DIR}/modbus-tcp-private.h" DESTINATION "${SHIM_INC_DIR}")
file(COPY "${SRC_DIR}/modbus-rtu-private.h" DESTINATION "${SHIM_INC_DIR}")

# --- 3) Generate shim sources (do not modify upstream sources on disk) ---
set(SHIM_SRC_DIR "${CMAKE_CURRENT_BINARY_DIR}/shim_src")
file(MAKE_DIRECTORY "${SHIM_SRC_DIR}")

function(make_shim_source in_file out_file)
  file(READ "${in_file}" TXT)

  # Force includes via include paths (so our shim headers win)
  string(REPLACE "#include \"modbus.h\""          "#include <modbus.h>"          TXT "${TXT}")
  string(REPLACE "#include \"modbus-tcp.h\""      "#include <modbus-tcp.h>"      TXT "${TXT}")
  string(REPLACE "#include \"modbus-rtu.h\""      "#include <modbus-rtu.h>"      TXT "${TXT}")
  string(REPLACE "#include \"modbus-version.h\""  "#include <modbus-version.h>"  TXT "${TXT}")

  # Some files include private/config with quotes; route those too
  string(REPLACE "#include \"config.h\""              "#include <config.h>"              TXT "${TXT}")
  string(REPLACE "#include \"modbus-private.h\""      "#include <modbus-private.h>"      TXT "${TXT}")
  string(REPLACE "#include \"modbus-tcp-private.h\""  "#include <modbus-tcp-private.h>"  TXT "${TXT}")
  string(REPLACE "#include \"modbus-rtu-private.h\""  "#include <modbus-rtu-private.h>"  TXT "${TXT}")

  # Optional: silence MSVC winsock setsockopt warnings (C4133)
  # On Windows setsockopt expects const char* optval; libmodbus passes &int.
  if (MSVC AND MODBUS_FIX_MSVC_SOCKOPT_WARNINGS)
    string(REPLACE ", &option, sizeof(int))" ", (const char*)&option, sizeof(int))" TXT "${TXT}")
  endif()

  file(WRITE "${out_file}" "${TXT}")
endfunction()

# Base sources
make_shim_source("${SRC_DIR}/modbus.c"       "${SHIM_SRC_DIR}/modbus.c")
make_shim_source("${SRC_DIR}/modbus-data.c"  "${SHIM_SRC_DIR}/modbus-data.c")

set(MODBUS_SOURCES
  "${SHIM_SRC_DIR}/modbus.c"
  "${SHIM_SRC_DIR}/modbus-data.c"
)

if(MODBUS_ENABLE_TCP)
  make_shim_source("${SRC_DIR}/modbus-tcp.c" "${SHIM_SRC_DIR}/modbus-tcp.c")
  list(APPEND MODBUS_SOURCES "${SHIM_SRC_DIR}/modbus-tcp.c")
endif()

if(MODBUS_ENABLE_RTU)
  make_shim_source("${SRC_DIR}/modbus-rtu.c" "${SHIM_SRC_DIR}/modbus-rtu.c")
  list(APPEND MODBUS_SOURCES "${SHIM_SRC_DIR}/modbus-rtu.c")
endif()

add_library(modbus STATIC ${MODBUS_SOURCES})

# --- MSVC: silence WinSock signature warning (setsockopt wants const char*) ---
if (MSVC)
  target_compile_options(modbus PRIVATE /wd4133)
endif()

# Use shim headers for build + consumers (PUBLIC), but we will INSTALL only public headers explicitly
target_include_directories(modbus
  PUBLIC
    "${SHIM_INC_DIR}"
)

if (WIN32)
  target_link_libraries(modbus PUBLIC ws2_32)
endif()

# --- Install: only public headers (no config.h, no *-private.h) ---
include(GNUInstallDirs)

install(TARGETS modbus
  ARCHIVE DESTINATION "${CMAKE_INSTALL_LIBDIR}"
)

set(PUBLIC_HEADERS
  "${SHIM_INC_DIR}/modbus.h"
  "${SHIM_INC_DIR}/modbus-version.h"
)

if(MODBUS_ENABLE_TCP)
  list(APPEND PUBLIC_HEADERS "${SHIM_INC_DIR}/modbus-tcp.h")
endif()

if(MODBUS_ENABLE_RTU)
  list(APPEND PUBLIC_HEADERS "${SHIM_INC_DIR}/modbus-rtu.h")
endif()

install(FILES ${PUBLIC_HEADERS}
  DESTINATION "${CMAKE_INSTALL_INCLUDEDIR}"
)