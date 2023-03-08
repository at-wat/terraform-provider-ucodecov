# terraform-provider-ucodecov

Codecov Provider (unofficial)

- [Document](docs/index.md)
- [License - Apache 2.0](./LICENSE)

# Breaking changes in v1

- Codecov API is updated to v2
- Requires v2 API token
  - Environment variable is changed to `CODECOV_API_V2_TOKEN`
  - Provider metadata is changed to `token_v2`
- `updatestamp` output is removed
