CMAKE_MINIMUM_REQUIRED(VERSION 2.8)

IF(COMMAND CMAKE_POLICY)
  CMAKE_POLICY(SET CMP0037 OLD)
ENDIF(COMMAND CMAKE_POLICY)
PROJECT(cppcmsskel)

set(CMAKE_CXX_COMPILER "/usr/bin/clang++")
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -Wall -std=c++0x")

set(PROJECT_VERSION "0.25.3")
set(PROJECT_SOVERSION "0")
set(ARCHIVE_NAME ${CMAKE_PROJECT_NAME}-${PROJECT_VERSION})
add_custom_target(
    dist
    COMMAND git archive --prefix=${ARCHIVE_NAME}/ HEAD
        | bzip2 > ${CMAKE_BINARY_DIR}/${ARCHIVE_NAME}.tar.bz2
	WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
)
###################
# Dependency check
##################
find_library(CPPCMS cppcms)
find_library(BOOSTER booster)
find_library(CPPDB cppdb)

find_library(MYSQL mysqlclient)
find_path(MYSQL_PATH mysql.h PATH_SUFFIXES mysql)
include_directories(${MYSQL_PATH})





###################
# source files
###################
file(
    GLOB
    framework_generics_files
    cppcms_skel/generics/*
)

file(
    GLOB
    framework_models_files
    cppcms_skel/models/*
)


file(
    GLOB
    framework_results_files
    cppcms_skel/results/*
)

file(
    GLOB
    framework_controllers_files
    cppcms_skel/controllers/webs/*
    cppcms_skel/controllers/generics/*
)


# used to create libccpcmsskel.so

include_directories(
    ${CMAKE_CURRENT_SOURCE_DIR}
    cppcms_skel
)


add_library(
    cppcmsskel
    SHARED
    ${framework_generics_files}
    ${framework_models_files}
    ${framework_controllers_files}
    ${framework_results_files}
)


# TODO cppdb should not be mandatory
set(
    CPPCMSSKEL_LIBS
    ${CPPCMS}
    ${CPPDB}
    ${BOOSTER}
    ${MYSQL}
)

target_link_libraries(
    cppcmsskel
    ${CPPCMSSKEL_LIBS}
)



# we install the "template" files (i.e c++ code with placeholders)
# a local copy of them will be made each time we create a new app
install(
    DIRECTORY template
    DESTINATION share/cppcmsskel
    PATTERN "README" EXCLUDE
)


# we install the "tools" (set of script to do basic tasks)
# in (/usr/local)/share/cppcmsskel
# we do not install them in bin to not polluate it
# instead the framework will create a link to them in the tools directory
# of a given application
install(
    DIRECTORY tools
    DESTINATION share/cppcmsskel
    USE_SOURCE_PERMISSIONS
    PATTERN "*.pyc" EXCLUDE
    PATTERN "__pycache__" EXCLUDE
    PATTERN "*.swp" EXCLUDE
)

# to install the libcppcmsskel
install(
    TARGETS cppcmsskel
    LIBRARY DESTINATION lib
)

set_target_properties(cppcmsskel PROPERTIES VERSION ${PROJECT_VERSION} SOVERSION ${PROJECT_SOVERSION})

# install the header files
install(
    DIRECTORY cppcms_skel
    DESTINATION include
    USE_SOURCE_PERMISSIONS
    PATTERN "README" EXCLUDE
    PATTERN "*.cpp" EXCLUDE
)

# to install create_new_cppcmsskel (command to initiate a new
# cppcmsskel directory)
install(
    FILES tools/create_new_cppcmsskel
    DESTINATION bin
    PERMISSIONS
        WORLD_EXECUTE WORLD_READ
        GROUP_EXECUTE GROUP_READ
        OWNER_EXECUTE OWNER_READ
)

