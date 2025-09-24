#!/usr/bin/env bash

CLOUD_INIT_DEV="/dev/sr0"
MNT_POINT="/mnt"
HOME_DIR="/root"
AUTH_KEYS_FILE="${HOME_DIR}/.ssh/authorized_keys"
USER_DATA="${MNT_POINT}/openstack/latest/user_data"

if [[ ! -b "$CLOUD_INIT_DEV" ]]; then
    echo "INFO: $CLOUD_INIT_DEV doesn't exist, exiting gracefully."
    exit 0
fi

if mount | grep ${CLOUD_INIT_DEV}; then
    echo "INFO: already found the mounted"
else
    echo "Mounting the ${CLOUD_INIT_DEV}...."
    if mount ${CLOUD_INIT_DEV} ${MNT_POINT}; then
       echo "INFO: successfully mounted ${CLOUD_INIT_DEV} to ${MNT_POINT}"
    else
       echo "ERROR: Failed to mount ${CLOUD_INIT_DEV} to ${MNT_POINT}, hence exiting with error."
        exit 1
    fi
fi

if [[ -f "$USER_DATA" ]]; then
    echo "INFO: $USER_DATA exist."
else
    echo "ERROR: $USER_DATA doesn't exist, hence exiting with error."
    exit 1
fi

mkdir -m 700 -p ${HOME_DIR}/.ssh

if [[ -f "$AUTH_KEYS_FILE" ]]; then
    echo "INFO: $AUTH_KEYS_FILE exists."
else
    install -m 600 /dev/null ${AUTH_KEYS_FILE}
    echo "Created a ${AUTH_KEYS_FILE}"
fi

if grep "$(cat $USER_DATA)" "$AUTH_KEYS_FILE"; then
    echo "INFO: $AUTH_KEYS_FILE already contains the $USER_DATA key, exiting gracefully"
    exit 0
fi

if cat $USER_DATA >> "$AUTH_KEYS_FILE"; then
    echo "INFO: Added the public ssh key to the $AUTH_KEYS_FILE"
else
    echo "ERROR: Fails to add $USER_DATA content to the $AUTH_KEYS_FILE, hence exiting with error"
    exit 1
fi

# -------------------------------
# Upgrade-warning script in /etc/profile.d
# -------------------------------
UPGRADE_WARN_SCRIPT="/etc/profile.d/upgrade-warning.sh"

cat <<'EOF' > "$UPGRADE_WARN_SCRIPT"
#!/usr/bin/env bash
# /etc/profile.d/upgrade-warning.sh
# Warns about available package and CentOS release updates at login

# Only run for interactive shells
[[ $- != *i* ]] && return

# Colors
RESET="\e[0m"

updates_found=0

# -------------------------------
# Check package updates
# -------------------------------
if command -v dnf >/dev/null 2>&1; then
    OUTPUT=$(dnf -q --refresh check-update 2>&1)
    RET=$?

    if [ $RET -eq 100 ]; then
        echo -e "\n${RESET}⚠️  Updates are available via dnf (security/patches).${RESET}"
        echo -e "   Run 'dnf update' and then 'reboot' to apply.\n"
        updates_found=1
    elif [ $RET -ne 0 ]; then
        echo -e "\n${RESET}❌ Error checking updates!${RESET}"
        echo "$OUTPUT" | grep -E 'repo|failure|cannot'
        echo -e "Please report this to PowerACL@ibm.com for the resolution"
        updates_found=1
    fi
fi

# -------------------------------
# If nothing found
# -------------------------------
if [ $updates_found -eq 0 ]; then
    echo -e "\n${RESET}✅ System is fully up to date.${RESET}\n"
fi
EOF

chmod 644 "$UPGRADE_WARN_SCRIPT"
echo -e "INFO: Installed patch-warning script into $UPGRADE_WARN_SCRIPT\n"

