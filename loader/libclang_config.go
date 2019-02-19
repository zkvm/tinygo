// +build !byollvm

package loader

/*
#cgo linux  CFLAGS: -I/usr/lib/llvm-7/include
#cgo linux  LDFLAGS: -L/usr/lib/llvm-7/lib -lclang
#cgo darwin CFLAGS: -I/usr/local/opt/llvm/include
#cgo darwin LDFLAGS: -L/usr/local/opt/llvm/lib -lclang -lffi
#cgo windows CFLAGS: -I"C:/Program Files/LLVM/include"
#cgo windows LDFLAGS: -L"C:/Program Files/LLVM/lib" -lclang
*/
import "C"
