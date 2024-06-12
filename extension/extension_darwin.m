#import <AppKit/AppKit.h>
#import <Foundation/Foundation.h>

void OpenPath(const char *path){
    @autoreleasepool {
        NSString *folderPath = [[NSString stringWithCString:path encoding:NSUTF8StringEncoding] stringByExpandingTildeInPath];
        NSURL *folderURL = [NSURL fileURLWithPath: folderPath];
        [[NSWorkspace sharedWorkspace] openURL: folderURL];
    }
}
