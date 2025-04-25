/*
 this package name is inspired by metal I/O

*/

package mio

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

// TODO:
//  logger and error handling

func CopyDirectory(src, dest string) error {

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {

		sourcePath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, er := os.Stat(sourcePath)
		if er != nil {
			return er
		}

		stat, ok := fileInfo.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", sourcePath)
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if e := CreateIfNotExists(destPath, 0755); e != nil {
				return e
			}
			if e := CopyDirectory(sourcePath, destPath); e != nil {
				return e
			}
		case os.ModeSymlink:
			if e := CopySymLink(sourcePath, destPath); e != nil {
				return e
			}
		//	sfi.Mode().IsRegular() with os.Stat() can help determine if the srcDir is a device and should be avoided
		default:
			if e := Copy(sourcePath, destPath); e != nil {
				return e
			}
		}

		if e := os.Lchown(destPath, int(stat.Uid), int(stat.Gid)); e != nil {
			return e
		}

		fInfo, er := entry.Info()
		if er != nil {
			return er
		}

		isSymlink := fInfo.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if e := os.Chmod(destPath, fInfo.Mode()); e != nil {
				return e
			}
		}
	}

	return nil
}

func Copy(srcFile, dstFile string) error {

	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}

	defer func(out *os.File) {
		er := out.Close()
		if er != nil {
			fmt.Printf("error closing destination file: %v\n", er)
		}
	}(out)

	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}

	defer func(in *os.File) {
		er := in.Close()
		if er != nil {
			fmt.Printf("error closing source file: %v\n", er)
		}
	}(in)

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

func removeDirectory(dirPath string) error {

	err := os.RemoveAll(dirPath)
	if err != nil {
		return fmt.Errorf("failed to remove directory recursively: %w", err)
	}

	return nil
}

func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

func CreateIfNotExists(dir string, perm os.FileMode) error {
	if Exists(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}

	return nil
}

func CopySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return err
	}
	return os.Symlink(link, dest)
}

func CalculateFileMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("error closing file during MD5 calculation: %v\n", err)
		}
	}(file)

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func VerifyFileIntegrity(filePath, expectedHash string) (bool, error) {
	calculatedHash, err := CalculateFileMD5(filePath)
	if err != nil {
		return false, err
	}
	return calculatedHash == expectedHash, nil
}
