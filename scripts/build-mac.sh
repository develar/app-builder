GOOS=darwin GOARCH=amd64 go build -ldflags='-s -w' -o dist/darwin_amd64/app-builder

# ln -sf ~/go/src/github.com/develar/app-builder/dist/darwin_amd64/app-builder ~/Documents/electron-builder/node_modules/app-builder-bin/mac/app-builder