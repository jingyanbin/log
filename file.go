package log

import (
	"os"
	"path/filepath"
)


func openFile(folderPath string, fileName string, flag int, perm os.FileMode) (file *os.File, err error) {
	path := filepath.Join(folderPath, fileName)
	file, err = os.OpenFile(path, flag, perm)
	if err != nil{
		var has bool
		has, err = existFolder(folderPath)
		if err != nil{
			return
		}
		if has == false{
			err = os.MkdirAll(folderPath, os.ModeDir)
			if err != nil{
				return
			}
			file, err = os.OpenFile(path, flag, perm)
		}
	}
	return
}

//文件或文件夹是否存在
func exist(path string) (bool, error){
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// 文件夹是否存在
func existFolder(path string) (bool, error){
	handle, err := os.Stat(path)
	if err == nil {
		if handle.IsDir(){
			return true, nil
		} else{
			return false, newError("not is folder: %s", path)
		}
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}


type handleFile struct {
	folderPath string
	fileName string
	flag int
	perm os.FileMode
	handle *os.File
}

func (my *handleFile)PathName() string {
	return filepath.Join(my.folderPath, my.fileName)
}

func (my *handleFile)SetPathName(folderPath, fileName string) bool {
	if my.folderPath == folderPath && my.fileName == fileName{
		return false
	}
	my.Close()
	my.folderPath = folderPath
	my.fileName = fileName
	return true
}

func (my *handleFile)WriteString(s string) (n int, err error){
	n, err = my.Write([]byte(s))
	return
}

func (my *handleFile)Write(b []byte) (n int, err error){
	if my.handle == nil{
		my.handle, err = openFile(my.folderPath,my.fileName,my.flag,my.perm)
		if err != nil{
			return
		}
	} else {
		var hasLog bool
		hasLog, err = exist(filepath.Join(my.folderPath, my.fileName))
		if !hasLog{
			my.handle.Close()
			my.handle, err = openFile(my.folderPath,my.fileName,my.flag,my.perm)
			if err != nil{
				return
			}
		}
	}
	n, err = my.handle.Write(b)
	return
}

func (my *handleFile)Close() {
	if my.handle == nil{
		return
	}
	my.handle.Close()
	my.handle = nil
}

var _execDir string

func execDir() string{
	if _execDir == ""{
		dir, err := os.Getwd()
		if err != nil{
			panic(err)
		}
		_execDir = dir
	}
	return _execDir
}
