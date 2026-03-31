# HTTP Route Configuration Entry
# Routes HTTP traffic from the API gateway to gotak-api service

Kind = "http-route"
Name = "gotak-api-route"

# Bind to the API gateway's HTTP listener
Parents = [
  {
    Kind        = "api-gateway"
    Name        = "gotak-gateway"
    SectionName = "http-listener"
  }
]

# Simple catch-all route - forward all requests to gotak-api
Rules = [
  {
    Matches = [
      {
        Path = {
          Match = "prefix"
          Value = "/"
        }
      }
    ]
    Services = [
      {
        Name = "gotak-api"
      }
    ]
    # Rewrite to preserve the original path
    Filters = {
      URLRewrite = {
        Path = "/"
      }
    }
  }
]
