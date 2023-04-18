FROM golang:1.19

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main .

EXPOSE 8080

ENV PINECONE_API_KEY=your_api_key
ENV PINECONE_API_URL=your_api_url
ENV PINECONE_INDEX_NAME=your_index_name
ENV PINECONE_PROJECT_NAME=your_project_name
ENV PINECONE_ENV=your_environment
ENV OPEN_AI_KEY=your_open_ai_key

CMD ["./main"]
