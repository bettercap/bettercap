package modules

type ByChannelSorter []*WiFiStation

func (a ByChannelSorter) Len() int      { return len(a) }
func (a ByChannelSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByChannelSorter) Less(i, j int) bool {
	if a[i].Channel == a[j].Channel {
		return a[i].HwAddress < a[j].HwAddress
	}
	return a[i].Channel < a[j].Channel
}

type ByEssidSorter []*WiFiStation

func (a ByEssidSorter) Len() int      { return len(a) }
func (a ByEssidSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByEssidSorter) Less(i, j int) bool {
	if a[i].ESSID() == a[j].ESSID() {
		return a[i].HwAddress < a[j].HwAddress
	}
	return a[i].ESSID() < a[j].ESSID()
}

type BywifiSeenSorter []*WiFiStation

func (a BywifiSeenSorter) Len() int      { return len(a) }
func (a BywifiSeenSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a BywifiSeenSorter) Less(i, j int) bool {
	return a[i].LastSeen.After(a[j].LastSeen)
}
