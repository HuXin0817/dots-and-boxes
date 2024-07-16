package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#include <Cocoa/Cocoa.h>
#include <stdlib.h>

// setApplicationIconImageFromData sets the application icon using image data in NSData format
void setApplicationIconImageFromData(const char* data, int length) {
    @autoreleasepool {
        NSData *imageData = [NSData dataWithBytes:data length:length];
        NSImage *image = [[NSImage alloc] initWithData:imageData];
        if (image) {
            [NSApp setApplicationIconImage:image];
        }
    }
}

// initializeNSApplication initializes the shared NSApplication instance
void initializeNSApplication() {
    @autoreleasepool {
        [NSApplication sharedApplication];
    }
}
*/
import "C"

import (
	"unsafe"
)

// init function initializes the NSApplication and sets the RefreshMacOSIcon function
func init() {
	// Call the initializeNSApplication function to initialize the shared NSApplication instance
	C.initializeNSApplication()

	// Define the RefreshMacOSIcon function to set the application icon using the provided image data
	RefreshMacOSIcon = func(data []byte) {
		// Convert the Go byte slice to a C pointer
		cData := C.CBytes(data)
		// Ensure the C memory is freed after use
		defer C.free(unsafe.Pointer(cData))
		// Call the setApplicationIconImageFromData function to set the application icon
		C.setApplicationIconImageFromData((*C.char)(cData), C.int(len(data)))
	}
}
