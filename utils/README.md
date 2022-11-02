# Simple interactive videos wrapper script

This script uploads a local .mp4 file to s3. It then runs `./create_titleid` and uses the generated ContentId to run `./transcodelauncher`

## Prerequisites
- `./create_titleid.mac_x86` and `./transcodelauncher.mac_x86` must be executable and in the same directory as this script.
- You must be in a root shell. Run `sudo bash`
- AWS credentials need to be exported and you need authorization to access to resources.

## To use
- Edit these lines so that the script points to the correct resources.
```AWS_S3_BUCKET=""
MAPPING_TABLE=""
ENCODING_TABLE=""
CDN_BUCKET=""
PIPELINE_ID=""
URI_BASE=""
```

- Check if the mp4 was shot in landscape or portrait. Use `-l` for landscape or `-p` for portrait.
- Run `./siv-make.bash <-l or -p> your-file.mp4 project-name`


