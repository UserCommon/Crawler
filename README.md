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
go run main.go -w * -url *
```

## TODO
- [x] Framework for project
- [x] Add LLM processing
- [x] Add database for recording websites
- [x] Add RAG search
- [ ] Add validation for llm-generated json
- [x] Add frontend
