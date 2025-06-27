# crawl4ai OpenWebUI proxy
This simple proxy server can be run in a docker container to let an [OpenWebUI](https://github.com/open-webui/open-webui) instance interact with a [crawl4ai](https://github.com/unclecode/crawl4ai) instance.
This makes the OpenWebUI's web search feature a lot faster and way more usable without paying for an API service. 🎉

## Usage
Given a `compose.yml` file that looks something like this:

```
services:
    crawl4ai-proxy:
        build: . (or whereever this repo is)
        networks:
            - openwebui

    openwebui:
        image: ghcr.io/open-webui/open-webui:ollama
        ports:
            - "8080:8080"
        deploy:
            resources:
                reservations:
                    devices:
                        - driver: nvidia
                          count: all
                          capabilities: [gpu]
        networks:
            - openwebui

    crawl4ai:
        image: unclecode/crawl4ai:0.6.0-r2
        shm_size: 1g
        networks:
            - openwebui

networks:
    - openwebui
```

Run `docker compose up -d`, visit `localhost:8080` in a browser, navigate to `Admin Panel->Web Search` and under the "Loader" section, set

    Web Loader Engine: external
    External Web Loader URL: http://crawl4ai-proxy:8000/crawl
    External Web Loader API Key: * (doesn't matter, but is a required field)
