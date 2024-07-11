//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#include <Cocoa/Cocoa.h>
#include <stdlib.h>

void setApplicationIconImage(const char* path) {
    @autoreleasepool {
        NSString *nsPath = [NSString stringWithUTF8String:path];
        NSImage *image = [[NSImage alloc] initWithContentsOfFile:nsPath];
        if (image) {
            [NSApp setApplicationIconImage:image];
        }
    }
}

void initializeNSApplication() {
    @autoreleasepool {
        [NSApplication sharedApplication];
    }
}
*/
import "C"

import (
	"log"
	"runtime"
	"time"
	"unsafe"
)

func init() {
	if runtime.GOOS != "darwin" {
		log.Println("This application is only supported on macOS.")
		return
	}
	C.initializeNSApplication()
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		for range ticker.C {
			path := C.CString(ImagePath)
			C.setApplicationIconImage(path)
			C.free(unsafe.Pointer(path))
		}
	}()
}
