package iprange

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleAddress(t *testing.T) {
	ipRange, err := Parse("192.168.1.1")
	assert.Nil(t, err)

	assert.Equal(t, net.IPv4(192, 168, 1, 1).To4(), ipRange.Min)
	assert.Equal(t, ipRange.Min, ipRange.Max)
}

func TestCIDRAddress(t *testing.T) {
	{
		ipRange, err := Parse("192.168.1.1/24")
		assert.Nil(t, err)

		assert.Equal(t, net.IPv4(192, 168, 1, 0).To4(), ipRange.Min)
		assert.Equal(t, net.IPv4(192, 168, 1, 255).To4(), ipRange.Max)
	}

	{
		ipRange, err := Parse("192.168.2.1/24")
		assert.Nil(t, err)

		assert.Equal(t, net.IPv4(192, 168, 2, 0).To4(), ipRange.Min)
		assert.Equal(t, net.IPv4(192, 168, 2, 255).To4(), ipRange.Max)

		out := ipRange.Expand()
		assert.Equal(t, int(0xffffffff-0xffffff00), len(out)-1)
		for i := 0; i < 256; i++ {
			assert.Equal(t, net.IP([]byte{192, 168, 2, byte(i)}), out[i])
		}
	}

	{
		ipRange, err := Parse("10.1.2.3/16")
		assert.Nil(t, err)

		assert.Equal(t, net.IPv4(10, 1, 0, 0).To4(), ipRange.Min)
		assert.Equal(t, net.IPv4(10, 1, 255, 255).To4(), ipRange.Max)

		out := ipRange.Expand()
		assert.Equal(t, int(0xffffffff-0xffff0000), len(out)-1)
		for i := 0; i < 65536; i++ {
			assert.Equal(t, net.IP([]byte{10, 1, byte(i / 256), byte(i % 256)}), out[i])
		}
	}

	{
		ipRange, err := Parse("10.1.2.3/32")
		assert.Nil(t, err)

		assert.Equal(t, net.IPv4(10, 1, 2, 3).To4(), ipRange.Min)
		assert.Equal(t, ipRange.Min, ipRange.Max)
	}
}

func TestWildcardAddress(t *testing.T) {
	ipRange, err := Parse("192.168.1.*")
	assert.Nil(t, err)

	assert.Equal(t, net.IPv4(192, 168, 1, 0).To4(), ipRange.Min)
	assert.Equal(t, net.IPv4(192, 168, 1, 255).To4(), ipRange.Max)
}

func TestRangeAddress(t *testing.T) {
	{
		ipRange, err := Parse("192.168.1.10-20")
		assert.Nil(t, err)

		assert.Equal(t, net.IPv4(192, 168, 1, 10).To4(), ipRange.Min)
		assert.Equal(t, net.IPv4(192, 168, 1, 20).To4(), ipRange.Max)
	}

	{
		ipRange, err := Parse("192.168.10-20.1")
		assert.Nil(t, err)

		assert.Equal(t, net.IPv4(192, 168, 10, 1).To4(), ipRange.Min)
		assert.Equal(t, net.IPv4(192, 168, 20, 1).To4(), ipRange.Max)
	}

	{
		ipRange, err := Parse("0-255.1.1.1")
		assert.Nil(t, err)

		assert.Equal(t, net.IPv4(0, 1, 1, 1).To4(), ipRange.Min)
		assert.Equal(t, net.IPv4(255, 1, 1, 1).To4(), ipRange.Max)
	}

	{
		ipRange, err := Parse("1-2.3-4.5-6.7-8")
		assert.Nil(t, err)

		assert.Equal(t, net.IPv4(1, 3, 5, 7).To4(), ipRange.Min)
		assert.Equal(t, net.IPv4(2, 4, 6, 8).To4(), ipRange.Max)

		out := ipRange.Expand()

		assert.Equal(t, 16, len(out))
		assert.Equal(t, out, []net.IP{
			net.IP([]byte{1, 3, 5, 7}),
			net.IP([]byte{1, 3, 5, 8}),
			net.IP([]byte{1, 3, 6, 7}),
			net.IP([]byte{1, 3, 6, 8}),
			net.IP([]byte{1, 4, 5, 7}),
			net.IP([]byte{1, 4, 5, 8}),
			net.IP([]byte{1, 4, 6, 7}),
			net.IP([]byte{1, 4, 6, 8}),
			net.IP([]byte{2, 3, 5, 7}),
			net.IP([]byte{2, 3, 5, 8}),
			net.IP([]byte{2, 3, 6, 7}),
			net.IP([]byte{2, 3, 6, 8}),
			net.IP([]byte{2, 4, 5, 7}),
			net.IP([]byte{2, 4, 5, 8}),
			net.IP([]byte{2, 4, 6, 7}),
			net.IP([]byte{2, 4, 6, 8}),
		})
	}
}

func TestMixedAddress(t *testing.T) {
	ipRange, err := Parse("192.168.10-20.*/25")
	assert.Nil(t, err)

	assert.Equal(t, net.IPv4(192, 168, 10, 0).To4(), ipRange.Min)
	assert.Equal(t, net.IPv4(192, 168, 10, 127).To4(), ipRange.Max)
}

func TestList(t *testing.T) {
	rangeList, err := ParseList("192.168.1.1, 192.168.1.1/24, 192.168.1.*, 192.168.1.10-20")
	assert.Nil(t, err)
	assert.Len(t, rangeList, 4)

	assert.Equal(t, net.IP([]byte{192, 168, 1, 1}), rangeList[0].Min)
	assert.Equal(t, net.IP([]byte{192, 168, 1, 1}), rangeList[0].Max)

	assert.Equal(t, net.IP([]byte{192, 168, 1, 0}), rangeList[1].Min)
	assert.Equal(t, net.IP([]byte{192, 168, 1, 255}), rangeList[1].Max)

	assert.Equal(t, net.IP([]byte{192, 168, 1, 0}), rangeList[2].Min)
	assert.Equal(t, net.IP([]byte{192, 168, 1, 255}), rangeList[2].Max)

	assert.Equal(t, net.IP([]byte{192, 168, 1, 10}), rangeList[3].Min)
	assert.Equal(t, net.IP([]byte{192, 168, 1, 20}), rangeList[3].Max)
}

func TestBadAddress(t *testing.T) {
	ipRange, err := Parse("192.168.10")
	assert.Nil(t, ipRange)
	assert.Error(t, err)
}

func TestBadList(t *testing.T) {
	rangeList, err := ParseList("192.168.1,, 192.168.1.1/24, 192.168.1.*, 192.168.1.10-20")
	assert.Error(t, err)
	assert.Nil(t, rangeList)
}

func TestListExpansion(t *testing.T) {
	rangeList, err := ParseList("192.168.1.10, 192.168.1.1-20, 192.168.1.10/29")
	assert.Nil(t, err)

	expanded := rangeList.Expand()
	assert.Len(t, expanded, 20)
}
