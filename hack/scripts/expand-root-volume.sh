#!/bin/bash
set -x

# 1. Rescan physical paths
for dev in /sys/block/sd*/device/rescan; do echo 1 > "$dev"; done
sleep 2

# 2. Resize Multipath Map
multipathd -k"resize map mpatha"
sleep 2

# 3. Handle GPT 'Fix' separately and silently
# We pipe 'Fix' to a print command. If there's no error, Fix is ignored.
printf "Fix\n" | parted /dev/mapper/mpatha ---pretend-input-tty print > /dev/null 2>&1

# 4. Resize the partition
printf "Yes\n" | parted /dev/mapper/mpatha ---pretend-input-tty resizepart 2 100%

# 5. Refresh mappings
kpartx -u /dev/mapper/mpatha
udevadm settle

# 6. Grow Filesystem
xfs_growfs /

# 7. Flag completion
touch /var/lib/pac-root-expanded
