# Set the base image
FROM ubuntu:24.04

# Copy the profile archive
COPY profile.tar.gz /

# Extract the profile and clean up the archive
RUN tar --skip-old-files -xvf /profile.tar.gz -C / && rm /profile.tar.gz

# Overwrite user and group data
COPY profile/etc/passwd profile/etc/group /etc/

# Set environment variables
ENV LANG=C.UTF-8
ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/snap/bin
ENV NOTIFY_SOCKET=/run/systemd/notify
ENV USER=mysql
ENV LOGNAME=mysql
ENV HOME=/nonexistent
ENV INVOCATION_ID=57deaa9208a44849bb19ca6024cd5af9
ENV JOURNAL_STREAM=7:6252083
ENV RUNTIME_DIRECTORY=/run/mysqld
ENV SYSTEMD_EXEC_PID=496003
ENV MYSQLD_PARENT_PID=1

# Set the user and group
USER mysql:mysql

# Set the working directory
WORKDIR /var/lib/mysql

# Expose ports
EXPOSE 3306/tcp
EXPOSE 33060/tcp

# Set the entry point
CMD ["/usr/sbin/mysqld"]
