FROM eclipse-temurin:21-jre

RUN apt-get update -qq && \
    apt-get install -y -qq curl && \
    curl -fsSL https://deb.nodesource.com/setup_22.x | bash - && \
    apt-get install -y -qq nodejs && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

RUN npm install -g firebase-tools

WORKDIR /firebase
