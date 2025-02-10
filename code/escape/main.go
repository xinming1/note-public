package main

// 参考:内存逃逸分析 https://www.yuque.com/aceld/golang/yyrlis, 详见note/内存逃逸分析.md

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
	//mainVal := foo(666)
	//println(*mainVal, mainVal)

	mainValUseNew := fooUseNew(666)
	println(*mainValUseNew, mainValUseNew)
}
