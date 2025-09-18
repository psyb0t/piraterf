#!/bin/bash

# PIrateRF System Branding Setup - Hack the Fucking Planet Edition
# This script applies crude hacker aesthetic to the Pi terminal

set -euo pipefail

# Source configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/pi_config.sh"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}🏴‍☠️ PIrateRF System Branding Setup - Let's Hack This Bastard! 🏴‍☠️${NC}"
echo -e "${YELLOW}Starting branding installation...${NC}"

# Create backup directory
backup_dir="/tmp/piraterf_backups_$(date +%Y%m%d_%H%M%S)"
mkdir -p "$backup_dir"
echo -e "${YELLOW}Created backup directory: $backup_dir${NC}"

# Backup existing files
echo -e "${YELLOW}Backing up existing system files...${NC}"
[ -f /etc/motd ] && cp /etc/motd "$backup_dir/"
[ -f /etc/bash.bashrc ] && cp /etc/bash.bashrc "$backup_dir/"

# Create the fucking MOTD
echo -e "${YELLOW}Setting up the fucking MOTD...${NC}"
cat > /etc/motd << 'MOTD_EOF'

 ____  _____        _       ____  _____
|  _ \|_ _|_ __ __ _| |_ ___|  _ \|  ___|
| |_) || || '__/ _` | __/ _ \ |_) | |_
|  __/ | || | | (_| | ||  __/  _ <|  _|
|_|   |___|_|  \__,_|\__\___|_| \_\_|

🏴‍☠️ WELCOME TO THE FUCKING PIRATE STATION 🏴‍☠️

    ⚡ The digital seas are yours to command, you magnificent bastard! ⚡
    🎯 Your mission: Broadcast chaos, spread signals, hack the planet!
    📡 This system is armed and ready to transmit pure fucking anarchy!

    💀 Remember: We are the pirates of the digital age! 💀
    🔥 No rules, no limits, just pure radio freedom! 🔥

    ⚠️  WARNING: This system is configured for maximum signal mayhem! ⚠️

    📋 LEGAL DISCLAIMER: Being legal is recommended or the RF police will fuck you up!
    🚨 DO NOT interfere with emergency services, aviation, or licensed operators!
    ⚖️  Users are responsible for compliance with local radio regulations!
    🔌 USE A FUCKING LOW PASS FILTER! RPi generates harmonics without one!
    🏴‍☠️ I am not liable for your stupid fucking decisions! 🏴‍☠️

    🚀 Ready to raise hell on the airwaves? Let's fucking do this! 🚀

MOTD_EOF

echo -e "${GREEN}✅ MOTD installed successfully!${NC}"

# Clean up /etc/bash.bashrc completely and rebuild it
echo -e "${YELLOW}Rebuilding /etc/bash.bashrc from scratch...${NC}"

cat > /etc/bash.bashrc << 'BASHRC_EOF'
# System-wide .bashrc file for interactive bash(1) shells.

# If not running interactively, don't do anything
case $- in
    *i*) ;;
      *) return;;
esac

# make less more friendly for non-text input files, see lesspipe(1)
[ -x /usr/bin/lesspipe ] && eval "$(SHELL=/bin/sh lesspipe)"

# PIrateRF Global Terminal Configuration
export PS1="\[\e[1;31m\]🏴‍☠️\[\e[0m\] \[\e[1;32m\]\u\[\e[0m\]@\[\e[1;32m\]\h\[\e[0m\]:\[\e[1;32m\]\w\[\e[0m\] \[\e[1;32m\]⚡\[\e[0m\] "

# PIrateRF Global Color Configuration - NEON GREEN THEME
export GREP_COLOR='1;32'
export GREP_COLORS='mt=1;32:fn=1;32:ln=1;36:se=1;30'
export LS_COLORS='di=1;32:fi=0;37:ln=1;36:ex=1;32:*.tar=1;31:*.zip=1;31:*.gz=1;31'

# System aliases with crude commentary
alias ls="echo '📁 Looking at this shit...'; ls -alph --color=auto"
alias ll="echo '📋 Long listing this crap...'; ls -la --color=auto"
alias la="echo '👀 Showing all the hidden shit...'; ls -A --color=auto"
alias l="echo '⚡ Quick file glance...'; ls -CF --color=auto"
alias ..="echo '⬆️ Getting the fuck out of here...'; cd .."
alias ...="echo '🚀 Going way the fuck up...'; cd ../.."
alias cd="echo '🚶 Moving to some other fucking directory...'; builtin cd"
alias grep="grep --color=auto"
alias fgrep="fgrep --color=auto"
alias egrep="egrep --color=auto"
alias ports="echo '🔌 Checking what fucking ports are open...'; netstat -tuln"
alias procs="echo '⚙️ Seeing what shit is running...'; ps aux | grep -v grep"
alias nets="echo '📡 Scanning for wireless shit...'; iwlist scan 2>/dev/null | grep -E \"(ESSID|Frequency|Signal)\" | head -20"
alias freq="echo '🔥 Checking CPU frequency bullshit...'; cat /proc/cpuinfo | grep -i mhz || echo 'Frequency info unavailable'"
alias who_is_here="echo '👥 Who the fuck is logged in?'; who"
alias sys_info="echo '💻 System info dump...'; uname -a"
alias where_am_i="echo '📍 Where the fuck am I?'; pwd"
alias shutdown_now="echo '💀 Shutting this fucker down NOW!'; sudo shutdown -h now"
alias reboot_now="echo '🔄 Rebooting this bastard NOW!'; sudo reboot"

BASHRC_EOF

echo -e "${GREEN}✅ /etc/bash.bashrc rebuilt successfully!${NC}"

# Test the bash syntax
if bash -n /etc/bash.bashrc 2>/dev/null; then
    echo -e "${GREEN}✅ Bash syntax check passed!${NC}"
else
    echo -e "${RED}💥 Bash syntax error detected! Restoring from backup...${NC}"
    if [[ -f "$backup_dir/bash.bashrc" ]]; then
        cp "$backup_dir/bash.bashrc" /etc/bash.bashrc
        echo -e "${GREEN}✅ Restored from backup${NC}"
    fi
    exit 1
fi

# Skip session signaling - just let user reconnect manually
echo -e "${YELLOW}Configuration files updated successfully...${NC}"

# Set proper permissions
chmod 644 /etc/motd
chmod 644 /etc/bash.bashrc

# Override user's ~/.bashrc completely with global config
USER_HOME=$(getent passwd $PI_USER | cut -d: -f6)
if [[ -n "$USER_HOME" && -d "$USER_HOME" ]]; then
    echo -e "${YELLOW}Overriding user ~/.bashrc with global PIrateRF config...${NC}"
    # Back up user's bashrc
    [ -f "$USER_HOME/.bashrc" ] && cp "$USER_HOME/.bashrc" "$backup_dir/user_bashrc"
    # Replace with just a source to global config
    cat > "$USER_HOME/.bashrc" << 'USER_BASHRC_EOF'
# PIrateRF User Configuration - Sources global config
source /etc/bash.bashrc
USER_BASHRC_EOF
    chown $PI_USER:$PI_USER "$USER_HOME/.bashrc"
    echo -e "${GREEN}✅ User bashrc overridden with global config${NC}"
fi

echo -e "${GREEN}🏴‍☠️ PIrateRF Branding Installation Complete! 🏴‍☠️${NC}"
echo -e "${CYAN}System is now branded with maximum fucking attitude!${NC}"
echo -e "${YELLOW}Backup files saved to: $backup_dir${NC}"
echo -e "${CYAN}⚠️  Reconnect to see the fucking changes!${NC}"
