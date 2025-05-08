# Stage 1: Create the file
FROM alpine:latest as file-creator

WORKDIR /app

# Create a file in the first stage
RUN echo "This is a file in the first stage." > myfile.txt

# Stage 2: Create a symlink to the file created in the first stage
RUN ln -s /app/myfile.txt symlink_to_file

# Keep the container running for demonstration purposes
CMD ["sh", "-c", "echo 'File and symlink created:' && ls -la && echo '\nContent through symlink:' && cat symlink_to_file && echo '\nSleeping...' && sleep infinity"]
