#!/bin/bash

declare -a ENTRY_POINTS
declare -a SOURCES 
declare -a REGIONS 
declare -a RUNTIMES
declare -a ENV_VARIABLES

echo "Deploying The Google Cloud Function/(s)"

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
    
    read -p "Enter Go runtime  (Press enter for default) (default: go124): " RUNTIME
    RUNTIME=${RUNTIME:-go124} 

    read -p "Enter the env variables with key value pairs, comma separated (Press enter to skip): " ENV_VARIABLE

    ENTRY_POINTS+=("$ENTRY_POINT")
    SOURCES+=("$SOURCE")
    REGIONS+=("$REGION")
    RUNTIMES+=("$RUNTIME")
    ENV_VARIABLES+=("$ENV_VARIABLE")

    read -p "Do you want to continue entering functions (Y/N): " ANSWER
    if [[ "$ANSWER" == "Y" || "$ANSWER" == "y" ]]; then
        continue
    else 
        break
    fi
done

for i in "${!ENTRY_POINTS[@]}"; do
    echo "Deploying '${ENTRY_POINTS[$i]}' to region '${REGIONS[$i]}'..."
    go env -w GOSUMDB='sum.golang.org'
    
    CMD="gcloud functions deploy \"${ENTRY_POINTS[$i]}\" \
      --gen2 \
      --runtime=\"${RUNTIMES[$i]}\" \
      --entry-point=\"${ENTRY_POINTS[$i]}\" \
      --source=\"${SOURCES[$i]}\" \
      --region=\"${REGIONS[$i]}\" \
      --trigger-http \
      --allow-unauthenticated" 

    #Check if the env variables are not empty
    if [ -n "${ENV_VARIABLES[$i]}" ]; then
        CMD="$CMD --set-env-vars=\"${ENV_VARIABLES[$i]}\""
    fi

    echo "Running the gc deploy command:"
    eval $CMD

    echo "Deployment complete. "${ENTRY_POINTS[$i]}" has been deployed"
done
