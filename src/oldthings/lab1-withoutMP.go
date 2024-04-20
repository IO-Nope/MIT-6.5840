package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"syscall"
)

func main() {
	dirname := "./pgs"
	f, err := os.Open(dirname)
	if err != nil {
		log.Fatal(err)
	}
	files, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		log.Fatal(err)
	}

	for _, file_ := range files {

		path := dirname + "/" + file_.Name()
		wPath := "./outputs/" + file_.Name()
		file, e := os.Open(path)
		wPath = strings.Replace(wPath, ".txt", "-WithoutMP.txt", -1)
		wFile, ew := os.OpenFile(wPath, syscall.O_CREAT|syscall.O_RDWR|syscall.O_TRUNC, 0644)
		if ew != nil {
			fmt.Println(ew)
			break
		}
		fmt.Println("\n" + file_.Name() + ":")
		n, _ := wFile.Seek(0, io.SeekEnd)
		_, werr := wFile.WriteAt([]byte(file_.Name()+":\n"), n)
		if werr != nil {
			fmt.Println(werr)
		}
		fmt.Println("Write Done!")
		count := make(map[string]int)
		buf := make([]byte, 1024)
		if e != nil {
			fmt.Println(e)
		}

		for {
			len, _ := file.Read(buf)
			if len == 0 {
				break
			}

			for i := range buf {
				if (buf[i] <= 'Z' && buf[i] >= 'A') || (buf[i] <= 'z' && buf[i] >= 'a') {
					if buf[i] <= 'Z' && buf[i] >= 'A' {
						buf[i] += ('a' - 'A')
					}
					continue
				}
				buf[i] = ' '
			}
			beindex := 0
			chidstr := ""
			for i := range buf {
				if buf[i] != ' ' {
					chidstr += string(buf[i])
					continue
				}
				if buf[i] == ' ' {
					if i == 0 {
						beindex = i
						continue
					}
					if i-1 == beindex {
						beindex = i
						continue
					}
					if value, isOK := count[chidstr]; isOK {
						count[chidstr] = value + 1
					} else {
						count[chidstr] = 1
					}
					chidstr = ""
				}
			}
			timer := 0
			for k, v := range count {
				var tarline = fmt.Sprintf("%10s\t%3d\t", k, v)
				n, _ := wFile.Seek(1, io.SeekEnd)
				_, werr := wFile.WriteAt([]byte(tarline), n)
				if werr != nil {
					fmt.Println(werr)
				}
				timer += 1
				if timer >= 5 {
					n, _ := wFile.Seek(1, io.SeekEnd)
					_, werr := wFile.WriteAt([]byte("\n"), n)
					if werr != nil {
						fmt.Println(werr)
					}
					timer = 0
				}
			}
		}
		file.Close()
		wFile.Close()
	}
}
