package gorpitx

const (
	hzToMhzDivisor    = 1000000.0 // conversion factor from Hz to MHz
	kHzToMHzDivisor   = 1000.0    // conversion factor from kHz to MHz
	khzToHzMultiplier = 1000.0    // conversion factor from kHz to Hz
	roundingOffset    = 0.5       // rounding offset for precision check
	decimalPrecision  = 10.0      // for 1 decimal place precision check
)

// hzToMHz converts frequency from hertz to megahertz.
func hzToMHz(hz float64) float64 {
	return hz / hzToMhzDivisor
}

// kHzToMHz converts frequency from kilohertz to megahertz.
func kHzToMHz(kHz float64) float64 {
	return kHz / kHzToMHzDivisor
}

// mHzToKHz converts frequency from megahertz to kilohertz.
func mHzToKHz(mHz float64) float64 {
	return mHz * kHzToMHzDivisor
}

// mHzToHz converts frequency from megahertz to hertz.
func mHzToHz(mHz float64) float64 {
	return mHz * hzToMhzDivisor
}

// getMinFreqHz returns the minimum supported frequency in Hz.
func getMinFreqHz() float64 {
	return float64(minFreqKHz) * khzToHzMultiplier // Convert kHz to Hz
}

// getMaxFreqHz returns the maximum supported frequency in Hz.
func getMaxFreqHz() float64 {
	return float64(maxFreqKHz) * khzToHzMultiplier // Convert kHz to Hz
}

// isValidFreqHz checks if a frequency in Hz is within RPiTX hardware limits.
func isValidFreqHz(freqHz float64) bool {
	return freqHz >= getMinFreqHz() && freqHz <= getMaxFreqHz()
}

// getMinFreqMHzDisplay returns the minimum supported frequency in MHz for
// display purposes.
func getMinFreqMHzDisplay() float64 {
	return kHzToMHz(float64(minFreqKHz))
}

// getMaxFreqMHzDisplay returns the maximum supported frequency in MHz for
// display purposes.
func getMaxFreqMHzDisplay() float64 {
	return kHzToMHz(float64(maxFreqKHz))
}

// hasValidFreqPrecision checks if frequency has acceptable precision.
// pifmrds works best with 1 decimal place (0.1 MHz precision).
func hasValidFreqPrecision(freqMHz float64) bool {
	// Round to 1 decimal place and compare
	rounded := float64(int(freqMHz*decimalPrecision+roundingOffset)) /
		decimalPrecision

	return freqMHz == rounded
}
