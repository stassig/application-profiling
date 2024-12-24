# Set the base image
FROM ubuntu:24.04

# Copy the profile archive
COPY profile.tar.gz /

# Extract the profile and clean up the archive
RUN tar --skip-old-files -xvf /profile.tar.gz -C / && rm /profile.tar.gz

# Add user and group files
COPY profile/etc/passwd /etc/passwd
COPY profile/etc/group /etc/group

# Set environment variables
ENV LANG=C.UTF-8
ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/snap/bin
ENV NOTIFY_SOCKET=/run/systemd/notify
ENV USER=mysql
ENV LOGNAME=mysql
ENV HOME=/nonexistent
ENV INVOCATION_ID=62ccca8c58754de495796641b5ef1dbe
ENV JOURNAL_STREAM=7:4477010
ENV RUNTIME_DIRECTORY=/run/mysqld
ENV SYSTEMD_EXEC_PID=817162
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
