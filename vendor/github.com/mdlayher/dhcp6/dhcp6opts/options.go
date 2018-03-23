package dhcp6opts

import (
	"github.com/mdlayher/dhcp6"
)

// GetClientID returns the Client Identifier Option value, as described in RFC
// 3315, Section 22.2.
//
// The DUID returned allows unique identification of a client to a server.
func GetClientID(o dhcp6.Options) (DUID, error) {
	v, err := o.GetOne(dhcp6.OptionClientID)
	if err != nil {
		return nil, err
	}

	return parseDUID(v)
}

// GetServerID returns the Server Identifier Option value, as described in RFC
// 3315, Section 22.3.
//
// The DUID returned allows unique identification of a server to a client.
func GetServerID(o dhcp6.Options) (DUID, error) {
	v, err := o.GetOne(dhcp6.OptionServerID)
	if err != nil {
		return nil, err
	}

	return parseDUID(v)
}

// GetIANA returns the Identity Association for Non-temporary Addresses Option
// value, as described in RFC 3315, Section 22.4.
//
// Multiple IANA values may be present in a single DHCP request.
func GetIANA(o dhcp6.Options) ([]*IANA, error) {
	vv, err := o.Get(dhcp6.OptionIANA)
	if err != nil {
		return nil, err
	}

	// Parse each IA_NA value
	iana := make([]*IANA, len(vv))
	for i := range vv {
		iana[i] = &IANA{}
		if err := iana[i].UnmarshalBinary(vv[i]); err != nil {
			return nil, err
		}
	}
	return iana, nil
}

// GetIATA returns the Identity Association for Temporary Addresses Option
// value, as described in RFC 3315, Section 22.5.
//
// Multiple IATA values may be present in a single DHCP request.
func GetIATA(o dhcp6.Options) ([]*IATA, error) {
	vv, err := o.Get(dhcp6.OptionIATA)
	if err != nil {
		return nil, err
	}

	// Parse each IA_NA value
	iata := make([]*IATA, len(vv))
	for i := range vv {
		iata[i] = &IATA{}
		if err := iata[i].UnmarshalBinary(vv[i]); err != nil {
			return nil, err
		}
	}
	return iata, nil
}

// GetIAAddr returns the Identity Association Address Option value, as described
// in RFC 3315, Section 22.6.
//
// The IAAddr option must always appear encapsulated in the Options map of a
// IANA or IATA option.  Multiple IAAddr values may be present in a single DHCP
// request.
func GetIAAddr(o dhcp6.Options) ([]*IAAddr, error) {
	vv, err := o.Get(dhcp6.OptionIAAddr)
	if err != nil {
		return nil, err
	}

	iaAddr := make([]*IAAddr, len(vv))
	for i := range vv {
		iaAddr[i] = &IAAddr{}
		if err := iaAddr[i].UnmarshalBinary(vv[i]); err != nil {
			return nil, err
		}
	}
	return iaAddr, nil
}

// GetOptionRequest returns the Option Request Option value, as described in
// RFC 3315, Section 22.7.
//
// The slice of OptionCode values indicates the options a DHCP client is
// interested in receiving from a server.
func GetOptionRequest(o dhcp6.Options) (OptionRequestOption, error) {
	v, err := o.GetOne(dhcp6.OptionORO)
	if err != nil {
		return nil, err
	}

	var oro OptionRequestOption
	err = oro.UnmarshalBinary(v)
	return oro, err
}

// GetPreference returns the Preference Option value, as described in RFC 3315,
// Section 22.8.
//
// The integer preference value is sent by a server to a client to affect the
// selection of a server by the client.
func GetPreference(o dhcp6.Options) (Preference, error) {
	v, err := o.GetOne(dhcp6.OptionPreference)
	if err != nil {
		return 0, err
	}

	var p Preference
	err = (&p).UnmarshalBinary(v)
	return p, err
}

// GetElapsedTime returns the Elapsed Time Option value, as described in RFC
// 3315, Section 22.9.
//
// The time.Duration returned reports the time elapsed during a DHCP
// transaction, as reported by a client.
func GetElapsedTime(o dhcp6.Options) (ElapsedTime, error) {
	v, err := o.GetOne(dhcp6.OptionElapsedTime)
	if err != nil {
		return 0, err
	}

	var t ElapsedTime
	err = (&t).UnmarshalBinary(v)
	return t, err
}

// GetRelayMessageOption returns the Relay Message Option value, as described
// in RFC 3315, Section 22.10.
//
// The RelayMessage option carries a DHCP message in a Relay-forward or
// Relay-reply message.
func GetRelayMessageOption(o dhcp6.Options) (RelayMessageOption, error) {
	v, err := o.GetOne(dhcp6.OptionRelayMsg)
	if err != nil {
		return nil, err
	}

	var r RelayMessageOption
	err = (&r).UnmarshalBinary(v)
	return r, err
}

// GetAuthentication returns the Authentication Option value, as described in
// RFC 3315, Section 22.11.
//
// The Authentication option carries authentication information to
// authenticate the identity and contents of DHCP messages.
func GetAuthentication(o dhcp6.Options) (*Authentication, error) {
	v, err := o.GetOne(dhcp6.OptionAuth)
	if err != nil {
		return nil, err
	}

	a := new(Authentication)
	err = a.UnmarshalBinary(v)
	return a, err
}

// GetUnicast returns the IP from a Unicast Option value, described in RFC
// 3315, Section 22.12.
//
// The IP return value indicates a server's IPv6 address, which a client may
// use to contact the server via unicast.
func GetUnicast(o dhcp6.Options) (IP, error) {
	v, err := o.GetOne(dhcp6.OptionUnicast)
	if err != nil {
		return nil, err
	}

	var ip IP
	err = ip.UnmarshalBinary(v)
	return ip, err
}

// GetStatusCode returns the Status Code Option value, described in RFC 3315,
// Section 22.13.
//
// The StatusCode return value may be used to determine a code and an
// explanation for the status.
func GetStatusCode(o dhcp6.Options) (*StatusCode, error) {
	v, err := o.GetOne(dhcp6.OptionStatusCode)
	if err != nil {
		return nil, err
	}

	s := new(StatusCode)
	err = s.UnmarshalBinary(v)
	return s, err
}

// GetRapidCommit returns the Rapid Commit Option value, described in RFC 3315,
// Section 22.14.
//
// Nil is returned if OptionRapidCommit was present in the Options map.
func GetRapidCommit(o dhcp6.Options) error {
	v, err := o.GetOne(dhcp6.OptionRapidCommit)
	if err != nil {
		return err
	}

	// Data must be completely empty; presence of the Rapid Commit option
	// indicates it is requested.
	if len(v) != 0 {
		return dhcp6.ErrInvalidPacket
	}
	return nil
}

// GetUserClass returns the User Class Option value, described in RFC 3315,
// Section 22.15.
//
// The Data structure returned contains any raw class data present in
// the option.
func GetUserClass(o dhcp6.Options) (Data, error) {
	v, err := o.GetOne(dhcp6.OptionUserClass)
	if err != nil {
		return nil, err
	}

	var d Data
	err = d.UnmarshalBinary(v)
	return d, err
}

// GetVendorClass returns the Vendor Class Option value, described in RFC 3315,
// Section 22.16.
//
// The VendorClass structure returned contains VendorClass in
// the option.
func GetVendorClass(o dhcp6.Options) (*VendorClass, error) {
	v, err := o.GetOne(dhcp6.OptionVendorClass)
	if err != nil {
		return nil, err
	}

	vc := new(VendorClass)
	err = vc.UnmarshalBinary(v)
	return vc, err
}

// GetVendorOpts returns the Vendor-specific Information Option value,
// described in RFC 3315, Section 22.17.
//
// The VendorOpts structure returned contains Vendor-specific Information data
// present in the option.
func GetVendorOpts(o dhcp6.Options) (*VendorOpts, error) {
	v, err := o.GetOne(dhcp6.OptionVendorOpts)
	if err != nil {
		return nil, err
	}

	vo := new(VendorOpts)
	err = vo.UnmarshalBinary(v)
	return vo, err
}

// GetInterfaceID returns the Interface-Id Option value, described in RFC 3315,
// Section 22.18.
//
// The InterfaceID structure returned contains any raw class data present in
// the option.
func GetInterfaceID(o dhcp6.Options) (InterfaceID, error) {
	v, err := o.GetOne(dhcp6.OptionInterfaceID)
	if err != nil {
		return nil, err
	}

	var i InterfaceID
	err = i.UnmarshalBinary(v)
	return i, err
}

// GetIAPD returns the Identity Association for Prefix Delegation Option value,
// described in RFC 3633, Section 9.
//
// Multiple IAPD values may be present in a a single DHCP request.
func GetIAPD(o dhcp6.Options) ([]*IAPD, error) {
	vv, err := o.Get(dhcp6.OptionIAPD)
	if err != nil {
		return nil, err
	}

	// Parse each IA_PD value
	iapd := make([]*IAPD, len(vv))
	for i := range vv {
		iapd[i] = &IAPD{}
		if err := iapd[i].UnmarshalBinary(vv[i]); err != nil {
			return nil, err
		}
	}

	return iapd, nil
}

// GetIAPrefix returns the Identity Association Prefix Option value, as
// described in RFC 3633, Section 10.
//
// Multiple IAPrefix values may be present in a a single DHCP request.
func GetIAPrefix(o dhcp6.Options) ([]*IAPrefix, error) {
	vv, err := o.Get(dhcp6.OptionIAPrefix)
	if err != nil {
		return nil, err
	}

	// Parse each IAPrefix value
	iaPrefix := make([]*IAPrefix, len(vv))
	for i := range vv {
		iaPrefix[i] = &IAPrefix{}
		if err := iaPrefix[i].UnmarshalBinary(vv[i]); err != nil {
			return nil, err
		}
	}

	return iaPrefix, nil
}

// GetRemoteIdentifier returns the Remote Identifier, described in RFC 4649.
//
// This option may be added by DHCPv6 relay agents that terminate
// switched or permanent circuits and have mechanisms to identify the
// remote host end of the circuit.
func GetRemoteIdentifier(o dhcp6.Options) (*RemoteIdentifier, error) {
	v, err := o.GetOne(dhcp6.OptionRemoteIdentifier)
	if err != nil {
		return nil, err
	}

	r := new(RemoteIdentifier)
	err = r.UnmarshalBinary(v)
	return r, err
}

// GetBootFileURL returns the Boot File URL Option value, described in RFC
// 5970, Section 3.1.
//
// The URL return value contains a URL which may be used by clients to obtain
// a boot file for PXE.
func GetBootFileURL(o dhcp6.Options) (*URL, error) {
	v, err := o.GetOne(dhcp6.OptionBootFileURL)
	if err != nil {
		return nil, err
	}

	u := new(URL)
	err = u.UnmarshalBinary(v)
	return u, err
}

// GetBootFileParam returns the Boot File Parameters Option value, described in
// RFC 5970, Section 3.2.
//
// The Data structure returned contains any parameters needed for a boot
// file, such as a root filesystem label or a path to a configuration file for
// further chainloading.
func GetBootFileParam(o dhcp6.Options) (BootFileParam, error) {
	v, err := o.GetOne(dhcp6.OptionBootFileParam)
	if err != nil {
		return nil, err
	}

	var bfp BootFileParam
	err = bfp.UnmarshalBinary(v)
	return bfp, err
}

// GetClientArchType returns the Client System Architecture Type Option value,
// described in RFC 5970, Section 3.3.
//
// The ArchTypes slice returned contains a list of one or more ArchType values.
// The first ArchType listed is the client's most preferable value.
func GetClientArchType(o dhcp6.Options) (ArchTypes, error) {
	v, err := o.GetOne(dhcp6.OptionClientArchType)
	if err != nil {
		return nil, err
	}

	var a ArchTypes
	err = a.UnmarshalBinary(v)
	return a, err
}

// GetNII returns the Client Network Interface Identifier Option value,
// described in RFC 5970, Section 3.4.
//
// The NII value returned indicates a client's level of Universal Network
// Device Interface (UNDI) support.
func GetNII(o dhcp6.Options) (*NII, error) {
	v, err := o.GetOne(dhcp6.OptionNII)
	if err != nil {
		return nil, err
	}

	n := new(NII)
	err = n.UnmarshalBinary(v)
	return n, err
}

// GetDNSServers returns the DNS Recursive Name Servers Option value, as
// described in RFC 3646, Section 3.
//
// The DNS servers are listed in the order of preference for use by the client
// resolver.
func GetDNSServers(o dhcp6.Options) (IPs, error) {
	v, err := o.GetOne(dhcp6.OptionDNSServers)
	if err != nil {
		return nil, err
	}

	var ips IPs
	err = ips.UnmarshalBinary(v)
	return ips, err
}
