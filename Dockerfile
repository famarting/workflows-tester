FROM maven:3.8.1-jdk-11-slim AS builder
WORKDIR /app
COPY pom.xml .
COPY src ./src/
RUN mvn install

FROM amazoncorretto:11-alpine
WORKDIR /app
COPY --from=builder /app/target/*.jar app.jar
EXPOSE 5001
ENV PORT=5001
VOLUME /tmp
ENTRYPOINT ["java","-jar","./app.jar"]