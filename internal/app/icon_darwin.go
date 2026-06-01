//go:build darwin

package app

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#include <stdlib.h>
#import <Cocoa/Cocoa.h>

static void musaSetDockIcon(const char *path) {
	@autoreleasepool {
		NSString *p = [NSString stringWithUTF8String:path];
		NSImage *img = [[NSImage alloc] initWithContentsOfFile:p];
		if (img != nil) {
			[NSApp setApplicationIconImage:img];
		}
	}
}
*/
import "C"
import "unsafe"

func setDockIcon(path string) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	C.musaSetDockIcon(cpath)
}
