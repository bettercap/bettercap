#import <Foundation/Foundation.h>
#import <CoreWLAN/CoreWLAN.h>

const char *GetSupportedFrequencies(const char *iface) {
    @autoreleasepool {
        NSString *interfaceName = [NSString stringWithUTF8String:iface];
        CWInterface *interface = [CWInterface interfaceWithName:interfaceName];
        if (!interface) {
            return NULL;
        }

        NSSet *supportedChannels = [interface supportedWLANChannels];
        NSMutableArray *frequencies = [NSMutableArray arrayWithCapacity:[supportedChannels count]];

        for (CWChannel *channel in supportedChannels) {
            [frequencies addObject:@(channel.frequency)];
        }

        NSError *error = nil;
        NSData *jsonData = [NSJSONSerialization dataWithJSONObject:frequencies options:0 error:&error];
        if (!jsonData) {
            NSLog(@"Failed to serialize frequencies: %@", error);
            return NULL;
        }

        NSString *jsonString = [[NSString alloc] initWithData:jsonData encoding:NSUTF8StringEncoding];
        return strdup([jsonString UTF8String]);
    }
}

bool SetInterfaceChannel(const char *iface, int channel) {
    @autoreleasepool {
        NSString *interfaceName = [NSString stringWithUTF8String:iface];
        CWInterface *interface = [CWInterface interfaceWithName:interfaceName];
        if (!interface) {
            return false;
        }

        NSError *error = nil;
        CWChannel *newChannel = [[CWChannel alloc] initWithChannelNumber:channel channelWidth:kCWChannelWidthUnknown];
        [interface setWLANChannel:newChannel error:&error];
        if (error) {
            NSLog(@"Failed to set channel: %@", error);
            return false;
        }

        return true;
    }
}
