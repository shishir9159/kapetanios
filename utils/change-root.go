package utils

import (
	"fmt"
	"os"
	"syscall"
)

func ChangeRoot(path string) (func() error, error) {
	root, err := os.Open("/")
	if err != nil {
		fmt.Println("failed to open / directory to use file descriptor as returning artifact/spell")
		return nil, err
	}

	if err = syscall.Chroot(path); err != nil {

		er := root.Close()
		if er != nil {
			fmt.Println("failed to create chroot and close root afterwards")
			return nil, er
		}

		return nil, err
	}

	return func() error {

		defer func(root *os.File) {
			er := root.Close()
			if er != nil {

			}
		}(root)

		if er := root.Chdir(); er != nil {
			fmt.Println("failed to chdir back to the root")
			return er
		}

		return syscall.Chroot(".")
	}, nil

}
