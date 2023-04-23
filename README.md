<p align="center">
  <img src="https://github.com/aymenfurter/florenceLLM/blob/main/webui/src/assets/logo.png?raw=true" alt="FlorenceLLM Logo" width="250">
</p>

<h1 align="center">FlorenceLLM</h1>

<p align="center">
  FlorenceLLM is an OpenAI GPT-3.5-based chatbot designed to help users find the right person to assist them within an organization. It achieves this by indexing git repositories and tracking who made which changes and when. This information allows the chatbot to identify the most suitable person to help with a specific topic.
</p>

<p align="center">
  <img src="https://github.com/aymenfurter/florenceLLM/blob/main/screenshot.png?raw=true" alt="Screenshot" width="600">
</p>

## Overview

At present, the indexing feature is under development, and its user interface has not been implemented. The chatbot is developed in Go, containerized, and deployed to Azure container apps.

During testing, Microsoft docs were indexed, and the chatbot's performance was excellent, as evidenced by the screenshot. To index your own repositories, refer to `indexer/indexer_test.go`. The deployment process is outlined in `workflows/deploy.yml`. Note that the indexing process involves embeddings requests, which may incur costs.

## Contributions

Pull requests and feedback is welcome! Feel free to test FlorenceLLM with your own git repositories. I hope you enjoy using it! 🤗

## Quick Links

- [Indexer Test](./indexer/indexer_test.go)
- [Deployment Configuration](./.github/workflows/deploy.yml)
