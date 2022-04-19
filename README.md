# stocks

fetch all the stocks from https://www.direxion.com/etfs and alert based on heuristics

current supported heuristics:
 - 10 top movers of the day in Leverage ETFs
   - Actives, Losers, Gainers

## Setup
Most of the code is in the repo except morning star API credentials. 

A `config.yaml` is needed in the root of the repository with the following format

```yaml
ms_api:
  host: ms-finance.p.rapidapi.com
  url: https://ms-finance.p.rapidapi.com/market/v2/get-movers
  key: { MS_API_KEY_SECRET }
```

With that, its simple to run

```bash
go run orchestrator.go
```