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

var CImagePath = C.CString(ImagePath)

func init() {
	C.initializeNSApplication()
	RefreshMacOSIcon = func() {
		C.setApplicationIconImage(CImagePath)
	}
}
