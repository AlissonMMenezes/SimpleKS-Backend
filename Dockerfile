FROM debian:buster-slim
RUN mkdir /app
COPY simpleks-backend /app/server

# Run the web service on container startup.
CMD ["/app/server"]
