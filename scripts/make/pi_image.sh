#!/bin/bash

# Pi Zero W Image Cloner & Shrinker - Clone Your Fucking RF Beast
# Clones and shrinks SD cards for maximum pirate distribution

set -e

# Source common functions and config
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

# Additional color codes
CYAN='\033[0;36m'

# Pirate-style logging functions
log_info() {
    echo -e "${BLUE}🔍 [INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}✅ [SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠️  [WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}💥 [ERROR]${NC} $1"
}

log_section() {
    echo ""
    echo -e "${CYAN}🏴‍☠️ === $1 === 🏴‍☠️${NC}"
    echo ""
}

log_section "PI Zero W Image Cloner & Shrinker - Let's Fucking Clone!"
log_info "Time to replicate your RF weapon for maximum chaos distribution!"

# Function to detect RPi devices
detect_fucking_rpi_devices() {
    log_section "Scanning for RPi Devices"
    log_info "Looking for your fucking RF-ready SD cards..."

    local devices=()
    local device_info=()

    # Get all block devices
    for device in /dev/sd* /dev/mmcblk*; do
        if [[ -b "$device" && ! "$device" =~ [0-9]$ ]]; then
            # Check partitions for this device
            local has_rootfs=false
            local has_bootfs=false
            local size=""

            # Get device size
            if command -v lsblk >/dev/null 2>&1; then
                size=$(lsblk -n -o SIZE "$device" 2>/dev/null | head -1 | tr -d ' ')
            fi

            # Check partition labels and types (more robust detection)
            for partition in ${device}*; do
                if [[ -b "$partition" ]]; then
                    local label=$(lsblk -n -o LABEL "$partition" 2>/dev/null)
                    local fstype=$(lsblk -n -o FSTYPE "$partition" 2>/dev/null)
                    local ptype=$(lsblk -n -o PARTTYPE "$partition" 2>/dev/null)

                    # Check for rootfs partition (ext4 on partition 2, or specific label/type)
                    if [[ "$label" == "rootfs" ]] ||
                       [[ "$fstype" == "ext4" && "$partition" =~ (2|p2)$ ]] ||
                       [[ "$ptype" == "0fc63daf-8483-4772-8e79-3d69d8477de4" ]]; then
                        has_rootfs=true
                    fi

                    # Check for bootfs partition (vfat on partition 1, or specific label/type)
                    if [[ "$label" == "bootfs" ]] ||
                       [[ "$fstype" == "vfat" && "$partition" =~ (1|p1)$ ]] ||
                       [[ "$fstype" == "fat32" && "$partition" =~ (1|p1)$ ]] ||
                       [[ "$ptype" == "c12a7328-f81f-11d2-ba4b-00a0c93ec93b" ]]; then
                        has_bootfs=true
                    fi
                fi
            done

            # If filesystem detection fails, try blkid as fallback
            if [[ ! $has_rootfs || ! $has_bootfs ]]; then
                for partition in ${device}*; do
                    if [[ -b "$partition" ]]; then
                        local blkid_info=$(sudo blkid "$partition" 2>/dev/null || true)

                        # Check for typical Pi patterns in blkid output
                        if [[ "$blkid_info" =~ LABEL=\"rootfs\" ]] ||
                           [[ "$blkid_info" =~ TYPE=\"ext4\" && "$partition" =~ (2|p2)$ ]]; then
                            has_rootfs=true
                        fi

                        if [[ "$blkid_info" =~ LABEL=\"bootfs\" ]] ||
                           [[ "$blkid_info" =~ TYPE=\"vfat\" && "$partition" =~ (1|p1)$ ]]; then
                            has_bootfs=true
                        fi
                    fi
                done
            fi

            if $has_rootfs && $has_bootfs; then
                devices+=("$device")
                device_info+=("$device ($size)")
                log_success "Found RF beast: $device ($size)"
            fi
        fi
    done

    echo

    if [ ${#devices[@]} -eq 0 ]; then
        log_error "No fucking RPi devices detected!"
        log_warning "Make sure your SD card/USB device is connected and contains a Raspbian installation."
        exit 1
    fi

    # Display menu
    log_info "Select device to clone:"
    for i in "${!device_info[@]}"; do
        echo "  $((i+1))) ${device_info[$i]}"
    done
    echo "  0) Manual entry"
    echo "  q) Quit"
    echo

    while true; do
        read -p "Enter choice: " choice

        if [[ "$choice" == "q" ]]; then
            log_info "Abandoning ship..."
            exit 0
        elif [[ "$choice" == "0" ]]; then
            read -p "Enter device path (e.g., /dev/sdb): " manual_device
            if [[ -b "$manual_device" ]]; then
                selected_device="$manual_device"
                break
            else
                log_error "Invalid device: $manual_device"
                continue
            fi
        elif [[ "$choice" =~ ^[0-9]+$ ]] && [ "$choice" -ge 1 ] && [ "$choice" -le ${#devices[@]} ]; then
            selected_device="${devices[$((choice-1))]}"
            break
        else
            log_error "Invalid fucking choice!"
        fi
    done
}

# Function to setup output directory and filename
setup_output() {
    # Create images directory if it doesn't exist
    mkdir -p images

    # Generate filename with current timestamp
    local timestamp=$(date +%Y-%m-%d_%H-%M-%S)
    output_file="images/piraterf_${timestamp}.img"

    log_info "Output file: $output_file"
}

# Function to clone device
clone_device() {
    log_section "Cloning $selected_device to $output_file"

    # Unmount any existing partitions to avoid conflicts
    log_info "Unmounting any mounted partitions from $selected_device..."

    # Check if any partitions are mounted and unmount them
    for partition in ${selected_device}*; do
        if [[ -b "$partition" ]]; then
            # Get mount point if partition is mounted
            local mount_point=$(findmnt -n -o TARGET "$partition" 2>/dev/null || true)
            if [[ -n "$mount_point" ]]; then
                log_info "Unmounting $partition from $mount_point"
                sudo umount "$partition" 2>/dev/null || true
            fi
        fi
    done

    sleep 1

    log_warning "This may take a while depending on SD card size..."
    log_info "Prepare for some fucking copying action!"

    if ! sudo dd if="$selected_device" of="$output_file" bs=512k status=progress conv=sync; then
        log_error "DD failed! Check permissions and device access."
        exit 1
    fi

    log_success "Clone completed successfully!"
}

# Function to shrink image
shrink_image() {
    log_section "Shrinking Image"
    log_info "Time to compress this fucking beast!"

    # Check if pishrink is available
    if ! command -v pishrink.sh >/dev/null 2>&1; then
        log_warning "pishrink.sh not found in PATH"
        read -p "Enter path to pishrink.sh [./pishrink.sh]: " pishrink_path
        pishrink_path=${pishrink_path:-./pishrink.sh}

        if [[ ! -f "$pishrink_path" ]]; then
            log_error "pishrink.sh not found at $pishrink_path"
            log_info "Download it from: https://github.com/Drewsif/PiShrink"
            exit 1
        fi
    else
        pishrink_path="pishrink.sh"
    fi

    log_info "Running pishrink on $output_file..."
    if ! sudo "$pishrink_path" "$output_file"; then
        log_error "Pishrink failed!"
        exit 1
    fi

    log_success "Image shrunk successfully!"
}

# Main execution
if [[ $EUID -ne 0 ]]; then
    log_warning "This script will need sudo privileges for dd and pishrink operations."
fi

detect_fucking_rpi_devices

setup_output

log_section "Final Confirmation"
log_warning "About to clone $selected_device to $output_file - make sure this is the correct fucking device!"
read -p "Continue? (y/N): " confirm

if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
    log_info "Operation aborted. Your device is safe... for now."
    exit 0
fi

clone_device
shrink_image

log_section "Mission Accomplished"
log_success "Your shrunk RF image is ready: $output_file"
log_info "You can now flash it with RPi Imager or dd to spread the chaos!"
