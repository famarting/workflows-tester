FROM openjdk:21-jdk-oracle AS builder
WORKDIR /app
COPY pom.xml .
COPY src ./src/
RUN mvn install

FROM openjdk:21-jdk-oracle
WORKDIR /app
COPY --from=builder /app/target/*.jar app.jar
EXPOSE 5001
ENV PORT=5001
VOLUME /tmp

USER root
COPY ca.der $JAVA_HOME/jre/lib/security
RUN \
    cd $JAVA_HOME/jre/lib/security \
    && keytool -importcert -alias startssl -keystore $JAVA_HOME/jre/lib/security/cacerts -storepass changeit -file ca.der

ENTRYPOINT ["java","-jar","./app.jar"]