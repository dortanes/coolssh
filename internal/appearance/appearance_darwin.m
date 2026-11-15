//go:build darwin && cgo

#import <AppKit/AppKit.h>

#include "appearance_darwin.h"

// Windows without an appearance of their own, and their effect views, inherit the application's;
// WKWebView reports it as prefers-color-scheme.
void coolssh_appearance_set(coolssh_appearance appearance) {
    dispatch_block_t set = ^{
        NSAppearance *chosen = nil;
        if (appearance == coolssh_appearance_light) {
            chosen = [NSAppearance appearanceNamed:NSAppearanceNameAqua];
        } else if (appearance == coolssh_appearance_dark) {
            chosen = [NSAppearance appearanceNamed:NSAppearanceNameDarkAqua];
        }
        [NSApplication sharedApplication].appearance = chosen;
    };
    if ([NSThread isMainThread]) {
        set();
    } else {
        dispatch_async(dispatch_get_main_queue(), set);
    }
}
