#import <AppKit/AppKit.h>
#import <Foundation/Foundation.h>

void OpenPath(const char *path){
    @autoreleasepool {
        NSString *folderPath = [[NSString stringWithCString:path encoding:NSUTF8StringEncoding] stringByExpandingTildeInPath];
        NSURL *folderURL = [NSURL fileURLWithPath: folderPath];
        [[NSWorkspace sharedWorkspace] openURL: folderURL];
    }
}

const char* GetAppDataDirectory() {
    @autoreleasepool {
        NSArray *paths = NSSearchPathForDirectoriesInDomains(NSApplicationSupportDirectory, NSUserDomainMask, YES);
        NSString *applicationSupportDirectory = [paths firstObject];
        const char *appDataDir = [applicationSupportDirectory UTF8String];
        return appDataDir;
    }
}