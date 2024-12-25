# Set the base image
FROM ubuntu:24.04

# Copy the profile archive
COPY profile.tar.gz /

# Extract the profile and clean up the archive
RUN tar --skip-old-files -xvf /profile.tar.gz -C / && rm /profile.tar.gz

# Overwrite user and group data
COPY profile/etc/passwd profile/etc/group /etc/

# Set environment variables

# Set the user and group
USER root:root

# Set the working directory
WORKDIR /

# Expose ports
EXPOSE 80/tcp

# Set the entry point
CMD ["/usr/sbin/nginx", "-g", "daemon off; master_process on;"]
