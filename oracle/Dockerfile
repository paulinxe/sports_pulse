FROM ghcr.io/foundry-rs/foundry:latest

# Switch to root to create the user
USER root

# Create user matching host IDs
ARG UID=1000 # Default value if not provided
ARG GID=1000 # Default value if not provided
RUN groupadd -g ${GID} foundry-user && \
    useradd -u ${UID} -g ${GID} -m -s /bin/bash appuser

USER ${UID}:${GID}

WORKDIR /app

