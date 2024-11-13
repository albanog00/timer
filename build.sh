#!/bin/sh

set -eu

root=$(dirname $(realpath $0))
src_path=$root/cmd/main.go
build_path=$root/build/timer
install_path=/usr/bin/timer

print_help () {
    echo "usage: $0 {run|install|installsymlink}"
    echo
    echo "Commands:"
    echo "  run            Builds the binary at 'cmd/main.go' in 'build/timer'."
    echo "  install        Copies the binary at 'build/timer' in '/usr/bin/timer'."
    echo "  installsymlink Create a symlink of 'build/timer' in '/usr/bin/timer'."
}

if [ $# = 0 ]; then
    print_help
    exit
fi

if [ $1 = "run" ]; then
    go build -o $build_path $src_path
    echo "Build done! The binary can be found at '$root/build/timer'."
    exit
fi

if [ $1 = "install" ]; then 
    if [ ! -f $build_path ]; then
        echo "Binary not found. Ensure that build_path is correct and Run '$0 run'."
        exit 1
    fi

    cp $build_path $install_path
    echo "Binary copied successfully at $install_path."
    exit
fi

if [ $1 = "installsymlink" ]; then
    if [ -f $install_path ]; then
        echo "Binary or symlink already exists at $install_path."
        exit 1
    fi

    if [ ! -f $build_path ]; then
        echo "Binary not found. Ensure that build_path is correct and Run '$0 run'."
        exit 1
    fi

    ln -s $build_path $install_path
    echo "Created a symlink to $build_path at $install_path"
    exit
fi

if [ $1 = "uninstall" ]; then
    if [ -f $install_path ]; then
        rm $install_path
        echo "$install_path deleted."
    else
        echo "$install_path not found."
    fi
    exit
fi

echo "invalid command '$1'."
print_help
exit 1
