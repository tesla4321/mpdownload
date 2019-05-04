package main

import (
	"fmt"
	"github.com/levigross/grequests"
	"os"
	"path"
	"runtime"
	"strconv"
)

var download_path string
var download_file *os.File
var url string

func createFile(file_size int64) (*os.File, error) {
	//TODO 剩余空间检测
	file, err := os.Create(download_path)
	if err != nil {
		fmt.Println("下载失败-创建文件失败")
		return nil, err
	}
	err = file.Truncate(file_size)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return file, nil
}

func download(start_index int64, end_index int64) error {
	ua := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.108 Safari/537.36"
	block_range := "bytes=" + strconv.FormatInt(start_index, 10) + "-" + strconv.FormatInt(end_index, 10)
	ro := &grequests.RequestOptions{Headers: map[string]string{"User-Agent": ua, "Range" : block_range}}
	resp, err := grequests.Get(url, ro)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	if resp.StatusCode == 206 {
		err = writeFile(resp.Bytes(), start_index)
		if err != nil {
			return err
		}
	}
	return nil
}


func writeFile(content []byte, start_index int64) error {
	_, err := download_file.WriteAt(content, start_index)
	if err != nil {
		return err
	}
	return nil
}

func run(block_size int, threads_num int, total_length int)  {
	finish_chan := make(chan int)
	for i := 0; i < threads_num; i++ {
		start_index := i * block_size
		end_index   := (i + 1) * block_size - 1
		if i == threads_num - 1 {
			end_index = total_length
		}
		go func(proc int) {
			err := download(int64(start_index), int64(end_index))
			if err != nil {
				fmt.Println(err.Error())
				finish_chan <- 0
				return
			}
			finish_chan <- proc + 1
		}(i)
	}

	for i := 0; i < threads_num; i++ {
		x := <- finish_chan
			if x == 0 {
				fmt.Println("有错误")
			}else {
				fmt.Println(x, "段下载完成")
			}
	}
	_ = download_file.Close()

}

func main()  {
	threads_num := runtime.NumCPU() + 1
	fmt.Println("线程数" , threads_num)
	url = "http://mirrors.163.com/archlinux/iso/latest/archlinux-2019.05.02-x86_64.iso.torrent"
	download_path = path.Base(url)//todo 文件名检测
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
		content_length, _ := strconv.Atoi(str[0])
		if content_length <= threads_num * 1000 {
			threads_num = 1
		}
		block_size := content_length / threads_num

		f, err := createFile(int64(content_length))
		if err != nil {
			fmt.Println("下载失败")
			fmt.Println(err.Error())
			f.Close()
			return
		}
		download_file = f
		run(block_size, threads_num, content_length)
	}else {
		fmt.Println("下载失败")
	}

}