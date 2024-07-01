package main

import "fmt"

// Speaker 接口
type Speaker interface {
	Speak() string
}

// Namer 接口
type Namer interface {
	Name() string
}

// Greeter 接口，组合了 Speaker 和 Namer 接口
type Greeter interface {
	Speaker
	Namer
}

// Dog 结构体
type Dog struct {
	name string
}

// Dog 的 Speak 方法
func (d *Dog) Speak() string {
	return fmt.Sprintf("%s says woof!", d.name)
}

// Dog 的 Name 方法
func (d *Dog) Name() string {
	return d.name
}

// GermanShepherd 结构体，组合了 Dog
type GermanShepherd struct {
	Dog
}

// Labrador 结构体，组合了 Dog
type Labrador struct {
	Dog
}

// 使用 Greeter 接口的函数
func GreetAndSpeak(g Greeter) {
	fmt.Println("Hello", g.Name())
	fmt.Println(g.Speak())
}

func main() {
	gs := GermanShepherd{Dog{name: "Rex"}}
	lb := Labrador{Dog{name: "Buddy"}}

	GreetAndSpeak(&gs)
	GreetAndSpeak(&lb)
}
