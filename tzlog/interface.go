package tzlog

import "fmt"

func D(tpl interface{}, args ...interface{}) {
	fmt.Println(append([]interface{}{tpl}, args...))
}

func I(tpl interface{}, args ...interface{}) {
	fmt.Println(append([]interface{}{tpl}, args...))

}

func E(tpl interface{}, args ...interface{}) {
	fmt.Println(append([]interface{}{tpl}, args...))

}

func W(tpl interface{}, args ...interface{}) {
	fmt.Println(append([]interface{}{tpl}, args...))

}
