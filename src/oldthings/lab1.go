package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

var wordCh = make(chan string, 5)
var wCount = make(map[string]int)
var mu sync.Mutex
var wg sync.WaitGroup
var done = false
var donech = make(chan bool)
var beginch = make(chan bool)

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
	var tarFs []string
	for i := range files {
		tarFs = append(tarFs, dirname+"/"+files[i].Name())
	}
	for i := range tarFs {
		fContext, err := os.ReadFile(tarFs[i])
		if err != nil {
			log.Fatal(err)
			continue
		}
		go Map(tarFs[i], string(fContext))
	}
	go func() {
		for range tarFs {
			<-donech
		}

		wg.Wait()
		donech <- true
	}()
Loop:
	for {
		select {
		case word := <-wordCh:
			go Reduce(word, "1")
		case <-donech:
			break Loop
		}
	}
	timer := 0
	for k, v := range wCount {
		fmt.Printf("\t%s : %d\t", k, v)
		timer += 1
		if timer >= 5 {
			fmt.Println()
			timer = 0
		}
	}
}

// key files name, val context
func Map(key string, val string) {
	valByte := []byte(val)
	for i := range valByte {
		if (valByte[i] <= 'Z' && valByte[i] >= 'A') || (valByte[i] <= 'z' && valByte[i] >= 'a') {
			if valByte[i] <= 'Z' && valByte[i] >= 'A' {
				valByte[i] += ('a' - 'A')
			}
			continue
		}
		valByte[i] = ' '
	}
	val = string(valByte)
	words := strings.Split(val, " ")
	for i := range words {
		if words[i] != "" {
			wg.Add(1)
			go func(k string, val string) {
				wordCh <- k
				fmt.Println(k)
			}(words[i], "1")
		}
	}
	donech <- true
}

// key words, val +1
func Reduce(key string, val string) {
	mu.Lock()
	defer mu.Unlock()
	defer wg.Done()
	wCount[key] += 1
}
