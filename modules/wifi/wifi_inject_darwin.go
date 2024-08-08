package wifi

func (mod *WiFiModule) injectPacket(data []byte) {
	mod.Error("wifi frame injection is not supported on macOS (see https://github.com/bettercap/bettercap/issues/448)")
}
