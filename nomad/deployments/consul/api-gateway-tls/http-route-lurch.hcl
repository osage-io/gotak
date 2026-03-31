# HTTP Route Configuration Entry for lurch.demoland.io
# Routes HTTPS traffic from the API gateway to lurch service

Kind = "http-route"
Name = "lurch-demoland-route"

# Bind to the API gateway's HTTPS listener
Parents = [
  {
    Kind        = "api-gateway"
    Name        = "demoland-gateway"
    SectionName = "https-listener"
  }
]

# Route based on hostname
Hostnames = ["lurch.demoland.io"]

# Route all requests to lurch service
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
        Name   = "lurch"
        Weight = 100
      }
    ]
  }
]
