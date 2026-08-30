#!/usr/bin/env bash
# Script to cleanup temporary files and resources

set -o errexit
set -o nounset
set -o pipefail

# Configuration
WORK_DIR="${WORK_DIR:-/tmp/centos-images}"
KEEP_METADATA="${KEEP_METADATA:-true}"
KEEP_LOGS="${KEEP_LOGS:-true}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Display disk usage before cleanup
show_disk_usage_before() {
    log_info "Disk usage before cleanup:"
    if [ -d "$WORK_DIR" ]; then
        du -sh "$WORK_DIR" 2>/dev/null || log_warn "Could not calculate disk usage"
        log_info "Files in work directory:"
        ls -lh "$WORK_DIR" 2>/dev/null || log_warn "Could not list files"
    else
        log_info "Work directory does not exist: $WORK_DIR"
    fi
}

# Cleanup qcow2 files
cleanup_qcow2_files() {
    log_info "Cleaning up qcow2 files..."
    
    local count=0
    if [ -d "$WORK_DIR" ]; then
        for file in "$WORK_DIR"/*.qcow2; do
            if [ -f "$file" ]; then
                log_info "Removing: $(basename "$file")"
                rm -f "$file"
                ((count++))
            fi
        done
        
        # Also remove checksum files
        for file in "$WORK_DIR"/*.SHA256SUM; do
            if [ -f "$file" ]; then
                log_info "Removing: $(basename "$file")"
                rm -f "$file"
                ((count++))
            fi
        done
    fi
    
    log_info "Removed $count qcow2 and checksum files"
}

# Cleanup OVA files
cleanup_ova_files() {
    log_info "Cleaning up OVA files..."
    
    local count=0
    if [ -d "$WORK_DIR" ]; then
        for file in "$WORK_DIR"/*.ova.gz "$WORK_DIR"/*.ova; do
            if [ -f "$file" ]; then
                log_info "Removing: $(basename "$file")"
                rm -f "$file"
                ((count++))
            fi
        done
    fi
    
    log_info "Removed $count OVA files"
}

# Cleanup temporary files
cleanup_temp_files() {
    log_info "Cleaning up temporary files..."
    
    local count=0
    if [ -d "$WORK_DIR" ]; then
        # Remove temp directories created by pvsadm
        for dir in "$WORK_DIR"/tmp* "$WORK_DIR"/pvsadm-*; do
            if [ -d "$dir" ]; then
                log_info "Removing directory: $(basename "$dir")"
                rm -rf "$dir"
                ((count++))
            fi
        done
        
        # Remove other temporary files
        for file in "$WORK_DIR"/*.tmp "$WORK_DIR"/*.temp; do
            if [ -f "$file" ]; then
                log_info "Removing: $(basename "$file")"
                rm -f "$file"
                ((count++))
            fi
        done
    fi
    
    log_info "Removed $count temporary files/directories"
}

# Cleanup logs (optional)
cleanup_logs() {
    if [ "$KEEP_LOGS" == "false" ]; then
        log_info "Cleaning up log files..."
        
        local count=0
        if [ -d "$WORK_DIR" ]; then
            for file in "$WORK_DIR"/*.log; do
                if [ -f "$file" ]; then
                    log_info "Removing: $(basename "$file")"
                    rm -f "$file"
                    ((count++))
                fi
            done
        fi
        
        log_info "Removed $count log files"
    else
        log_info "Keeping log files (KEEP_LOGS=true)"
    fi
}

# Cleanup metadata (optional)
cleanup_metadata() {
    if [ "$KEEP_METADATA" == "false" ]; then
        log_info "Cleaning up metadata files..."
        
        local count=0
        if [ -d "$WORK_DIR" ]; then
            for file in "$WORK_DIR"/*.json; do
                if [ -f "$file" ]; then
                    log_info "Removing: $(basename "$file")"
                    rm -f "$file"
                    ((count++))
                fi
            done
        fi
        
        log_info "Removed $count metadata files"
    else
        log_info "Keeping metadata files (KEEP_METADATA=true)"
    fi
}

# Archive important files before cleanup
archive_important_files() {
    log_info "Archiving important files..."
    
    if [ -d "$WORK_DIR" ]; then
        local archive_dir="${WORK_DIR}/archive"
        mkdir -p "$archive_dir"
        
        # Archive metadata
        for file in "$WORK_DIR"/*.json; do
            if [ -f "$file" ]; then
                cp "$file" "$archive_dir/" 2>/dev/null || true
            fi
        done
        
        # Archive logs
        for file in "$WORK_DIR"/*.log; do
            if [ -f "$file" ]; then
                cp "$file" "$archive_dir/" 2>/dev/null || true
            fi
        done
        
        # Create tarball
        if [ -d "$archive_dir" ] && [ "$(ls -A "$archive_dir")" ]; then
            local timestamp=$(date +%Y%m%d-%H%M%S)
            local archive_file="${WORK_DIR}/centos-image-automation-${timestamp}.tar.gz"
            
            tar -czf "$archive_file" -C "$WORK_DIR" archive/
            log_info "Created archive: $archive_file"
            
            # Remove archive directory
            rm -rf "$archive_dir"
        fi
    fi
}

# Display disk usage after cleanup
show_disk_usage_after() {
    log_info "Disk usage after cleanup:"
    if [ -d "$WORK_DIR" ]; then
        du -sh "$WORK_DIR" 2>/dev/null || log_warn "Could not calculate disk usage"
        
        if [ "$(ls -A "$WORK_DIR" 2>/dev/null)" ]; then
            log_info "Remaining files:"
            ls -lh "$WORK_DIR" 2>/dev/null || true
        else
            log_info "Work directory is empty"
        fi
    else
        log_info "Work directory does not exist: $WORK_DIR"
    fi
}

# Remove work directory completely (optional)
remove_work_dir() {
    if [ "${REMOVE_WORK_DIR:-false}" == "true" ]; then
        log_warn "Removing entire work directory: $WORK_DIR"
        rm -rf "$WORK_DIR"
        log_info "Work directory removed"
    fi
}

# Main execution
main() {
    log_info "Starting cleanup process"
    log_info "Work directory: $WORK_DIR"
    log_info "Keep metadata: $KEEP_METADATA"
    log_info "Keep logs: $KEEP_LOGS"
    
    # Show disk usage before
    show_disk_usage_before
    
    echo ""
    
    # Archive important files first
    archive_important_files
    
    # Cleanup in order
    cleanup_qcow2_files
    cleanup_ova_files
    cleanup_temp_files
    cleanup_logs
    cleanup_metadata
    
    echo ""
    
    # Show disk usage after
    show_disk_usage_after
    
    # Optionally remove entire work directory
    remove_work_dir
    
    log_info "=========================================="
    log_info "Cleanup completed successfully!"
    log_info "=========================================="
}

# Run main function
main "$@"


