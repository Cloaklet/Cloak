#import <AppKit/AppKit.h>

void RevealInFinder(const char * path){
    @autoreleasepool {
        NSString *filePath = [[NSString stringWithCString:path encoding:NSUTF8StringEncoding] stringByExpandingTildeInPath];
        NSMutableArray *urls = [NSMutableArray arrayWithCapacity:1];
        [urls addObject:[[NSURL fileURLWithPath:filePath] absoluteURL]];
        [[NSWorkspace sharedWorkspace] activateFileViewerSelectingURLs:urls];
    }
}