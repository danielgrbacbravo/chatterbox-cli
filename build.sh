#!/bin/bash
# Name of your project
PROJECT_NAME="chatterbox-cli"

# Output name
OUTPUT_NAME="chatterbox"

# Go path
GO_PATH=$(which go)

# Platforms to build for
PLATFORMS=("darwin/amd64" "darwin/arm64" "windows/amd64" "linux/amd64")

# Build directory
BUILD_DIR="./bin"

# Create the build directory if it doesn't exist
mkdir -p $BUILD_DIR

# Build for all platforms
for platform in "${PLATFORMS[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}

    output_name=$BUILD_DIR/$OUTPUT_NAME'-'$GOOS'-'$GOARCH

    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi  

    env GOOS=$GOOS GOARCH=$GOARCH $GO_PATH build -o $output_name $PROJECT_NAME

    echo 'Built for '$platform
done
