package main

import (
	"fmt"
	"github.com/levigross/grequests"
	"os"
	"path"
	"runtime"
	"strconv"
)

var downloadPath string
var downloadFile *os.File
var url string

func createFile(fileSize int64) (*os.File, error) {
	//TODO 剩余空间检测
	file, err := os.Create(downloadPath)
	if err != nil {
		fmt.Println("下载失败-创建文件失败")
		return nil, err
	}
	err = file.Truncate(fileSize)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return file, nil
}

func download(startIndex int64, endIndex int64) error {
	ua := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.108 Safari/537.36"
	blockRange := "bytes=" + strconv.FormatInt(startIndex, 10) + "-" + strconv.FormatInt(endIndex, 10)
	ro := &grequests.RequestOptions{Headers: map[string]string{"User-Agent": ua, "Range" : blockRange}}
	resp, err := grequests.Get(url, ro)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	if resp.StatusCode == 206 {
		err = writeFile(resp.Bytes(), startIndex)
		if err != nil {
			return err
		}
	}
	return nil
}


func writeFile(content []byte, startIndex int64) error {
	_, err := downloadFile.WriteAt(content, startIndex)
	if err != nil {
		return err
	}
	return nil
}

func run(blockSize int, threadsNum int, totalLength int)  {
	finishChan := make(chan int)
	for i := 0; i < threadsNum; i++ {
		startIndex := i * blockSize
		endIndex := (i + 1) * blockSize - 1
		if i == threadsNum- 1 {
			endIndex = totalLength
		}
		go func(proc int) {
			err := download(int64(startIndex), int64(endIndex))
			if err != nil {
				fmt.Println(err.Error())
				finishChan <- 0
				return
			}
			finishChan <- proc + 1
		}(i)
	}

	for i := 0; i < threadsNum; i++ {
		x := <-finishChan
			if x == 0 {
				fmt.Println("有错误")
			}else {
				fmt.Println(x, "段下载完成")
			}
	}
	_ = downloadFile.Close()

}

func main()  {
	threadsNum := runtime.NumCPU() + 1
	fmt.Println("线程数" , threadsNum)
	url = "http://mirrors.163.com/archlinux/iso/latest/archlinux-2019.05.02-x86_64.iso.torrent"
	downloadPath = path.Base(url) //todo 文件名检测
	ua := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.108 Safari/537.36"
	ro := &grequests.RequestOptions{Headers: map[string]string{"User-Agent": ua}}
	resp, err := grequests.Head(url,ro)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	//todo 多线程下载支持检测
	str, ok := resp.Header["Content-Length"]
	if resp.StatusCode == 200 && ok && len(str) == 1 {
		contentLength, _ := strconv.Atoi(str[0])
		if contentLength <= threadsNum* 1000 {
			threadsNum = 1
		}
		blockSize := contentLength / threadsNum

		f, err := createFile(int64(contentLength))
		if err != nil {
			fmt.Println("下载失败")
			fmt.Println(err.Error())
			f.Close()
			return
		}
		downloadFile = f
		run(blockSize, threadsNum, contentLength)
	}else {
		fmt.Println("下载失败")
	}

}