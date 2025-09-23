#!/bin/zsh

# ----------
# Check OrbStack (Docker Process)
# ----------
echo "Start Compiling task ..."

# 检查 OrbStack 进程是否运行
if pgrep -x "OrbStack" > /dev/null; then
    echo "OrbStack is running"
else
    echo "Error: OrbStack is not running" >&2
    exit 1
fi

# ----------
# Build on MacOS
# ----------
# MacOS (Apple Sillicon)
fyne p --release --name GenDict --app-id cn.iamcc.gendict --app-version 1.0.0
# Move
mkdir -p ./fyne-cross/dist/darwin-arm64
mv ./GenDict.app ./fyne-cross/dist/darwin-arm64
cd ./fyne-cross/dist/darwin-arm64/
zip -r GenDict-darwin-arm64.zip GenDict.app
rm -rf ./GenDict.app
cd ../../../

# ----------
# Build Cross
# ----------
# Linux (x86-64)
fyne-cross linux --name GenDict -arch amd64 --app-version 1.0.0 --app-id cn.iamcc.gendict
# Linux (ARM64)
fyne-cross linux --name GenDict -arch arm64 --app-version 1.0.0 --app-id cn.iamcc.gendict
# Windows (x86-64)
fyne-cross windows --name GenDict -arch amd64 --app-version 1.0.0 --app-id cn.iamcc.gendict
# Windows (ARM64)
fyne-cross windows --name GenDict -arch arm64 --app-version 1.0.0 --app-id cn.iamcc.gendict

# Move package and clean folder
cd ./fyne-cross/dist
# Move
mv darwin-arm64/GenDict-darwin-arm64.zip ./GenDict-darwin-arm64.zip
mv linux-amd64/GenDict.tar.xz ./GenDict-linux-amd64.tar.xz
mv linux-arm64/GenDict.tar.xz ./GenDict-linux-arm64.tar.xz
mv windows-amd64/GenDict.zip ./GenDict-windows-amd64.zip
mv windows-arm64/GenDict.zip ./GenDict-windows-arm64.zip
# Clean
rmdir darwin-arm64/
rmdir linux-amd64/
rmdir linux-arm64/
rmdir windows-amd64/
rmdir windows-arm64/

echo "Compiling task finished ..."