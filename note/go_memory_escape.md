## 内存逃逸分析 原文链接：https://www.yuque.com/aceld/golang/yyrlis

### 示例1, 局部变量逃逸分析
#### 代码
```go
package main

//go:noinline
func foo(argVal int) *int {

	var fooVal1 int = 11
	var fooVal2 int = 12
	var fooVal3 int = 13
	var fooVal4 int = 14
	var fooVal5 int = 15

	println(&argVal, &fooVal1, &fooVal2, &fooVal3, &fooVal4, &fooVal5)

	//返回foo_val3给main函数
	return &fooVal3
}

func main() {
	mainVal := foo(666)
	println(*mainVal, mainVal)
}
```
#### 结论: fooVal3 内存逃逸
```shell
go run -gcflags "-m -l" main.go
# command-line-arguments
./main.go:10:6: moved to heap: fooVal3
./main.go:23:24: new(int) does not escape
./main.go:24:24: new(int) does not escape
./main.go:25:24: new(int) escapes to heap
./main.go:26:24: new(int) does not escape
./main.go:27:24: new(int) does not escape
0xc000073f28 0xc000073f08 0xc000073f00 0xc00000a078 0xc000073ef8 0xc000073ef0
13 0xc00000a078

```

### 示例2, 使用new创建的变量, 逃逸分析
#### 代码
```go
package main
//go:noinline
func fooUseNew(argVal int) *int {

	var fooVal1 *int = new(int)
	var fooVal2 *int = new(int)
	var fooVal3 *int = new(int)
	var fooVal4 *int = new(int)
	var fooVal5 *int = new(int)

	println(&argVal, &fooVal1, &fooVal2, &fooVal3, &fooVal4, &fooVal5)

	//返回foo_val3给main函数
	return fooVal3
}

func main() {
	mainValUseNew := fooUseNew(666)
	println(*mainValUseNew, mainValUseNew)
}

```

#### 结论：new创建的局部变量也会逃逸
```shell
   go run -gcflags "-m -l" main.go
# command-line-arguments
./main.go:10:6: moved to heap: fooVal3
./main.go:23:24: new(int) does not escape
./main.go:24:24: new(int) does not escape
./main.go:25:24: new(int) escapes to heap
./main.go:26:24: new(int) does not escape
./main.go:27:24: new(int) does not escape
0xc000073f28 0xc000073f10 0xc000073f08 0xc000073f00 0xc000073ef8 0xc000073ef0
0 0xc00000a078

```

### 示例3，[]interface{}数据类型，通过[]赋值必定会出现逃逸。
#### 代码
```go
package main

func main() {
	data := []interface{}{100, 200}
	data[0] = 100
}
```
#### 结论
```shell
   go run -gcflags "-m -l" m3.go
# command-line-arguments
./m3.go:4:23: []interface {}{...} does not escape
./m3.go:4:24: 100 does not escape
./m3.go:4:29: 200 does not escape
./m3.go:5:12: 100 escapes to heap
```

### 示例4，map[string]interface{}类型尝试通过赋值，必定会出现逃逸。
#### 代码
```go
package main

func main() {
	data := make(map[string]interface{})
	data["key"] = 200
}
```
#### 结论
```shell
   go run -gcflags "-m -l" m4.go 
# command-line-arguments
./m4.go:4:14: make(map[string]interface {}) does not escape
./m4.go:5:16: 200 escapes to heap

```

### 示例5，map[interface{}]interface{}类型尝试通过赋值，会导致key和value的赋值，出现逃逸。
#### 代码
```go
package main

func main() {
    data := make(map[interface{}]interface{})
    data[100] = 200
}
```
#### 结论
```shell
   go run -gcflags "-m -l" m5.go
# command-line-arguments
./m5.go:4:14: make(map[interface {}]interface {}) does not escape
./m5.go:5:7: 100 escapes to heap
./m5.go:5:14: 200 escapes to heap
```

### 示例6，map[string][]string数据类型，赋值会发生[]string发生逃逸。
#### 代码
```go
package main

func main() {
    data := make(map[string][]string)
    data["key"] = []string{"value"}
}
```
#### 结论
```shell
   go run -gcflags "-m -l" m6.go
# command-line-arguments
./m6.go:4:14: make(map[string][]string) does not escape
./m6.go:5:24: []string{...} escapes to heap
```

### 示例7，[]*int数据类型，赋值的右值会发生逃逸现象。
#### 代码
```go
package main

func main() {
	a := 10
	data := []*int{nil}
	data[0] = &a
}
```
#### 结论
```shell
   go run -gcflags "-m -l" m7.go
# command-line-arguments
./m7.go:4:2: moved to heap: a
./m7.go:5:16: []*int{...} does not escape
```

### 示例8，func(*int)函数类型，进行函数赋值，会使传递的形参出现逃逸现象。
#### 代码
```go
package main

import "fmt"

func foo8(a *int) {
	return
}

func main() {
	data := 10
	f := foo8
	f(&data)
	fmt.Println(data)
}

```
#### 结论
```shell
   go run -gcflags "-m -l" m8.go
# command-line-arguments
./m8.go:5:11: a does not escape
./m8.go:13:13: ... argument does not escape
./m8.go:13:14: data escapes to heap
10
```


### 示例9，func([]string): 函数类型，进行[]string{"value"}赋值，会使传递的参数出现逃逸现象。
#### 代码
```go
package main

import "fmt"

func foo9(a []string) {
	return
}

func main() {
	s := []string{"aceld"}
	foo9(s)
	fmt.Println(s)
}

```
#### 结论
```shell
   go run -gcflags "-m -l" m9.go
# command-line-arguments
./m9.go:5:11: a does not escape
./m9.go:10:15: []string{...} escapes to heap
./m9.go:12:13: ... argument does not escape
./m9.go:12:14: s escapes to heap
[aceld]
```


### 示例10，chan []string数据类型，想当前channel中传输[]string{"value"}会发生逃逸现象。
#### 代码
```go
package main

func main() {
	ch := make(chan []string)

	s := []string{"aceld"}

	go func() {
		ch <- s
	}()
}
```
#### 结论
```shell
   go run -gcflags "-m -l" m10.go
# command-line-arguments
./m10.go:6:15: []string{...} escapes to heap
./m10.go:8:5: func literal escapes to heap
```
