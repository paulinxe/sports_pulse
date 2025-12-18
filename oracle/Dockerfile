FROM ghcr.io/foundry-rs/foundry:latest

# Switch to root to create the user
USER root

# Create a user with UID 1002 and GID 1002 to match host permissions
RUN groupadd -g 1002 foundry-user && \
    useradd -u 1002 -g 1002 -m -s /bin/bash foundry-user

# Switch to the new user
USER foundry-user

WORKDIR /app

