//go:build darwin && cgo

#import <AppKit/AppKit.h>

#include "appicon_darwin.h"

// The PNG is malloc'd for the caller to free; NULL when the app has no icon.
void *coolssh_appicon_png(int side, int *length) {
    __block void *bytes = NULL;
    dispatch_block_t render = ^{
        @autoreleasepool {
            NSImage *icon = [NSApplication sharedApplication].applicationIconImage;
            if (icon == nil) {
                return;
            }
            NSBitmapImageRep *bitmap = [[NSBitmapImageRep alloc]
                initWithBitmapDataPlanes:NULL
                              pixelsWide:side
                              pixelsHigh:side
                           bitsPerSample:8
                         samplesPerPixel:4
                                hasAlpha:YES
                                isPlanar:NO
                          colorSpaceName:NSDeviceRGBColorSpace
                             bytesPerRow:0
                            bitsPerPixel:0];
            [NSGraphicsContext saveGraphicsState];
            [NSGraphicsContext setCurrentContext:[NSGraphicsContext graphicsContextWithBitmapImageRep:bitmap]];
            [icon drawInRect:NSMakeRect(0, 0, side, side)];
            [NSGraphicsContext restoreGraphicsState];
            NSData *png = [bitmap representationUsingType:NSBitmapImageFileTypePNG properties:@{}];
            bytes = malloc(png.length);
            memcpy(bytes, png.bytes, png.length);
            *length = (int)png.length;
        }
    };
    if ([NSThread isMainThread]) {
        render();
    } else {
        dispatch_sync(dispatch_get_main_queue(), render);
    }
    return bytes;
}
