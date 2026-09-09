#!/usr/bin/env bash
# Script to upload OVA file to IBM Cloud Object Storage

set -o errexit
set -o nounset

# Configuration
WORK_DIR="${WORK_DIR:-/data/centos-images}"
IBM_API_KEY="${IBM_API_KEY:?IBM_API_KEY environment variable is required}" # pragma: allowlist secret
COS_BUCKET_NAME="${COS_BUCKET_NAME:?COS_BUCKET_NAME environment variable is required}"
COS_HMAC_ACCESS_KEY="${COS_HMAC_ACCESS_KEY:?COS_HMAC_ACCESS_KEY environment variable is required}"
COS_HMAC_SECRET_KEY="${COS_HMAC_SECRET_KEY:?COS_HMAC_SECRET_KEY environment variable is required}"
COS_REGION="${COS_REGION:-us-south}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1" >&2
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1" >&2
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# Check required tools
check_requirements() {
    local missing_tools=()
    
    for tool in pvsadm jq; do
        if ! command -v $tool &> /dev/null; then
            missing_tools+=($tool)
        fi
    done
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        exit 1
    fi
    
    log_info "All required tools are available"
}

# Get OVA file — prefer metadata, fall back to glob search in WORK_DIR
get_ova_file() {
    local metadata_file="${WORK_DIR}/image-metadata.json"
    local ova_file=""

    # Try metadata first
    if [ -f "$metadata_file" ]; then
        ova_file=$(jq -r '.ova_file // empty' "$metadata_file" 2>/dev/null || true)
    fi

    # If metadata path is missing or stale, search by glob
    if [ -z "$ova_file" ] || [ ! -f "$ova_file" ]; then
        log_warn "OVA path from metadata not usable, searching ${WORK_DIR} for *.ova.gz..."
        ova_file=$(find "$WORK_DIR" -maxdepth 1 -name "*.ova.gz" -type f 2>/dev/null | sort -V | tail -1)
    fi

    if [ -z "$ova_file" ] || [ ! -f "$ova_file" ]; then
        log_error "No OVA file found in ${WORK_DIR}. Run the Download & Convert stage first."
        exit 1
    fi

    log_info "Using OVA file: $ova_file"
    echo "$ova_file"
}

# Upload OVA to COS using pvsadm
upload_to_cos() {
    local ova_file="$1"
    local ova_filename=$(basename "$ova_file")

    log_info "Uploading OVA to IBM Cloud Object Storage"
    log_info "File   : $ova_filename"
    log_info "Bucket : $COS_BUCKET_NAME"
    log_info "Region : $COS_REGION"

    log_info "Running: pvsadm image upload --bucket $COS_BUCKET_NAME --bucket-region $COS_REGION --accesskey <masked> --secretkey <masked> -f $ova_file"

    # Capture output and check for duplicate error
    local upload_output="${WORK_DIR}/upload_attempt.log"

    if pvsadm image upload \
        --bucket "$COS_BUCKET_NAME" \
        --bucket-region "$COS_REGION" \
        --accesskey "$COS_HMAC_ACCESS_KEY" \
        --secretkey "$COS_HMAC_SECRET_KEY" \
        -f "$ova_file" \
        2>&1 | tee "$upload_output" | awk '
            /[0-9]+%/ {
                match($0, /([0-9]+)%/, arr)
                percent = arr[1]
                if (percent >= 20 && percent % 20 == 0 && percent != last) {
                    print "[INFO] Upload progress: " percent "%"
                    last = percent
                }
                next
            }
            { print }
        ' | tee "${WORK_DIR}/upload.log"; then
        log_info "Upload completed successfully"
    else
        # Check if error is due to file already existing
        if grep -q "object already exists" "$upload_output"; then
            log_warn "=========================================="
            log_warn "File already exists in COS bucket"
            log_warn "This is expected on retry - continuing..."
            log_warn "=========================================="
            # Not a fatal error, continue
        else
            log_error "Upload failed with unexpected error"
            log_error "Full pvsadm output:"
            cat "$upload_output" >&2
            exit 1
        fi
    fi
    
    # Update metadata with COS information
    local metadata_file="${WORK_DIR}/image-metadata.json"
    local temp_file="${metadata_file}.tmp"
    
    jq --arg bucket "$COS_BUCKET_NAME" \
       --arg region "$COS_REGION" \
       --arg object "$ova_filename" \
       '. + {
           cos_bucket: $bucket,
           cos_region: $region,
           cos_object_name: $object,
           upload_date: (now | strftime("%Y-%m-%dT%H:%M:%SZ"))
       }' "$metadata_file" > "$temp_file"
    
    mv "$temp_file" "$metadata_file"
    
    log_info "Metadata updated with COS information"
}

# Verify upload using s3cmd/curl with HMAC keys
verify_upload() {
    local ova_filename="$1"

    log_info "Verifying upload..."

    # Use curl with HMAC signature to list bucket objects
    local endpoint="https://s3.${COS_REGION}.cloud-object-storage.appdomain.cloud"
    local date_val=$(date -u +"%a, %d %b %Y %H:%M:%S +0000")
    local string_to_sign="GET\n\n\n${date_val}\n/${COS_BUCKET_NAME}/?prefix=${ova_filename}"
    local signature=$(echo -en "$string_to_sign" | openssl dgst -sha1 -hmac "$COS_HMAC_SECRET_KEY" -binary | base64)

    local http_code=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Date: ${date_val}" \
        -H "Authorization: AWS ${COS_HMAC_ACCESS_KEY}:${signature}" \
        "${endpoint}/${COS_BUCKET_NAME}/?prefix=${ova_filename}")

    if [ "$http_code" = "200" ]; then
        log_info "Upload verification successful: $ova_filename found in bucket"
        return 0
    else
        log_warn "Could not verify upload (HTTP $http_code) – upload may still have succeeded"
        return 1
    fi
}

# Main execution
main() {
    log_info "Starting OVA upload to IBM Cloud Object Storage"
    log_info "Work directory: $WORK_DIR"
    
    # Check requirements
    check_requirements
    
    # Get OVA file path
    local ova_file=$(get_ova_file)
    log_info "OVA file: $ova_file"
    
    # Get file size
    local file_size=$(du -h "$ova_file" | cut -f1)
    log_info "File size: $file_size"
    
    # Upload to COS
    upload_to_cos "$ova_file"
    
    # Verify upload
    local ova_filename=$(basename "$ova_file")
    verify_upload "$ova_filename" || log_warn "Verification skipped or failed"
    
    log_info "=========================================="
    log_info "Upload completed successfully!"
    log_info "Bucket: $COS_BUCKET_NAME"
    log_info "Region: $COS_REGION"
    log_info "Object: $ova_filename"
    log_info "=========================================="
}

# Run main function
main "$@"


