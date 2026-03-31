# HTTP Route Configuration Entry for gotak.demoland.io
# Routes HTTPS traffic from the API gateway to gotak-api service

Kind = "http-route"
Name = "gotak-demoland-route"

# Bind to the API gateway's HTTPS listener
Parents = [
  {
    Kind        = "api-gateway"
    Name        = "demoland-gateway"
    SectionName = "https-listener"
  }
]

# Route based on hostname
Hostnames = ["gotak.demoland.io"]

# Route all requests to gotak-api service
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
        Name   = "gotak-api"
        Weight = 100
      }
    ]
  }
]
