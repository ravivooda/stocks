# stocks

fetch all the stocks from different providers. Currently, supports:

- Direxion
- MicroSector ([limited](external/securities/microsector/holdings))
- [ProShares](external/securities/proshares/README.md)
- [Master Data Reports](external/securities/masterdatareports/README.md)
- [Invesco](external/securities/invesco/README.md)

If you are interested in all the tickers, please see [seeds.yaml](database/seeds.yaml)

## Alerts

current supported heuristics:

- 10 top movers of the day in Leverage ETFs with the highest exposure
    - Actives, Losers, Gainers

## Insights

These are implemented in [insights](insights) package

### Overlap between etfs

Finds n*n overlaps between all ETFs and output them to

1. csv for data dumping
2. Creates static websites where
    1. summaries for each ETF is generated
    2. overlaps between each ETF is generated

## Setup

Most of the code is in the repo except morning star API credentials.

A `secrets.yaml` is needed in the root of the repository with the following format

```yaml
ms_api:
  key: { MS_API_KEY_SECRET }
notifications:
  should_send_email: false # Set it to false when testing
uploads:
  should_upload_insights_output_to_gcp: true
```

With that, it's simple to run

```bash
go run orchestrator.go
```

--
# Credits

Using open sourced theme [quixlab](https://themefisher.com/products/quixlab).

[![DigitalOcean Referral Badge](https://web-platforms.sfo2.cdn.digitaloceanspaces.com/WWW/Badge%201.svg)](https://www.digitalocean.com/?refcode=d37e50f9fde4&utm_campaign=Referral_Invite&utm_medium=Referral_Program&utm_source=badge)

--

[![Deploy To DO](https://github.com/ravivooda/stocks/actions/workflows/deploy_to_do.yml/badge.svg)](https://github.com/ravivooda/stocks/actions/workflows/deploy_to_do.yml)
