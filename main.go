package main

import (
	"fmt"
	"gopump/core"
)

type Car struct {
	Weight int
	Name   string
}

type RedCar struct {
	Car
}

func (c *Car) String() {
	fmt.Println("lala")
}

func main() {
	//fmt.Println(os.Getenv("path"))
	srv := core.NewServer()
	srv.AddRouter("/test", test)
	srv.Start(10001)
	// var a RedCar
	// a.Weight = 10
	// a.Name = "ttt"
	// a.String()

}

func test(s *core.Sender) {
	//fmt.Println(len(s.Reqs))
	index := -1
	for k, v := range s.Reqs {
		if values, ok := v.URL.Query()["id"]; ok {
			if values[0] == "lala" {
				index = k
			}
		}
	}
	//fmt.Println(string(s.Context.Data))

	//s.Context.Data = []byte("我是谁")
	if index != -1 {
		s.Reqs = append(s.Reqs[:index], s.Reqs[index+1:]...)
	}
}