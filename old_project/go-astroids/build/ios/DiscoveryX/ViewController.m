#import "ViewController.h"
#import <Foundation/Foundation.h>

@implementation ViewController

- (void)onErrorOnGameUpdate:(NSError*)err {
    // Custom error handling for game updates
    NSLog(@"DiscoveryX Error: %@", err);
}

@end
