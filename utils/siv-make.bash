#! /usr/bin/env bash
set -euo pipefail

# Check if the mp4 file is in portrait, landscape or square. If landscape use the
# -l flag, if portrait use the -p flag.

# First command-line argument is either -l or -h for the transcoder yaml.
# Second is the file to be uploaded to s3.
# Third is the filebase/cdn-bucket/uribase identifier

export AWS_REGION="eu-west-1"

AWS_S3_BUCKET=""
MAPPING_TABLE=""
ENCODING_TABLE=""
CDN_BUCKET=""
PIPELINE_ID=""
URI_BASE=""
ENDPOINT=""

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'


if (( $# < 3 )); then
    echo "Please specify -l or -h to use the horizontal/vertical transcodeset"
    exit 1
fi

while getopts 'lh' OPTION; do
    case "$OPTION" in
        l)
            TRANSCODESET="horizonal_transcode_set.yaml"
            ;;
        p)
            TRANSCODESET="vertical_transcode_set.yaml"
            ;;
        ?)
            echo "Please specify (l)andscape or (p)ortrait"
            exit 1
            ;;
    esac
done

aws s3 cp $2 s3://${AWS_S3_BUCKET}

CONTENT_ID=$(./create_titleid.mac_x86 \
    -table ${MAPPING_TABLE} \
    -filebase $3 \
    | grep ContentId \
    | awk '{print $NF}' \
    | tr "," " ")


./transcodelauncher.mac_x86 \
    -input $2 \
    -pipeline ${PIPELINE_ID} \
    -transcodeset ${TRANSCODESET} \
    -contentId ${CONTENT_ID} \
    -table ${ENCODING_TABLE} \
    -cdnbucket ${CDN_BUCKET}:/$3 \
    -uribase ${URI_BASE}$3

echo -e "${YELLOW}Job URL:"
echo -e "${NC}####################################################################################"
echo -e "${GREEN}${ENDPOINT}$3"
echo -e "${NC}####################################################################################"