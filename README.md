# Crawler
Crawl and Analyze!

## About

This is simple Web-Crawler which i made just for fun.
The idea which making this crawler unique? is:
1. Get url
2. Write Html and Url
3. Feed html to llm to get info about website
4. Write url and description to database

## How to use

```bash
# launch kafka and postgres services
docker compose up -d
# Run llms
echo "Starting LLM service..."
cd service-llm-describer && go run main.go & 

# Run crawler
echo "Starting Crawler..."
cd ../service-crawler && go run main.go -w * -url *
```

## TODO
- [x] Framework for project
- [x] Add LLM processing
- [x] Add database for recording websites
- [x] Add RAG search
- [ ] Add validation for llm-generated json
- [x] Add frontend
