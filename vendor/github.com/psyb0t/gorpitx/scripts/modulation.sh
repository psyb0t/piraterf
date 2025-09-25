#!/bin/bash

GAIN=${2:-1.0}

case "$1" in
    # AM modes
    "AM")
        csdr convert_s16_f | csdr gain_ff "$GAIN" | csdr dsb_fc | csdr add_dcoffset_cc | csdr agc_ff
        ;;

    # DSB modes
    "DSB")
        csdr convert_s16_f | csdr gain_ff "$GAIN" | csdr dsb_fc | csdr agc_ff
        ;;

    # USB modes
    "USB")
        csdr convert_s16_f | csdr gain_ff "$GAIN" | csdr dsb_fc | csdr bandpass_fir_fft_cc 0.002 0.06 0.01 | csdr agc_ff
        ;;

    # LSB modes
    "LSB")
        csdr convert_s16_f | csdr gain_ff "$GAIN" | csdr dsb_fc | csdr bandpass_fir_fft_cc -0.06 -0.002 0.01 | csdr agc_ff
        ;;

    # FM mode
    "FM")
        csdr convert_s16_f | csdr gain_ff "$GAIN" | csdr fmmod_fc
        ;;

    # Raw conversion
    "RAW")
        csdr convert_s16_f | csdr gain_ff "$GAIN"
        ;;

    *)
        echo "Usage: simple_csdr [MODE] [GAIN]"
        echo ""
        echo "Modes:"
        echo "  AM                             - Amplitude modulation with AGC"
        echo "  DSB                            - Double sideband with AGC (fast, both USB/LSB)"
        echo "  USB                            - Upper sideband with AGC (SLOW on Pi Zero!)"
        echo "  LSB                            - Lower sideband with AGC (SLOW on Pi Zero!)"
        echo "  FM                             - Frequency modulation"
        echo "  RAW                            - Just convert + gain (no AGC)"
        echo ""
        echo "WARNING: USB/LSB modulations use heavy bandpass filtering that causes"
        echo "         latency, weird modulation, and dropouts on Pi Zero."
        echo "         Use DSB for better performance."
        echo ""
        echo "GAIN defaults to 1.0"
        exit 1
        ;;
esac
