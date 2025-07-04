#!/bin/bash

declare -a ENTRY_POINTS
declare -a SOURCES 
declare -a REGIONS 
declare -a RUNTIMES

echo "Deploying The Google Cloud Function/(s)"

read -p "Enter the service account for the deploy: " SERVICE_ACCOUNT
    if [ -z "$SERVICE_ACCOUNT" ]; then
        echo "Service Account cannot be empty. Exiting..."
        exit 1
    fi 

while true; do
    read -p "Enter entry point: " ENTRY_POINT
    if [ -z "$ENTRY_POINT" ]; then 
        echo "Entry point cannot be empty. Enter the exact matching entry point in the init function."
        break
    fi
    
    read -p "Enter the function source: " SOURCE
    if [ -z "$SOURCE" ]; then
        echo "Function source cannot be empty"
        break
    fi

    read -p "Enter region (Press enter for default) (default: us-west2): " REGION
    REGION=${REGION:-us-west2}
    
    read -p "Enter Go runtime  (Press enter for default) (default: go123): " RUNTIME
    RUNTIME=${RUNTIME:-go123} 

    ENTRY_POINTS+=("$ENTRY_POINT")
    SOURCES+=("$SOURCE")
    REGIONS+=("$REGION")
    RUNTIMES+=("$RUNTIME")

    read -p "Do you want to continue entering functions (Y/N): " ANSWER
    if [[ "$ANSWER" == "Y" || "$ANSWER" == "y" ]]; then
        continue
    else 
        break
    fi
done

for i in "${!ENTRY_POINTS[@]}"; do
    echo "Deploying '${ENTRY_POINTS[$i]}' to region '${REGIONS[$i]}'..."
    
    gcloud functions deploy "${ENTRY_POINTS[$i]}"\
      --gen2 \
      --service-account="${SERVICE_ACCOUNT}" \
      --runtime="${RUNTIMES[$i]}" \
      --entry-point="${ENTRY_POINTS[$i]}" \
      --source="${SOURCES[$i]}" \
      --region="${REGIONS[$i]}" \
      --trigger-http \
      --allow-unauthenticated

    echo "Deployment complete. "${ENTRY_POINTS[$i]}" has been deployed"
done
