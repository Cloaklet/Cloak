#import <AppKit/AppKit.h>
#import <Foundation/Foundation.h>

void RevealInFinder(const char *path){
    @autoreleasepool {
        NSString *filePath = [[NSString stringWithCString:path encoding:NSUTF8StringEncoding] stringByExpandingTildeInPath];
        NSMutableArray *urls = [NSMutableArray arrayWithCapacity:1];
        [urls addObject:[[NSURL fileURLWithPath:filePath] absoluteURL]];
        [[NSWorkspace sharedWorkspace] activateFileViewerSelectingURLs:urls];
    }
}

const char* GetAppDataDirectory() {
    NSArray *paths = NSSearchPathForDirectoriesInDomains(NSApplicationSupportDirectory, NSUserDomainMask, YES);
    NSString *applicationSupportDirectory = [paths firstObject];
    const char *appDataDir = [applicationSupportDirectory UTF8String];
    return appDataDir;
}