//
//  MobileEbitenViewWithErrorHandling.m
//  goinovation
//
//  Created by Hajime Hoshi on 2019/08/18.
//  Copyright © 2019 Hajime Hoshi. All rights reserved.
//

#import "MobileEbitenViewControllerWithErrorHandling.h"

#import <Foundation/Foundation.h>

@implementation MobileEbitenViewControllerWithErrorHandling {
}

- (void)onErrorOnGameUpdate:(NSError*)err {
    // You can define your own error handling e.g., using Crashlytics.
    NSLog(@"Inovation Error!: %@", err);
}

@end
