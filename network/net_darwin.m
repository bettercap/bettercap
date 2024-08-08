#import <Foundation/Foundation.h>
#import <CoreWLAN/CoreWLAN.h>

// The go side of things expects frequencies.
int chan2freq(int channel) {
    if(channel <= 13){ 
        return ((channel - 1) * 5) + 2412;
	} else if(channel == 14) {
		return 2484;
	} else if(channel <= 173) {
		return ((channel - 7) * 5) + 5035;
	} else if(channel == 177) {
		return 5885;
	}
    return 0;
}

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
            // The go side of things expects frequencies.
            [frequencies addObject:@(chan2freq(channel.channelNumber))];
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
        NSSet *supportedChannels = [interface supportedWLANChannels];
        for (CWChannel * channelObj in supportedChannels) {
            // it looks like we can't directly build a CWChannel object anymore
            if ([channelObj channelNumber] == channel) {
                [interface setWLANChannel:channelObj error:nil];
                 if (error) {
                    NSLog(@"Failed to set channel: %@", error);
                    return false;
                }
                return true;
            }
        }

        NSLog(@"channel %d not supported", channel);

        return false;
    }
}