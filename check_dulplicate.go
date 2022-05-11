package main

import (
	"fmt"
	"image"
	"os"
	"runtime"
	"strconv"
	"sync"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"

	"github.com/corona10/goimagehash"
)

type imagecheck struct {
	name string
	dh   *goimagehash.ImageHash
}

func (ic *imagecheck) String() string {
	return ic.name
}

func main() {
	throttle, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}
	imgs, err := os.ReadDir("imgs")
	if err != nil {
		panic(err)
	}
	err = os.Chdir("imgs")
	if err != nil {
		panic(err)
	}
	chklst := make([]imagecheck, 0, len(imgs))
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	part := len(imgs) / runtime.NumCPU()
	wg.Add(runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		from := i * part
		to := (i + 1) * part
		if to > len(imgs) {
			to = len(imgs)
		}
		go func(from, to int) {
			for i := from; i < to; i++ {
				img := imgs[i]
				if !img.IsDir() {
					f, err := os.Open(img.Name())
					if err != nil {
						panic(err)
					}
					im, _, err := image.Decode(f)
					if err != nil {
						panic(err)
					}
					dh, err := goimagehash.DifferenceHash(im)
					if err != nil {
						panic(err)
					}
					mu.Lock()
					chklst = append(chklst, imagecheck{
						name: img.Name(),
						dh:   dh,
					})
					mu.Unlock()
					_ = f.Close()
				}
			}
			wg.Done()
		}(from, to)
	}
	wg.Wait()
	fmt.Println("read file success")
	dups := make([][]*imagecheck, len(chklst))
	for i := 0; i < len(chklst); i++ {
		for j := len(chklst) - 1; j > i; j-- {
			dis, err := chklst[i].dh.Distance(chklst[j].dh)
			if err != nil {
				panic(err)
			}
			if dis < throttle {
				dups[i] = append(dups[i], &chklst[j])
			}
		}
	}
	hasfound := false
	for i, lst := range dups {
		if len(lst) > 0 {
			lst = append(lst, &chklst[i])
			fmt.Println("found dulplicate: ", lst)
			hasfound = true
		}
	}
	if hasfound {
		os.Exit(1)
	}
	os.Exit(0)
}
