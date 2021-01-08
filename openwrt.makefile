include $(TOPDIR)/rules.mk

PKG_NAME:=bettercap
PKG_VERSION:=2.28
PKG_RELEASE:=2

GO_PKG:=github.com/bettercap/bettercap

PKG_SOURCE:=$(PKG_NAME)-$(PKG_VERSION).tar.gz
PKG_SOURCE_URL:=https://codeload.github.com/bettercap/bettercap/tar.gz/v${PKG_VERSION}?
PKG_HASH:=5bde85117679c6ed8b5469a5271cdd5f7e541bd9187b8d0f26dee790c37e36e9
PKG_BUILD_DIR:=$(BUILD_DIR)/$(PKG_NAME)-$(PKG_VERSION)

PKG_LICENSE:=GPL-3.0
PKG_LICENSE_FILES:=LICENSE.md
PKG_MAINTAINER:=Dylan Corrales <deathcamel57@gmail.com>

PKG_BUILD_DEPENDS:=golang/host
PKG_BUILD_PARALLEL:=1
PKG_USE_MIPS16:=0

include $(INCLUDE_DIR)/package.mk
include ../../../packages/lang/golang/golang-package.mk

define Package/bettercap/Default
  TITLE:=The Swiss Army knife for 802.11, BLE and Ethernet networks reconnaissance and MITM attacks.
  URL:=https://www.bettercap.org/
  DEPENDS:=$(GO_ARCH_DEPENDS) libpcap libusb-1.0
endef

define Package/bettercap
$(call Package/bettercap/Default)
  SECTION:=net
  CATEGORY:=Network
endef

define Package/bettercap/description
  bettercap is a powerful, easily extensible and portable framework written
  in Go which aims to offer to security researchers, red teamers and reverse
  engineers an easy to use, all-in-one solution with all the features they
  might possibly need for performing reconnaissance and attacking WiFi
  networks, Bluetooth Low Energy devices, wireless HID devices and Ethernet networks.
endef

define Package/bettercap/install
	$(call GoPackage/Package/Install/Bin,$(PKG_INSTALL_DIR))
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) $(PKG_INSTALL_DIR)/usr/bin/bettercap  $(1)/usr/bin/bettercap
endef

$(eval $(call GoBinPackage,bettercap))
$(eval $(call BuildPackage,bettercap))