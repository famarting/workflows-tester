FROM openjdk:21-jdk-oracle
EXPOSE 5001
ENV PORT=5001
VOLUME /tmp
COPY target/*.jar app.jar

USER root
COPY ca.der /tmp/ca.der
RUN keytool -importcert -alias startssl -keystore $JAVA_HOME/jre/lib/security/cacerts -storepass changeit -file /tmp/ca.der

ENTRYPOINT ["java", "-Dhttps.proxyHost=host.docker.internal", "-Dhttps.proxyPort=8443", "-jar","./app.jar"]