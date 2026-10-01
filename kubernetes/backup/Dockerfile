# Backup image: rclone with the sqlite3 CLI preinstalled, so the backup job
# does not download packages on every run.
#
# Build and push (the tag follows the rclone version):
#   docker build -t ypeskov/kkal-tracker-backup:1.75.1 kubernetes/backup
#   docker push ypeskov/kkal-tracker-backup:1.75.1
FROM rclone/rclone:1.75.1@sha256:45401ad7410db1d67ffdb58e19059ad20b0d8e0285a60e38bbec55cc1019c7a5
RUN apk add --no-cache sqlite
