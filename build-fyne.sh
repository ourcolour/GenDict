#!/bin/zsh

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
cd ../../../

# ----------
# Build Cross
# ----------

# Linux (x86-64)
fyne-cross linux --name GenDict -arch amd64 --app-version 1.0.0 --app-id cn.iamcc.gendict
mv linux-amd64/GenDict.tar.xz linux-amd64/GenDict-linux-amd64.tar.xz

# Linux (ARM64)
fyne-cross linux --name GenDict -arch arm64 --app-version 1.0.0 --app-id cn.iamcc.gendict
mv linux-arm64/GenDict.tar.xz linux-arm64/GenDict-linux-arm64.tar.xz

# Windows (x86-64)
fyne-cross windows --name GenDict -arch amd64 --app-version 1.0.0 --app-id cn.iamcc.gendict
mv windows-amd64/GenDict.zip linux-amd64/GenDict-windows-amd64.zip

# Windows (ARM64)
fyne-cross windows --name GenDict -arch arm64 --app-version 1.0.0 --app-id cn.iamcc.gendict
mv windows-arm64/GenDict.zip windows-arm64/GenDict-windows-arm64.zip
