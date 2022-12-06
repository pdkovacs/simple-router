goos_list="linux darvin windows"
goarch_list="amd64 arm64"

mkdir -p release

for GOOS in $goos_list;
do
    for GOARCH in $goarch_list;
    do
        echo $GOOS/$GOARCH
        set -x
        executable_file_name=release/simple-router_${GOOS}_${GOARCH}
        test "$GOOS" == "windows" && executable_file_name="${executable_file_name}.exe"
        go build -o ${executable_file_name} ./cmd/main.go
    done
done