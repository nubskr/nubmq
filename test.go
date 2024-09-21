package main

import (
    "fmt"
	"sync"
	"runtime"
)

func lmao(waitgroup *sync.WaitGroup,chn *chan string) {
	fmt.Println("One started")
	cnt := 0
	for i:= 1; i < 1000000000; i++{
		// fmt.Println("lmao")
		cnt++
	}
	*chn <- "lmao"
	waitgroup.Done()
}

func bruh(waitgroup *sync.WaitGroup,chn *chan string) {
	fmt.Println("two started")
	cnt := 0
	for i:= 1; i < 1000000000; i++{
		// fmt.Println("lmao")
		cnt++
	}
	*chn <- "bruh"
	waitgroup.Done()
}

func lmao_wut(waitgroup *sync.WaitGroup,chn *chan string) {
	for i:= 1; i < 1000000; i++{
		fmt.Println("lmao")
	}
	*chn <- "lmao_wut bro"
	waitgroup.Done()
}

func bruh_wut(waitgroup *sync.WaitGroup,chn *chan string) {
	for i:= 1; i < 1000000; i++{
		fmt.Println("bruh")
	}
	*chn <- "you serious ?"
	waitgroup.Done()
}

func main(){
	runtime.GOMAXPROCS(runtime.NumCPU())
	var wg sync.WaitGroup

	chn := make(chan string)

	wg.Add(2)
	// these lil shits execute in random order, lmao
	go lmao(&wg,&chn)

	go bruh(&wg,&chn)

	one, two := <-chn,<-chn

	fmt.Println(one,two)

	wg.Wait()

	// now we take those shits and add something else to them
	// ----------------------------------------------------

	wg.Add(2)

	// fmt.Println("starting phase 2")

	// go lmao_wut(&wg,&chn)
	// go bruh_wut(&wg,&chn)

	// omg, hewwo:= <-chn,<-chn

	// fmt.Println(omg,hewwo)
}