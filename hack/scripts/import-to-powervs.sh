#!/usr/bin/env bash
# Script to import OVA image from COS to PowerVS workspace

set -o errexit
set -o nounset
set -o pipefail

# Configuration
WORK_DIR="${WORK_DIR:-/tmp/centos-images}"
IBM_API_KEY="${IBM_API_KEY:?IBM_API_KEY environment variable is required}" # pragma: allowlist secret
POWERVS_INSTANCE_ID="${POWERVS_INSTANCE_ID:-}"
POWERVS_INSTANCE_NAME="${POWERVS_INSTANCE_NAME:-}"
POWERVS_WORKSPACE_CRN="${POWERVS_WORKSPACE_CRN:-}"
COS_BUCKET_NAME="${COS_BUCKET_NAME:?COS_BUCKET_NAME environment variable is required}"
COS_REGION="${COS_REGION:-us-south}"
COS_HMAC_ACCESS_KEY="${COS_HMAC_ACCESS_KEY:-}"
COS_HMAC_SECRET_KEY="${COS_HMAC_SECRET_KEY:-}"
IMPORT_TIMEOUT="${IMPORT_TIMEOUT:-3600}"  # 1 hour timeout

# Check that at least one PowerVS identifier is provided
if [ -z "$POWERVS_INSTANCE_ID" ] && [ -z "$POWERVS_INSTANCE_NAME" ] && [ -z "$POWERVS_WORKSPACE_CRN" ]; then
    log_error "One of POWERVS_INSTANCE_ID, POWERVS_INSTANCE_NAME, or POWERVS_WORKSPACE_CRN is required"
    exit 1
fi

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

# Get PowerVS instance identifier (ID or name)
get_pvs_instance() {
    # Priority: POWERVS_INSTANCE_ID > POWERVS_INSTANCE_NAME > extract from CRN
    if [ -n "$POWERVS_INSTANCE_ID" ]; then
        echo "$POWERVS_INSTANCE_ID"
        return 0
    elif [ -n "$POWERVS_INSTANCE_NAME" ]; then
        echo "$POWERVS_INSTANCE_NAME"
        return 0
    elif [ -n "$POWERVS_WORKSPACE_CRN" ]; then
        # CRN format: crn:v1:bluemix:public:power-iaas:region:a/account:workspace-id::
        # Extract the workspace-id part (second to last segment)
        local instance_id=$(echo "$POWERVS_WORKSPACE_CRN" | awk -F: '{print $(NF-1)}')
        if [ -n "$instance_id" ]; then
            echo "$instance_id"
            return 0
        fi
    fi
    
    log_error "Could not determine PowerVS instance ID or name"
    return 1
}

# Check if image already exists in PowerVS
check_image_exists() {
    local image_name="$1"
    local pvs_instance="$2"
    local get_flag="$3"
    
    log_info "Checking if image already exists in PowerVS..."
    
    export IBMCLOUD_API_KEY="$IBM_API_KEY"
    
    # List images and check if our image exists
    local existing_image=$(pvsadm get images $get_flag "$pvs_instance" --json 2>/dev/null | \
        jq -r --arg name "$image_name" '.[] | select(.name == $name) | .name' | head -1)
    
    if [ -n "$existing_image" ]; then
        log_info "Image already exists in PowerVS: $existing_image"
        return 0
    else
        log_info "Image not found in PowerVS, will proceed with import"
        return 1
    fi
}

# Import image to PowerVS
import_image() {
    local image_name="$1"
    local cos_object_name="$2"
    
    log_info "Importing image to PowerVS workspace"
    log_info "Image name: $image_name"
    log_info "COS object: $cos_object_name"
    
    # Set IBM Cloud API key
    export IBMCLOUD_API_KEY="$IBM_API_KEY"
    
    # Get PowerVS instance identifier
    local pvs_instance=$(get_pvs_instance)
    if [ $? -ne 0 ]; then
        log_error "Failed to get PowerVS instance identifier"
        exit 1
    fi
    
    # Determine if it's an ID or name
    local import_flag=""
    local get_flag=""
    if [ -n "$POWERVS_INSTANCE_ID" ] || [ -n "$POWERVS_WORKSPACE_CRN" ]; then
        import_flag="--pvs-instance-id"
        get_flag="--pvs-instance-id"
        log_info "Using PowerVS Instance ID: $pvs_instance"
    else
        import_flag="--pvs-instance-name"
        get_flag="--pvs-instance-name"
        log_info "Using PowerVS Instance Name: $pvs_instance"
    fi
    
    # Check if image already exists
    if check_image_exists "$image_name" "$pvs_instance" "$get_flag"; then
        log_info "=========================================="
        log_info "Image already exists in PowerVS workspace"
        log_info "Skipping import to avoid duplicate"
        log_info "To force re-import, delete the existing image first"
        log_info "=========================================="
        return 0
    fi
    
    # Import using pvsadm
    log_info "Starting import with pvsadm..."
    
    local import_output="${WORK_DIR}/import-output.json"
    
    # Build import command
    local import_cmd="pvsadm image import $import_flag \"$pvs_instance\" --bucket \"$COS_BUCKET_NAME\" --bucket-region \"$COS_REGION\" --object \"$cos_object_name\" --pvs-image-name \"$image_name\""
    
    # Add HMAC credentials if provided
    if [ -n "$COS_HMAC_ACCESS_KEY" ] && [ -n "$COS_HMAC_SECRET_KEY" ]; then
        import_cmd="$import_cmd --accesskey \"$COS_HMAC_ACCESS_KEY\" --secretkey \"$COS_HMAC_SECRET_KEY\""
        log_info "Using HMAC credentials for COS access"
    else
        log_info "No HMAC credentials provided, pvsadm will auto-generate service credentials"
    fi
    
    import_cmd="$import_cmd --watch"
    
    log_info "Running import command..."
    eval "$import_cmd" 2>&1 | tee "${WORK_DIR}/import.log"
    
    local import_status=$?
    
    if [ $import_status -eq 0 ]; then
        log_info "Import completed successfully"
    else
        log_error "Import failed with status: $import_status"
        exit 1
    fi
}

# Get imported image ID
get_image_id() {
    local image_name="$1"
    
    log_info "Retrieving image ID from PowerVS workspace..."
    
    export IBMCLOUD_API_KEY="$IBM_API_KEY"
    
    # Extract PowerVS instance ID from CRN
    local pvs_instance_id=$(extract_pvs_instance "$POWERVS_WORKSPACE_CRN")
    
    # List images and find the one we just imported
    local image_id=$(pvsadm get images --pvs-instance-id "$pvs_instance_id" --json 2>/dev/null | \
        jq -r --arg name "$image_name" '.[] | select(.name == $name) | .imageID' | head -1)
    
    if [ -z "$image_id" ] || [ "$image_id" == "null" ]; then
        log_warn "Could not retrieve image ID automatically"
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


