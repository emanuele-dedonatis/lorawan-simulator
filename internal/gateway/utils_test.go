package gateway

import (
	"testing"

	"github.com/brocaar/lorawan"
	"github.com/stretchr/testify/assert"
)

func TestFormatEUI(t *testing.T) {
	t.Run("formats EUI64 with dashes", func(t *testing.T) {
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		result := formatEUI(eui)
		expected := "01-02-03-04-05-06-07-08"
		assert.Equal(t, expected, result)
	})

	t.Run("formats EUI64 with hex letters in lowercase", func(t *testing.T) {
		eui := lorawan.EUI64{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11}
		result := formatEUI(eui)
		expected := "aa-bb-cc-dd-ee-ff-00-11"
		assert.Equal(t, expected, result)
	})

	t.Run("formats EUI64 with all zeros", func(t *testing.T) {
		eui := lorawan.EUI64{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
		result := formatEUI(eui)
		expected := "00-00-00-00-00-00-00-00"
		assert.Equal(t, expected, result)
	})

	t.Run("formats EUI64 with all 0xff", func(t *testing.T) {
		eui := lorawan.EUI64{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
		result := formatEUI(eui)
		expected := "ff-ff-ff-ff-ff-ff-ff-ff"
		assert.Equal(t, expected, result)
	})

	t.Run("formats EUI64 with mixed values", func(t *testing.T) {
		eui := lorawan.EUI64{0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0}
		result := formatEUI(eui)
		expected := "12-34-56-78-9a-bc-de-f0"
		assert.Equal(t, expected, result)
	})

	t.Run("pads single digit hex values with leading zero", func(t *testing.T) {
		eui := lorawan.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
		result := formatEUI(eui)
		// Should be "01-02..." not "1-2..."
		assert.Contains(t, result, "01-02-03")
	})
}
