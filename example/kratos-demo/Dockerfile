# kratos-demo - Multi-service Container
# This Dockerfile creates a container that runs both message and weather services
FROM debian:stable-slim

# Install essential packages for running Go binaries
RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
        netbase \
    && rm -rf /var/lib/apt/lists/ \
    && apt-get autoremove -y && apt-get autoclean -y

# Copy service binaries to the container
COPY ./app/message/bin /app
COPY ./app/weather/bin /app

# Copy service startup script
COPY ./service.sh /app

# Expose ports for both services
# Weather service HTTP port
EXPOSE 8080
# Message service HTTP port
EXPOSE 8081

# Start both services using the service script
ENTRYPOINT ["sh", "/app/service.sh"]