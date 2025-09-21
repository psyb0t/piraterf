#!/bin/bash
set -e

# Script parameters
BAUD_RATE="$1"
FREQUENCY="$2"

# Validate parameters
if [ -z "$BAUD_RATE" ] || [ -z "$FREQUENCY" ]; then
    echo "Usage: $0 <baud_rate> <frequency_hz>" >&2
    exit 1
fi

# Generate unique temp file
TEMP_FILE="/tmp/fsk_$$.wav"

# Cleanup function
cleanup() {
    rm -f "$TEMP_FILE"
}
trap cleanup EXIT

# Process pipeline with progress reporting
echo "Encoding input to FSK audio at ${BAUD_RATE} baud..."
if ! cat | minimodem --tx "$BAUD_RATE" -f "$TEMP_FILE"; then
    echo "Failed to encode input to FSK audio" >&2
    exit 1
fi

echo "Converting to 16-bit 48kHz stereo and transmitting at ${FREQUENCY} Hz..."
if ! sox "$TEMP_FILE" -t raw -e signed -b 16 -r 48000 -c 2 - | "${RPITX_PATH}/sendiq" -i /dev/stdin -s 48000 -f "$FREQUENCY" -t i16; then
    echo "Failed to convert and transmit FSK data" >&2
    exit 1
fi

echo "FSK transmission completed successfully"