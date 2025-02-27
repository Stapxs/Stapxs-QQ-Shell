#!/bin/bash

# 设置目标平台和架构
PLATFORMS=("linux/amd64" "linux/arm64" "windows/amd64" "darwin/amd64" "darwin/arm64")
OUTPUT_DIR="./build"            # 定义输出目录
APP_NAME="ssqq-shell"           # 定义程序名称

# 通过 git 获取短 hash 作为版本号
VERSION=$(git rev-parse --short HEAD)

mkdir -p "$OUTPUT_DIR"

# 遍历平台进行编译
for PLATFORM in "${PLATFORMS[@]}"
do
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}
    OUTPUT_NAME="${APP_NAME}-${VERSION}-${GOOS}-${GOARCH}"

    # 为 Windows 添加 .exe 后缀
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME+=".exe"
    fi

    # 最终输出路径
    OUTPUT_PATH="${OUTPUT_DIR}/${OUTPUT_NAME}"

    echo "正在构建：$GOOS/$GOARCH"
    GOOS=$GOOS GOARCH=$GOARCH go build -o "$OUTPUT_PATH"

    # 检查编译结果
    if [ $? -ne 0 ]; then
        echo "构建失败：$GOOS/$GOARCH"
    else
        echo "构建成功：$OUTPUT_PATH"
    fi
done
