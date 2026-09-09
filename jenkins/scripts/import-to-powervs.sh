#!/usr/bin/env bash
# Script to import OVA image from COS to PowerVS workspace

set -o errexit
set -o nounset

# Configuration
WORK_DIR="${WORK_DIR:-/data/centos-images}"
IBM_API_KEY="${IBM_API_KEY:?IBM_API_KEY environment variable is required}" # pragma: allowlist secret
POWERVS_WORKSPACE_CRN="${POWERVS_WORKSPACE_CRN:?POWERVS_WORKSPACE_CRN environment variable is required}"
COS_BUCKET_NAME="${COS_BUCKET_NAME:?COS_BUCKET_NAME environment variable is required}"
COS_REGION="${COS_REGION:-us-south}"
COS_HMAC_ACCESS_KEY="${COS_HMAC_ACCESS_KEY:?COS_HMAC_ACCESS_KEY environment variable is required}"
COS_HMAC_SECRET_KEY="${COS_HMAC_SECRET_KEY:?COS_HMAC_SECRET_KEY environment variable is required}"

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

# Get image metadata
get_image_metadata() {
    local metadata_file="${WORK_DIR}/image-metadata.json"
    
    if [ ! -f "$metadata_file" ]; then
        log_error "Metadata file not found: $metadata_file"
        exit 1
    fi
    
    cat "$metadata_file"
}

# Import image to PowerVS
import_image() {
    local image_name="$1"
    local cos_object_name="$2"

    # Extract instance GUID from CRN for --pvs-instance-id
    # CRN format: crn:v1:bluemix:public:power-iaas:region:a/account:GUID::
    local pvs_instance_id
    pvs_instance_id=$(echo "$POWERVS_WORKSPACE_CRN" | awk -F: '{print $8}')

    export IBMCLOUD_API_KEY="$IBM_API_KEY"

    log_info "Importing image to PowerVS workspace"
    log_info "Workspace CRN : $POWERVS_WORKSPACE_CRN"
    log_info "Instance GUID : $pvs_instance_id"
    log_info "Image name    : $image_name"
    log_info "COS object    : $cos_object_name"
    log_info "COS bucket    : $COS_BUCKET_NAME"
    log_info "COS region    : $COS_REGION"

    log_info "Running: pvsadm image import --pvs-instance-id $pvs_instance_id --bucket $COS_BUCKET_NAME --bucket-region $COS_REGION --object $cos_object_name --pvs-image-name $image_name --accesskey <masked> --secretkey <masked> --watch"

    local import_raw="${WORK_DIR}/import_attempt.log"

    pvsadm image import \
        --pvs-instance-id "$pvs_instance_id" \
        --bucket "$COS_BUCKET_NAME" \
        --bucket-region "$COS_REGION" \
        --object "$cos_object_name" \
        --pvs-image-name "$image_name" \
        --accesskey "$COS_HMAC_ACCESS_KEY" \
        --secretkey "$COS_HMAC_SECRET_KEY" \
        --watch \
        2>&1 | tee "$import_raw" | tee "${WORK_DIR}/import.log"

    local import_status=${PIPESTATUS[0]}

    if [ $import_status -ne 0 ]; then
        # PowerVS returns 400 when the image name already exists or was recently
        # deleted and is in a cooling-off period. Both mean the image is already
        # present — treat as success, not a pipeline failure.
        if grep -qiE "already exists|recently deleted" "$import_raw"; then
            log_warn "=========================================="
            log_warn "Image '$image_name' already exists in PowerVS — skipping import"
            log_warn "To re-import: delete it from the PowerVS console, wait a few minutes,"
            log_warn "then re-run the pipeline."
            log_warn "=========================================="
            echo "SKIPPED: image already exists: $image_name" >> "${WORK_DIR}/import.log"
            return 0
        fi

        log_error "Import failed with status: $import_status"
        cat "$import_raw" >&2
        exit 1
    fi

    log_info "Import completed successfully"
}

# Get imported image ID from the import.log (pvsadm prints it on success)
get_image_id() {
    local image_name="$1"

    log_info "Retrieving image ID from import log..."

    # pvsadm prints: "Successfully imported the image: <name> with ID: <id> within ..."
    local image_id=$(grep -oP 'with ID: \K[a-f0-9-]+' "${WORK_DIR}/import.log" 2>/dev/null | head -1)

    if [ -z "$image_id" ]; then
        log_warn "Could not parse image ID from import log"
        return 1
    fi

    log_info "Image ID: $image_id"
    echo "$image_id"
}

# Update metadata with PowerVS information
update_metadata() {
    local image_id="$1"
    local metadata_file="${WORK_DIR}/image-metadata.json"
    local temp_file="${metadata_file}.tmp"
    
    jq --arg workspace "$POWERVS_WORKSPACE_CRN" \
       --arg image_id "$image_id" \
       '. + {
           powervs_workspace_crn: $workspace,
           powervs_image_id: $image_id,
           import_date: (now | strftime("%Y-%m-%dT%H:%M:%SZ"))
       }' "$metadata_file" > "$temp_file"
    
    mv "$temp_file" "$metadata_file"
    
    log_info "Metadata updated with PowerVS information"
}

# Main execution
main() {
    log_info "Starting image import to PowerVS workspace"
    log_info "Work directory: $WORK_DIR"
    
    # Check requirements
    check_requirements
    
    # Get metadata
    local metadata=$(get_image_metadata)
    local image_name=$(echo "$metadata" | jq -r '.image_name')
    local cos_object_name=$(echo "$metadata" | jq -r '.cos_object_name')
    
    if [ -z "$image_name" ] || [ "$image_name" == "null" ]; then
        log_error "Image name not found in metadata"
        exit 1
    fi
    
    if [ -z "$cos_object_name" ] || [ "$cos_object_name" == "null" ]; then
        log_error "COS object name not found in metadata"
        exit 1
    fi
    
    log_info "Image name: $image_name"
    log_info "COS object: $cos_object_name"
    
    # Import image
    import_image "$image_name" "$cos_object_name"
    
    # Get image ID
    local image_id=""
    if image_id=$(get_image_id "$image_name"); then
        # Update metadata
        update_metadata "$image_id"
    else
        log_warn "Could not retrieve image ID, but import may have succeeded"
    fi
    
    log_info "=========================================="
    log_info "Import completed successfully!"
    log_info "Workspace: $POWERVS_WORKSPACE_CRN"
    log_info "Image name: $image_name"
    if [ -n "$image_id" ]; then
        log_info "Image ID: $image_id"
    fi
    log_info "=========================================="
    
    # Display next steps
    log_info ""
    log_info "Next steps:"
    log_info "1. Verify the image in PowerVS console"
    log_info "2. Test deploy a VM with the new image"
    log_info "3. Update Catalog CRD with new image ID"
}

# Run main function
main "$@"


