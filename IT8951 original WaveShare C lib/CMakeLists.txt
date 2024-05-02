cmake_minimum_required(VERSION 3.25)

project(epd VERSION 1.0)

add_library(epd STATIC Config/DEV_Config.c e-Paper/EPD_IT8951.c)
target_link_libraries(epd LINK_PUBLIC -lbcm2835)
target_include_directories(epd PUBLIC
                          "${PROJECT_BINARY_DIR}"
                          "${PROJECT_SOURCE_DIR}/Config" "${PROJECT_SOURCE_DIR}/e-Paper"
    )
